#!/usr/bin/env python3
"""Fault/cancellation checks: independently named acceptance instance ONLY."""
import json
import pathlib
import subprocess
import time
import requests
STATE=pathlib.Path('/home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance')
BASE='http://127.0.0.1:28481/api/openai/v1'
cfg=STATE/'config/anonymous-api.json'
original=json.loads(cfg.read_text())
result={}
def atomic(v):
    tmp=cfg.with_suffix('.next');tmp.write_text(json.dumps(v));tmp.replace(cfg)
def check(key,v):result[key]=bool(v);assert v,key
def own(name):
    v=json.loads(subprocess.check_output(['docker','inspect',name]))[0]
    assert v['Config']['Labels'].get('ljc.acceptance')=='anonymous-api'
def healthy():
    for i in range(60):
        try:
            if requests.get(BASE+'/models',timeout=2).status_code==200:return
        except requests.RequestException:pass
        time.sleep(1)
    raise RuntimeError('acceptance not ready')
body={'model':'fixture-qa','messages':[{'role':'user','content':'根据独立API验收资料，检查周期和备份保留份数分别是多少？'}]}
try:
    healthy()
    own('weknora-api-test-redis')
    # Real untrusted peer on the same isolated Docker network, with no allowed CIDR.
    denied=subprocess.run(['docker','exec','weknora-api-test-redis','wget','-S','-O-',
                           'http://weknora-api-acceptance:8080/api/openai/v1/models'],capture_output=True,text=True)
    check('real_disallowed_peer_403','403' in denied.stderr)
    check('direct_header_cannot_spoof_peer',requests.get(BASE+'/models',headers={'X-Forwarded-For':'198.51.100.1'},timeout=5).status_code==200)
    short=dict(original);short['timeout_seconds']=1;atomic(short)
    start=time.monotonic();r=requests.post(BASE+'/chat/completions',json=body,timeout=10)
    check('real_timeout_504',r.status_code==504 and time.monotonic()-start<6)
    atomic(original)
    before=subprocess.check_output(['docker','logs','weknora-api-acceptance'],stderr=subprocess.STDOUT,text=True).count('error_type=canceled_or_timeout')
    streamed=dict(body,stream=True)
    streamed['messages']=[{'role':'user','content':'请根据独立API验收资料，逐条详细解释所有检查和备份要求，每条用一段话说明。'}]
    with requests.post(BASE+'/chat/completions',json=streamed,stream=True,timeout=20) as r:
        check('disconnect_started_stream',r.status_code==200)
        for line in r.iter_lines(chunk_size=1):
            if line.startswith(b'data:'):break
    time.sleep(2)
    active=subprocess.check_output(['docker','exec','weknora-api-test-redis','redis-cli','ZCARD','weknora:{anonymous-api}:active'],text=True).strip()
    check('disconnect_releases_shared_lease',active=='0')
    after=subprocess.check_output(['docker','logs','weknora-api-acceptance'],stderr=subprocess.STDOUT,text=True).count('error_type=canceled_or_timeout')
    check('disconnect_reaches_execution_context',after>before)
    known=json.loads((STATE/'sdk-server.json').read_text())['sync']['id']
    changed=json.loads(json.dumps(original));changed['profiles'][0]['id']='temporarily-withdrawn';atomic(changed)
    check('real_scope_revocation_404',requests.get(BASE+'/responses/'+known+'/citations',timeout=5).status_code==404)
    atomic(original)
    own('weknora-api-acceptance')
    subprocess.run(['docker','restart','weknora-api-acceptance'],check=True,stdout=subprocess.DEVNULL)
    healthy()
    check('postgres_mapping_survives_restart',requests.get(BASE+'/responses/'+known+'/citations',timeout=5).status_code==200)
finally:
    atomic(original)
path=STATE/'edge-checks.json';path.write_text(json.dumps(result,indent=2))
print(json.dumps({'passed':len(result),'report':str(path),'production_changed':False}))
