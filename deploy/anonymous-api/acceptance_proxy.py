#!/usr/bin/env python3
"""Expose only the isolated anonymous endpoints for the specified acceptance IP."""
import argparse
import json
import pathlib
import subprocess
import ipaddress
p=argparse.ArgumentParser()
p.add_argument('--client-ip',required=True)
a=p.parse_args()
ip=ipaddress.ip_address(a.client_ip)
assert ip.version==4
state=pathlib.Path('/home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance')
name='weknora-api-test-proxy'
def run(args): return subprocess.check_output(args,text=True).strip()
old=subprocess.run(['docker','inspect',name],capture_output=True,text=True)
if old.returncode==0:
    assert json.loads(old.stdout)[0]['Config']['Labels'].get('ljc.acceptance')=='anonymous-api'
    subprocess.run(['docker','rm','-f',name],check=True,stdout=subprocess.DEVNULL)
frontend=json.loads(run(['docker','inspect','WeKnora-frontend-custom-preview']))[0]
conf=state/'acceptance-nginx.conf'
conf.write_text('''events {}\nhttp { access_log off; server { listen 8080;
location /api/openai/v1/ { proxy_pass http://weknora-api-acceptance:8080;
proxy_set_header X-Forwarded-For $remote_addr;
proxy_set_header Host $host;
proxy_buffering off; proxy_read_timeout 310s;
} location / { return 404; } } }\n''')
run(['docker','run','-d','--name',name,'--label','ljc.acceptance=anonymous-api',
     '--network','weknora-api-test-net','-p','10.20.0.15:28482:8080',
     '-v',str(conf)+':/etc/nginx/nginx.conf:ro','--entrypoint','nginx',
     frontend['Image'],'-g','daemon off;'])
proxy=json.loads(run(['docker','inspect',name]))[0]
proxy_ip=proxy['NetworkSettings']['Networks']['weknora-api-test-net']['IPAddress']
cfg=state/'config/anonymous-api.json'
c=json.loads(cfg.read_text())
c['trusted_proxy_cidrs']=[proxy_ip+'/32']
c['allowed_cidrs']=['127.0.0.1/32',json.loads((state/'runtime.json').read_text())['gateway']+'/32',str(ip)+'/32']
tmp=cfg.with_suffix('.next');tmp.write_text(json.dumps(c,ensure_ascii=False,indent=2));tmp.replace(cfg)
print(json.dumps({'base':'http://10.20.0.15:28482/api/openai/v1','allowed_client':str(ip),'trusted_proxy':proxy_ip,'production_changed':False}))
