#!/usr/bin/env python3
"""Read-only checks for the promoted production release; never restarts or migrates."""
import json
from pathlib import Path
import socket
import subprocess
import urllib.request

PRIVATE = Path('/home/ubuntu/weknora-fork-migration/20261021-production')

def run(*args):
    return subprocess.check_output(args, text=True).strip()

def inspect(name):
    return json.loads(run('docker', 'inspect', name))[0]

def check():
    state = json.loads((PRIVATE / 'state.json').read_text())['stage']
    assert state in ['cutover', 'complete'], 'Release not yet promoted: ' + state
    plan = json.loads((PRIVATE / 'compose.production.json').read_text())
    result = {'stage':state, 'canonical_frontend':'http://10.20.0.15:18182', 'artifacts':[]}
    for key in ['postgres', 'app', 'docreader', 'frontend']:
        service = plan['services'][key]
        d = inspect(service['container_name'])
        expected = json.loads(run('docker','image','inspect',service['image']))[0]['Id']
        assert d['Image'] == expected, key + ' image changed'
        assert d['State']['Running'], key + ' stopped'
        if d['State'].get('Health'):
            assert d['State']['Health']['Status']=='healthy', key + ' unhealthy'
        assert d['HostConfig']['RestartPolicy']['Name']=='unless-stopped', key + ' restart policy'
        result['artifacts'].append({'component':key,'container':service['container_name'],'image_id':d['Image']})
    app=inspect('WeKnora-app')
    env=dict(x.split('=',1) for x in app['Config']['Env'] if '=' in x)
    assert env['DB_HOST']=='WeKnora-postgres-20261021'
    assert env['REDIS_ADDR']=='WeKnora-redis:6379'
    assert env['FRONTEND_BASE_URL']=='http://10.20.0.15:18182'
    assert env['AUTO_MIGRATE']=='false' and env['ARCHIVAL_CITATIONS_ENABLED']=='true'
    assert not any(m.get('Name','').startswith('ljc-upstream-acceptance-') for m in app['Mounts'])
    ledger=run('docker','exec','WeKnora-postgres-20261021','sh','-ec','psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select version,dirty from schema_migrations; select version,dirty from ljc_schema_migrations;"')
    assert ledger=='128|f\n4|f',ledger
    result['production_ledger']=ledger
    for port,path in [(18081,'/health'),(18182,'/')]:
        with urllib.request.urlopen('http://127.0.0.1:'+str(port)+path,timeout=10) as r:
            assert r.status==200
    with urllib.request.urlopen('http://10.20.0.15:18083/health',timeout=10) as r:
        gpu=json.load(r)
    assert gpu['device'].startswith('cuda') and gpu['model']=='bge-reranker-v2-m3'
    assert run('systemctl','is-active','weknora-bge-rerank.service')=='active'
    assert run('systemctl','is-enabled','weknora-bge-rerank.service')=='enabled'
    result['gpu_rerank']=gpu
    if state=='complete':
        old=inspect('ljc-upstream-acceptance-custom-ui')
        assert not old['State']['Running'] and old['HostConfig']['RestartPolicy']['Name']=='no'
        for host in ['127.0.0.1','10.20.0.15']:
            with socket.socket() as s:
                s.settimeout(2)
                assert s.connect_ex((host,28182)) != 0,'Retired port28182 still listening'
        result['retired_28182']=True
    return result

if __name__=='__main__':
    print(json.dumps(check(),indent=2))
