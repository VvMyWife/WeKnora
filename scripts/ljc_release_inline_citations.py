"""Guarded image-only promotion. No migrations, secret output or queue writes."""
import argparse
import copy
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
CONFIG = Path('/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json')
REVISION = '51935395ed18967f60b2017536ef469d02734271'
BASES = {'WeKnora-app': 'sha256:ccba5915625333f46d2cd3c4e57525049c4605512dcfbbbad43c4528f4becc23', 'WeKnora-frontend-custom-preview': 'sha256:b57758c9199d56d54197f2442db639bd175d45838980fa2be48c50d66d17dae5'}
PROTECTED = ['WeKnora-postgres-20261021', 'WeKnora-docreader', 'WeKnora-redis', 'WeKnora-neo4j', 'WeKnora-frontend', 'scu-qwen35-9b']

def run(*args):
    result = subprocess.run(args, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if result.returncode:
        raise RuntimeError(f'command failed: {args[0]} {args[1]} (exit {result.returncode}); inspect privately')
    return result.stdout

def inspect(name):
    return json.loads(run('docker', 'inspect', name))[0]

def same_except_images(before, after):
    before, after = copy.deepcopy(before), copy.deepcopy(after)
    for service in ['app', 'frontend']:
        before['services'][service].pop('image', None)
        after['services'][service].pop('image', None)
    return before == after

def health():
    for attempt in range(25):
        try:
            for url in ['http://127.0.0.1:18081/health', 'http://127.0.0.1:18182/']:
                with urllib.request.urlopen(url, timeout=2) as response:
                    if response.status != 200:
                        raise RuntimeError('unhealthy')
            return
        except Exception:
            if attempt == 24:
                raise RuntimeError('health did not recover; rolling back')
            time.sleep(1)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--apply', action='store_true')
    parser.add_argument('--restart-authorized', action='store_true')
    args = parser.parse_args()
    if args.apply and not args.restart_authorized:
        parser.error('restart authorization is required')
    for name, image in BASES.items():
        if inspect(name)['Image'] != image:
            raise RuntimeError(f'{name} changed; inspect state before continuing')
    protected = {name: inspect(name)['Id'] for name in PROTECTED}
    overlay = ROOT / 'deploy/inline-citations/compose.release.json'
    targets = json.loads(overlay.read_text())['services']
    for service in targets.values():
        image = json.loads(run('docker', 'image', 'inspect', service['image']))[0]
        if image['Config']['Labels'].get('org.opencontainers.image.revision') != REVISION:
            raise RuntimeError('release image source mismatch')
    release = CONFIG.parent / f'inline-citations-{REVISION[:9]}'
    release.mkdir(mode=0o700, exist_ok=True)
    backup, candidate = release/'compose.before.json', release/'compose.next.json'
    if backup.exists():
        raise RuntimeError('release already attempted; inspect before replaying')
    original = json.loads(run('docker', 'compose', '-f', str(CONFIG), 'config', '--format', 'json'))
    run('docker', 'compose', '-f', str(CONFIG), '-f', str(overlay), 'config', '--format', 'json', '--output', str(candidate))
    os.chmod(candidate, 0o600)
    if not same_except_images(original, json.loads(candidate.read_text())):
        raise RuntimeError('non-image runtime configuration changed')
    print('Verified: only app/frontend images change; no migration or protected-service restart.', flush=True)
    if not args.apply:
        return
    shutil.copyfile(CONFIG, backup)
    os.chmod(backup, 0o600)
    try:
        shutil.copyfile(candidate, CONFIG.with_suffix('.next'))
        os.chmod(CONFIG.with_suffix('.next'), 0o600)
        os.replace(CONFIG.with_suffix('.next'), CONFIG)
        run('docker', 'compose', '-f', str(CONFIG), 'config', '--quiet')
        run('docker', 'compose', '-f', str(CONFIG), 'up', '-d', '--no-deps', 'app', 'frontend')
        run('docker', 'exec', 'WeKnora-frontend-custom-preview', 'nginx', '-t')
        run('docker', 'exec', 'WeKnora-frontend-custom-preview', 'nginx', '-s', 'reload')
        health()
        for name, identity in protected.items():
            if inspect(name)['Id'] != identity:
                raise RuntimeError(f'protected container changed: {name}')
        result = {'source_revision': REVISION, 'protected_container_ids_unchanged': True, 'images': {n: inspect(n)['Image'] for n in BASES}, 'health': 'passed', 'browser_acceptance': 'pending'}
        (release/'result.json').write_text(json.dumps(result, indent=2)+'\n')
        print(json.dumps(result), flush=True)
    except Exception:
        shutil.copyfile(backup, CONFIG)
        run('docker', 'compose', '-f', str(CONFIG), 'up', '-d', '--no-deps', 'app', 'frontend')
        health()
        print('Rolled back app/frontend images; database and evidence retained.', flush=True)
        raise

if __name__ == '__main__':
    main()
