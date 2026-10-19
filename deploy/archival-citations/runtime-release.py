#!/usr/bin/env python3
"""Explicit release actions; check is read-only with respect to live services.

Usage: python3 runtime-release.py RELEASE_DIR check|backup|app|frontend|rollback
The operator must obtain the current user's restart window before mutations.
"""
import json,os,pathlib,subprocess,sys,time,urllib.request

ROOT=pathlib.Path('/home/ubuntu/RAG_Weknora')
folder=pathlib.Path(sys.argv[1]).resolve()
if not folder.is_relative_to(ROOT/'runtime/archival-citations'):
    raise SystemExit('release directory outside the recorded workspace')
action=sys.argv[2]
release=json.loads((folder/'release.json').read_text())
before=json.loads((folder/'app-before.json').read_text())[0]
ui_before=json.loads((folder/'frontend-before.json').read_text())[0]
os.umask(0o077)

def run(*args,**kw):return subprocess.check_output(args,text=True,**kw).strip()
def inspect(name):return json.loads(run('docker','inspect',name))[0]
def save(name,value):
    path=folder/name;path.write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n');path.chmod(0o600)
def env_dict(row):return dict(v.split('=',1) for v in row['Config']['Env'])
def compose(image,enabled):
    env=env_dict(before);env['ARCHIVAL_CITATIONS_ENABLED']='true' if enabled else 'false'
    overlay=folder/'compose.runtime.json'
    save(overlay.name,{'services':{'app':{'image':image,'environment':env}}})
    files=before['Config']['Labels']['com.docker.compose.project.config_files'].split(',')
    args=['docker','compose','-p','weknora_v080']
    for f in files+[str(overlay)]:args.extend(['-f',f])
    return args
def validate():
    if inspect('WeKnora-app')['Image']!=release['previous_app_id']:
        raise RuntimeError('live app changed since baseline; inspect before replacing')
    if inspect('WeKnora-frontend-custom-preview')['Image']!=release['previous_frontend_id']:
        raise RuntimeError('live custom frontend changed since baseline')
    args=compose(release['app_image'],True)
    cfg=json.loads(run(*args,'config','--format','json',cwd=ROOT))['services']['app']
    wanted=env_dict(before);wanted['ARCHIVAL_CITATIONS_ENABLED']='true'
    if cfg['environment']!=wanted:
        different=sorted(k for k in set(cfg['environment'])|set(wanted) if cfg['environment'].get(k)!=wanted.get(k))
        raise RuntimeError('unexpected runtime environment changes: '+','.join(different))
    old={m['Destination']:(m.get('Name') if m['Type']=='volume' else m['Source'],not m['RW']) for m in before['Mounts']}
    new={v['target']:(v['source'],v.get('read_only',False)) for v in cfg.get('volumes',[])}
    # Named Compose volumes use logical names; resolve their external runtime names.
    all_cfg=json.loads(run(*args,'config','--format','json',cwd=ROOT))
    for target,(source,ro) in list(new.items()):
        source=all_cfg.get('volumes',{}).get(source,{}).get('name',source);new[target]=(source,ro)
    if old!=new:raise RuntimeError('runtime mount drift; do not deploy')
    save('validated-runtime.json',{'env_names':sorted(wanted),'mounts':new,'image':cfg['image'],'original_frontend_untouched':True})
    print('PASS: only app image and archival flag change; mounts and existing environment preserved')
def healthy():
    for _ in range(60):
        try:
            with urllib.request.urlopen('http://127.0.0.1:18081/health',timeout=2) as r:
                if r.status==200:
                    # The upstream health route tolerates migration failures.
                    # Require an actually readable migration source as appuser.
                    subprocess.run(['docker','exec','--user','appuser','WeKnora-app','sh','-c','test -r /app/migrations/versioned/000094_archival_citations.up.sql && test -x /app/migrations/versioned'],check=True)
                    return
        except Exception:pass
        time.sleep(2)
    raise RuntimeError('app did not become healthy')
def replace_app(image,enabled):
    subprocess.run(compose(image,enabled)+['up','-d','--no-deps','app'],cwd=ROOT,check=True)
    healthy()
def restore_frontend():
    if run('docker','ps','-a','--filter','name=^/WeKnora-frontend-custom-before-citations$','--format','{{.Names}}'):
        if run('docker','ps','-a','--filter','name=^/WeKnora-frontend-custom-preview$','--format','{{.Names}}'):
            subprocess.run(['docker','stop','WeKnora-frontend-custom-preview'],check=True)
            subprocess.run(['docker','rename','WeKnora-frontend-custom-preview','WeKnora-frontend-custom-citations-stopped'],check=True)
        subprocess.run(['docker','rename','WeKnora-frontend-custom-before-citations','WeKnora-frontend-custom-preview'],check=True)
        subprocess.run(['docker','start','WeKnora-frontend-custom-preview'],check=True)

if action=='check':validate()
elif action=='backup':
    validate()
    with (folder/'database-before.dump').open('xb') as f:
        subprocess.run(['docker','exec','WeKnora-postgres','sh','-c','exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc'],stdout=f,check=True)
    print('database backup saved (not modified)')
elif action=='app':
    validate()
    if not (folder/'database-before.dump').is_file():raise RuntimeError('database backup required')
    try:replace_app(release['app_image'],True)
    except Exception:
        replace_app(release['rollback_image'],False);raise
    migration=run('docker','exec','WeKnora-postgres','sh','-c','psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select version,dirty from schema_migrations"')
    if migration!='94|f':
        replace_app(release['rollback_image'],False)
        raise RuntimeError('migration not at clean version 94; reverted old binary')
    save('app-deployed.json',{'image_id':inspect('WeKnora-app')['Image'],'migration':migration,'time':time.time()})
    print('app healthy; citation feature enabled')
elif action=='frontend':
    if inspect('WeKnora-app')['Image']!=release['app_image_id']:raise RuntimeError('new backend not running')
    if inspect('WeKnora-frontend-custom-preview')['Image']!=release['previous_frontend_id']:raise RuntimeError('custom baseline changed')
    if run('docker','ps','-a','--filter','name=^/WeKnora-frontend-custom-before-citations$','--format','{{.Names}}'):raise RuntimeError('rollback container name already used')
    subprocess.run(['docker','stop','WeKnora-frontend-custom-preview'],check=True)
    subprocess.run(['docker','rename','WeKnora-frontend-custom-preview','WeKnora-frontend-custom-before-citations'],check=True)
    args=['docker','run','-d','--name','WeKnora-frontend-custom-preview','--restart','unless-stopped','--network',ui_before['HostConfig']['NetworkMode'],'-p','18182:80']
    for name,value in env_dict(ui_before).items():args.extend(['-e',name+'='+value])
    try:
        subprocess.run(args+[release['frontend_image']],check=True)
        time.sleep(2)
        with urllib.request.urlopen('http://127.0.0.1:18182/login',timeout=10) as r:
            if r.status!=200:raise RuntimeError('login page failed')
    except Exception:
        restore_frontend();replace_app(release['rollback_image'],False);raise
    save('frontend-deployed.json',{'image_id':inspect('WeKnora-frontend-custom-preview')['Image'],'time':time.time()})
    print('custom login page healthy; old container retained')
elif action=='rollback':
    restore_frontend();replace_app(release['rollback_image'],False)
    print('old binary restored; new evidence tables retained')
else:raise SystemExit('unknown release action')
