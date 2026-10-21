#!/usr/bin/env python3
"""Explicit staged production release. No action is automatically replayed."""
import argparse
import datetime
import fcntl
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
PRIVATE = Path('/home/ubuntu/weknora-fork-migration/20261021-production')
NETWORK = 'weknora_v080_WeKnora-network'
PG = 'WeKnora-postgres-20261021'
PG_VOLUME = 'weknora-production-postgres-20261021'
BASELINE = 'a418adc45f383b9accb0d51bdcf226c468e656fb'
IMAGES = {
    'app': 'sha256:76399c0e703a7118ec3d7b5f7cd5140388f0c147bad32c2df6b154b1f92c51e1',
    'frontend': 'sha256:66b71bd73accc993f5c641dbeab4bc2bb8c999eae1f6050428c1fc147892043f',
    'docreader': 'sha256:fd18e28658738021e322fae832a11b8964341d1c19c52fcf6fb4cf35e10d9f2e',
    'postgres': 'docker.m.daocloud.io/paradedb/paradedb:v0.25.11-pg17',
}
CORE = ['users', 'tenants', 'knowledge_bases', 'knowledges', 'chunks', 'sessions',
        'messages', 'wiki_pages', 'archive_mapping_versions', 'message_citation_evidence',
        'enrichment_recoveries', 'tenant_api_keys']


def call(*args, **kwargs):
    return subprocess.check_output(args, text=True, **kwargs).strip()


def run(*args, **kwargs):
    subprocess.run(args, check=True, **kwargs)


def inspect(name):
    return json.loads(call('docker', 'inspect', name))[0]


def env(data):
    return dict(x.split('=', 1) for x in data['Config']['Env'] if '=' in x)


def write(name, data):
    path = PRIVATE / name
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
    path.chmod(0o600)


def sql(container, statement):
    return call('docker', 'exec', '-i', container, 'sh', '-ec',
                'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At -v ON_ERROR_STOP=1',
                input=statement)


def baseline():
    return {x['Name'].lstrip('/'): x for x in json.loads((PRIVATE / 'containers-private.json').read_text())}


def save_state(stage):
    write('state.json', {'stage': stage, 'at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
                         'head': call('git', '-C', str(ROOT), 'rev-parse', 'HEAD')})


def require_stage(stage):
    actual = json.loads((PRIVATE / 'state.json').read_text())['stage']
    if actual != stage:
        raise SystemExit(f'Refusing replay: expected {stage}, found {actual}; inspect before continuing')


def counts(container):
    result = {}
    for table in CORE:
        result[table] = int(sql(container, 'select count(*) from ' + table))
    return result


def compose(*args):
    run('docker', 'compose', '-f', str(PRIVATE / 'compose.production.json'), *args)


def snapshot_volume(volume, filename):
    assert volume and '/' not in volume
    run('docker', 'run', '--rm', '--user', '0', '--entrypoint', 'sh',
        '-v', volume + ':/source:ro', '-v', str(PRIVATE) + ':/backup', IMAGES['docreader'],
        '-ec', 'umask 077; tar -C /source -cf /backup/' + filename + ' .; chown ' + str(os.getuid()) + ':' + str(os.getgid()) + ' /backup/' + filename)
    run('tar', '-tf', str(PRIVATE / filename), stdout=subprocess.DEVNULL)


def volume_for(data, target):
    return next(m['Name'] for m in data['Mounts'] if m['Destination'] == target and m['Type'] == 'volume')


