#!/usr/bin/env python3
"""Pinned app-only release. Applying requires this feature's restart approval."""
import argparse
import copy
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.request

COMPOSE=Path('/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json')
TARGET='WeKnora-app'
BASELINE='sha256:94f2d38aa0663ee9441157715007f72f5514a40d2b56ae06318b8908b2e366e8'

def run(*args):
    return subprocess.check_output(args,text=True).strip()

def containers():
    ids=run('docker','ps','-aq').splitlines()
    return {c['Name'].lstrip('/'): (c['Id'],c['Image']) for c in json.loads(run('docker','inspect',*ids))}

def up():
    run('docker','compose','-f',str(COMPOSE),'up','-d','--no-deps','app')
    run('docker','exec','WeKnora-frontend-custom-preview','nginx','-t')
    run('docker','exec','WeKnora-frontend-custom-preview','nginx','-s','reload')

def main():
    p=argparse.ArgumentParser()
    p.add_argument('--image',required=True)
    p.add_argument('--apply',action='store_true')
    p.add_argument('--restart-authorized',action='store_true')
    a=p.parse_args()
    if not a.image.startswith('sha256:') or len(a.image)!=71: raise SystemExit('Immutable image required')
    assert run('docker','image','inspect','--format','{{.Id}}',a.image)==a.image
    before=containers()
    if before[TARGET][1]!=BASELINE: raise SystemExit('Baseline changed or already deployed; do not replay')
    original=COMPOSE.read_bytes(); config=json.loads(original); candidate=copy.deepcopy(config)
    candidate['services']['app']['image']=a.image
    check=copy.deepcopy(candidate); check['services']['app']['image']=config['services']['app']['image']
    assert check==config
    if not a.apply:
        print(json.dumps({'before':BASELINE,'after':a.image,'protected':len(before)-1})); return
    if not a.restart_authorized: raise SystemExit('Fresh restart authorization required')
    release=COMPOSE.parent/('agent-citations-'+a.image[7:19]); release.mkdir(mode=0o700,exist_ok=False)
    backup=release/'compose.before.json'; backup.write_bytes(original); os.chmod(backup,0o600)
    staged=release/'compose.after.json'; staged.write_text(json.dumps(candidate,indent=2)+'\n'); os.chmod(staged,0o600)
    run('docker','compose','-f',str(staged),'config','--quiet')
    assert containers()==before and COMPOSE.read_bytes()==original
    try:
        COMPOSE.write_bytes(staged.read_bytes()); up()
        for attempt in range(45):
            try:
                with urllib.request.urlopen('http://127.0.0.1:18081/health',timeout=3) as r: assert r.status==200
                with urllib.request.urlopen('http://127.0.0.1:18182/',timeout=3) as r: assert r.status==200
                break
            except Exception:
                if attempt==44: raise
                time.sleep(1)
        after=containers(); assert after[TARGET][1]==a.image
        for name,state in before.items():
            if name!=TARGET: assert after.get(name)==state, 'Protected container changed: '+name
        result={'status':'promoted','image':a.image,'protected_unchanged':len(before)-1,'rollback_compose':str(backup)}
        (release/'result.json').write_text(json.dumps(result,indent=2)+'\n'); print(json.dumps(result))
    except Exception:
        COMPOSE.write_bytes(original); up(); raise

if __name__=='__main__': main()