def preflight():
    if (PRIVATE / 'state.json').exists():
        require_stage('prepared')
    b = baseline()
    for name, old in b.items():
        now = inspect(name)
        assert now['Id'] == old['Id'] and now['Image'] == old['Image'], name + ' drift'
    assert call('git', '-C', '/home/ubuntu/RAG_Weknora', 'rev-parse', 'HEAD') == BASELINE
    assert sql('WeKnora-postgres', 'select version,dirty from schema_migrations') == '94|f'
    assert not sql('WeKnora-postgres', 'select task_type,count(*) from task_pending_ops group by 1')
    ae = env(b['WeKnora-app'])
    def redis(*a):
        return call('docker', 'exec', '-e', 'REDISCLI_AUTH=' + ae.get('REDIS_PASSWORD', ''),
                    'WeKnora-redis', 'redis-cli', '--raw', *a)
    pending = []
    for queue in redis('SMEMBERS', 'asynq:queues').splitlines():
        for suffix, command in [('active', 'LLEN'), ('pending', 'LLEN'), ('scheduled', 'ZCARD'), ('retry', 'ZCARD')]:
            number = int(redis(command, 'asynq:{' + queue + '}:' + suffix))
            if number:
                pending.append([queue, suffix, number])
    if pending:
        raise SystemExit('Live or scheduled work exists: ' + json.dumps(pending))
    for image in IMAGES.values():
        run('docker', 'image', 'inspect', image, stdout=subprocess.DEVNULL)
    run('docker', 'exec', 'WeKnora-app', 'curl', '--fail', '--silent', '--connect-timeout', '5',
        'http://192.168.16.1:18085/health')
    newenv = dict(ae)
    newenv.update(DB_HOST=PG, AUTO_MIGRATE='false', AUTO_RECOVER_DIRTY='false',
                  FRONTEND_BASE_URL='http://10.20.0.15:18182', APP_EXTERNAL_URL='http://10.20.0.15:18182')
    pe = env(b['WeKnora-postgres'])
    dsn = 'postgres://' + urllib.parse.quote(pe['POSTGRES_USER'], safe='') + ':' + urllib.parse.quote(pe['POSTGRES_PASSWORD'], safe='') + '@' + PG + ':5432/' + urllib.parse.quote(pe['POSTGRES_DB'], safe='') + '?sslmode=disable'
    (PRIVATE / 'migration.env').write_text('LJC_MIGRATION_DSN=' + dsn + '\n')
    (PRIVATE / 'migration.env').chmod(0o600)
    mounts = []
    for m in b['WeKnora-app']['Mounts']:
        if m['Destination'] == '/app/config/config.yaml':
            mounts.append(str(ROOT / 'config') + ':/app/config:ro')
        else:
            mounts.append((m.get('Name') if m['Type'] == 'volume' else m['Source']) + ':' + m['Destination'] + (':ro' if not m['RW'] else ''))
    doc = b['WeKnora-docreader']
    dm = [(m.get('Name') if m['Type'] == 'volume' else m['Source']) + ':' + m['Destination'] + (':ro' if not m['RW'] else '') for m in doc['Mounts']]
    volumes = {PG_VOLUME: {'external': True}}
    for d in [b['WeKnora-app'], doc]:
        for m in d['Mounts']:
            if m['Type'] == 'volume':
                volumes[m['Name']] = {'external': True}
    labels = {'ljc.operation': 'production-20261021', 'ljc.canonical_frontend': '18182'}
    config = {
        'name': 'ljc-weknora-production',
        'services': {
            'postgres': {'image': IMAGES['postgres'], 'container_name': PG, 'restart': 'unless-stopped',
                'labels': labels, 'environment': pe, 'networks': ['production'],
                'command': ['postgres', '-c', 'maintenance_work_mem=2GB', '-c', 'max_parallel_maintenance_workers=0'],
                'volumes': [PG_VOLUME + ':/var/lib/postgresql/data', str(PRIVATE) + ':/backup:ro'],
                'healthcheck': {'test': ['CMD-SHELL', 'pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'], 'interval': '5s', 'timeout': '3s', 'retries': 30}},
            'docreader': {'image': IMAGES['docreader'], 'container_name': 'WeKnora-docreader',
                'restart': 'unless-stopped', 'init': True, 'labels': labels, 'environment': env(doc),
                'networks': ['production'], 'volumes': dm,
                'healthcheck': {'test': ['CMD', '/bin/grpc_health_probe', '-addr=:50051'], 'interval': '10s', 'timeout': '5s', 'retries': 30}},
            'app': {'image': IMAGES['app'], 'container_name': 'WeKnora-app', 'restart': 'unless-stopped',
                'init': True, 'labels': labels, 'environment': newenv, 'ports': ['18081:8080'],
                'volumes': mounts, 'networks': ['production'], 'extra_hosts': ['host.docker.internal:host-gateway'],
                'healthcheck': {'test': ['CMD-SHELL', 'curl -fsS http://127.0.0.1:8080/health >/dev/null'], 'interval': '10s', 'timeout': '5s', 'retries': 30}},
            'frontend': {'image': IMAGES['frontend'], 'container_name': 'WeKnora-frontend-custom-preview',
                'restart': 'unless-stopped', 'labels': labels, 'ports': ['18182:80'],
                'environment': {'APP_HOST': 'WeKnora-app', 'APP_PORT': '8080', 'DEFAULT_LOCALE': 'zh-CN',
                    'MAX_FILE_SIZE_MB': env(b['WeKnora-frontend-custom-preview']).get('MAX_FILE_SIZE_MB', '50')},
                'networks': ['production']},
        }, 'networks': {'production': {'external': True, 'name': NETWORK}}, 'volumes': volumes}
    # Compose performs interpolation even in JSON environment values.
    for service in config['services'].values():
        service['environment'] = {k: v.replace('$', '$$') for k, v in service.get('environment', {}).items()}
    write('compose.production.json', config)
    compose('config', '--quiet')
    write('preflight-results.json', {'queues_active_pending_scheduled_retry': 0, 'durable_pending_ops': 0,
          'production_ledger': '94|f', 'images': IMAGES, 'canonical_frontend': 18182,
          'historical_statuses_preserved': True, 'candidate_data_not_used': True})
    save_state('prepared')
    print('Prepared production composition; credentials remain private. No services changed.')


def backup():
    require_stage('prepared')
    preflight()  # revalidate immediately before the maintenance boundary
    save_state('backing-up')
    b = baseline()
    run('docker', 'stop', '-t', '120', 'WeKnora-frontend-custom-preview', 'WeKnora-frontend', 'WeKnora-app', 'WeKnora-docreader')
    assert not inspect('WeKnora-app')['State']['Running']
    write('protected-counts-before.json', counts('WeKnora-postgres'))
    with (PRIVATE / 'database.dump').open('wb') as f:
        run('docker', 'exec', 'WeKnora-postgres', 'sh', '-ec', 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc', stdout=f)
    run('docker', 'exec', 'WeKnora-postgres', 'sh', '-ec', 'pg_dumpall -U "$POSTGRES_USER" --globals-only', stdout=(PRIVATE / 'globals.sql').open('wb'))
    run('docker', 'exec', '-i', 'WeKnora-postgres', 'pg_restore', '--list', stdin=(PRIVATE / 'database.dump').open('rb'), stdout=subprocess.DEVNULL)
    run('docker', 'stop', '-t', '60', 'WeKnora-redis', 'WeKnora-neo4j')
    snapshot_volume(volume_for(b['WeKnora-app'], '/data/files'), 'files.tar')
    snapshot_volume(volume_for(b['WeKnora-redis'], '/data'), 'redis.tar')
    snapshot_volume(volume_for(b['WeKnora-neo4j'], '/data'), 'neo4j.tar')
    run('tar', '-C', '/home/ubuntu/RAG_Weknora', '-czf', str(PRIVATE / 'source-config.tar.gz'), 'config', 'skills/preloaded', '.env')
    run('docker', 'start', 'WeKnora-redis', 'WeKnora-neo4j')
    hashes = {}
    for name in ['database.dump', 'globals.sql', 'files.tar', 'redis.tar', 'neo4j.tar', 'source-config.tar.gz']:
        with (PRIVATE / name).open('rb') as f:
            digest = hashlib.sha256()
            for block in iter(lambda: f.read(4 * 1024 * 1024), b''):
                digest.update(block)
            hashes[name] = digest.hexdigest()
    write('backup-sha256.json', hashes)
    save_state('backed-up')
    print('Frozen production backup complete; old application remains stopped.')


def restore():
    require_stage('backed-up')
    try:
        inspect(PG)
    except subprocess.CalledProcessError:
        pass
    else:
        raise SystemExit('Target database container already exists; inspect, never drop or overwrite')
    try:
        call('docker', 'volume', 'inspect', PG_VOLUME)
    except subprocess.CalledProcessError:
        run('docker', 'volume', 'create', '--label', 'ljc.operation=production-20261021', PG_VOLUME)
    else:
        raise SystemExit('Target volume already exists; inspect before continuing')
    save_state('restoring')
    compose('up', '-d', 'postgres')
    wait_health(PG)
    sql(PG, 'CREATE EXTENSION vector WITH SCHEMA public;')
    run('docker', 'exec', PG, 'sh', '-ec', 'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --no-owner --no-privileges --exit-on-error /backup/database.dump')
    assert sql(PG, 'select version,dirty from schema_migrations') == '94|f'
    assert counts(PG) == json.loads((PRIVATE / 'protected-counts-before.json').read_text())
    save_state('restored')


def wait_health(name):
    for _ in range(90):
        d = inspect(name)
        h = d['State'].get('Health', {}).get('Status', '')
        if h == 'healthy':
            return
        if not d['State']['Running'] or h == 'unhealthy':
            raise SystemExit(name + ' not healthy; inspect private logs')
        time.sleep(2)
    raise SystemExit(name + ' health timed out; inspect existing container, do not recreate blindly')


def migrate():
    require_stage('restored')
    save_state('migrating')
    for mode in ['status', 'bootstrap-legacy', 'up', 'up', 'status']:
        with (PRIVATE / ('migration-' + mode + '.log')).open('a') as log:
            run('docker', 'run', '--rm', '--network', NETWORK, '--env-file', str(PRIVATE / 'migration.env'),
                '--entrypoint', '/app/ljc-migrate', IMAGES['app'], '-mode', mode, stdout=log, stderr=subprocess.STDOUT)
    ledger = sql(PG, 'select version,dirty from schema_migrations; select version,dirty from ljc_schema_migrations;')
    assert ledger == '128|f\n4|f', ledger
    after = counts(PG)
    assert after == json.loads((PRIVATE / 'protected-counts-before.json').read_text()), after
    write('migration-results.json', {'ledger': ledger, 'protected_counts': after, 'repeat_up_clean': True, 'new_volume': PG_VOLUME})
    save_state('migrated')


def cutover():
    require_stage('migrated')
    save_state('cutting-over')
    for name in ['WeKnora-app', 'WeKnora-docreader', 'WeKnora-frontend-custom-preview']:
        assert not inspect(name)['State']['Running']
        run('docker', 'update', '--restart=no', name)
        run('docker', 'rename', name, name + '-baseline-20261021')
    run('docker', 'stop', '-t', '60', 'WeKnora-postgres')
    run('docker', 'update', '--restart=no', 'WeKnora-postgres')
    compose('up', '-d', 'docreader')
    wait_health('WeKnora-docreader')
    compose('up', '-d', 'app')
    wait_health('WeKnora-app')
    compose('up', '-d', 'frontend')
    run('docker', 'exec', 'WeKnora-frontend-custom-preview', 'nginx', '-t')
    # Preserve the legacy comparison entry but make its cached backend resolution current.
    run('docker', 'start', 'WeKnora-frontend')
    run('docker', 'exec', 'WeKnora-frontend', 'nginx', '-s', 'reload')
    save_state('cutover')


def verify():
    require_stage('cutover')
    for name in ['WeKnora-app', 'WeKnora-docreader', PG]:
        wait_health(name)
    for port in [18081, 18182]:
        with urllib.request.urlopen('http://127.0.0.1:' + str(port) + ('/health' if port == 18081 else '/'), timeout=10) as r:
            assert r.status == 200
    assert inspect('WeKnora-frontend-custom-preview')['Image'] == IMAGES['frontend']
    assert inspect('WeKnora-app')['Image'] == IMAGES['app']
    assert env(inspect('WeKnora-app'))['DB_HOST'] == PG
    assert env(inspect('WeKnora-app'))['JWT_SECRET'] == env(baseline()['WeKnora-app'])['JWT_SECRET']
    assert sql(PG, 'select version,dirty from schema_migrations; select version,dirty from ljc_schema_migrations;') == '128|f\n4|f'
    assert counts(PG) == json.loads((PRIVATE / 'protected-counts-before.json').read_text())
    write('cutover-results.json', {'frontend': 18182, 'backend': 18081, 'images': IMAGES,
          'database': PG, 'protected_counts_preserved': True, 'jwt_secret_preserved': True,
          'browser_acceptance_required_before_retiring_28182': True})
    print('Runtime and data checks passed; perform authenticated browser acceptance before retire.')


def retire():
    require_stage('cutover')
    proof = json.loads((PRIVATE / 'browser-acceptance-results.json').read_text())
    assert proof['passed'] and proof['port'] == 18182
    run('docker', 'stop', '-t', '30', 'ljc-upstream-acceptance-custom-ui')
    run('docker', 'update', '--restart=no', 'ljc-upstream-acceptance-custom-ui')
    assert not inspect('ljc-upstream-acceptance-custom-ui')['State']['Running']
    for d in json.loads(call('docker', 'inspect', *call('docker', 'ps', '-q').splitlines())):
        assert all(p['HostPort'] != '28182' for ports in (d['NetworkSettings']['Ports'] or {}).values() if ports for p in ports)
    save_state('complete')
    print('18182 canonical; former28182 container stopped with restart disabled.')


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('stage', choices=['preflight', 'backup', 'restore', 'migrate', 'cutover', 'verify', 'retire', 'status'])
    p.add_argument('--confirm-production-release', action='store_true')
    a = p.parse_args()
    if a.stage == 'status':
        print((PRIVATE / 'state.json').read_text() if (PRIVATE / 'state.json').exists() else 'not started')
        return
    if not a.confirm_production_release:
        raise SystemExit('Explicit --confirm-production-release required; see production README')
    PRIVATE.mkdir(mode=0o700, exist_ok=True)
    lock = (PRIVATE / 'release.lock').open('w')
    fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
    os.umask(0o077)
    os.chdir(ROOT)
    globals()[a.stage]()


if __name__ == '__main__':
    main()
