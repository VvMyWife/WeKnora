#!/usr/bin/env python3
"""Explicit prepare/apply/rollback. Production changes require a maintenance flag."""
import argparse
import datetime
import ipaddress
import json
import os
import pathlib
import subprocess
import time
import uuid
import requests

REPO=pathlib.Path(__file__).resolve().parents[2]
ROOT=pathlib.Path('/home/ubuntu/RAG_Weknora')
CANON=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json')
def run(args, **kw):
    p=subprocess.run(args,capture_output=True,text=True,**kw)
    if p.returncode: raise RuntimeError('command failed: '+args[0]+'; inspect private release logs')
    return p.stdout.strip()
def inspect(name):return json.loads(run(['docker','inspect',name]))[0]
def write(path,obj):
    path.write_text(json.dumps(obj,ensure_ascii=False,indent=2));path.chmod(0o600)
def env_of(app):return dict(x.split('=',1) for x in app['Config']['Env'] if '=' in x)
def sql(env,query):
    return run(['docker','exec',env['DB_HOST'],'psql','-XAt','-v','ON_ERROR_STOP=1','-U',env['DB_USER'],'-d',env['DB_NAME'],'-c',query])
def atomic(path,obj):
    tmp=path.with_suffix('.next');write(tmp,obj);os.replace(tmp,path)
def compose(directory):
    return ['docker','compose','-p','ljc-weknora-production','-f',str(directory/'compose.api.json')]
def main():
    p=argparse.ArgumentParser()
    p.add_argument('action',choices=['prepare','apply','rollback'])
    p.add_argument('--release-dir')
    p.add_argument('--kb',default='8575ca20-5e7e-48c4-9fc0-786d979fdf5d')
    p.add_argument('--agent',default='builtin-quick-answer')
    p.add_argument('--alias',default='policy-qa')
    p.add_argument('--cidr',action='append',default=[])
    p.add_argument('--maintenance-approved',action='store_true')
    a=p.parse_args()
    app=inspect('WeKnora-app');env=env_of(app)
    if a.action=='prepare':
        for value in a.cidr:ipaddress.ip_network(value,strict=True)
        if not a.cidr:raise RuntimeError('specify verified source CIDRs explicitly')
        uuid.UUID(a.kb)
        tenant=sql(env,"SELECT tenant_id FROM knowledge_bases WHERE id='"+a.kb+"' AND deleted_at IS NULL")
        if not tenant.isdigit():raise RuntimeError('knowledge scope not found')
        d=ROOT/'runtime/anonymous-api-releases'/datetime.datetime.now().strftime('%Y%m%d-%H%M%S')
        d.mkdir(parents=True,mode=0o700)
        baseline=json.loads(CANON.read_text());service=baseline['services']['app']
        if service['image']!=app['Image']:raise RuntimeError('canonical Compose and runtime image differ')
        write(d/'compose.before.json',baseline)
        frontend=inspect('WeKnora-frontend-custom-preview')
        networks=app['NetworkSettings']['Networks']
        shared=[n for n in networks if n in frontend['NetworkSettings']['Networks']]
        if len(shared)!=1:raise RuntimeError('review frontend proxy network')
        proxy=frontend['NetworkSettings']['Networks'][shared[0]]['IPAddress']
        cfg={'enabled':True,'tenant_id':int(tenant),'user_id':str(uuid.uuid4()),
             'allowed_cidrs':a.cidr,'trusted_proxy_cidrs':[proxy+'/32'],
             'global_concurrency':4,'ip_concurrency':2,'requests_per_minute':10,'timeout_seconds':300,
             'profiles':[{'id':a.alias,'agent_id':a.agent,'knowledge_base_ids':[a.kb],'max_tokens':2048}]}
        write(d/'config.proposed.json',cfg)
        published=ROOT/'config/anonymous-api'
        published.mkdir(parents=True,exist_ok=True)
        # No production environment references this directory until apply.
        service['image']='weknora-app:anonymous-api-20261104'
        service['environment']['WEKNORA_ANONYMOUS_API_CONFIG']='/app/api-config/anonymous-api.json'
        service['volumes'].append({'type':'bind','source':str(published),'target':'/app/api-config','read_only':True})
        write(d/'compose.api.json',baseline)
        write(d/'state.json',{'phase':'prepared','old_image':app['Image'],
                             'new_image':inspect(service['image'])['Id'],'network':shared[0],
                             'frontend_proxy':proxy,'published_config':str(published/'anonymous-api.json'),
                             'source_commit':run(['git','-C',str(REPO),'rev-parse','HEAD'])})
        print(json.dumps({'prepared':str(d),'production_changed':False,'source_cidrs':a.cidr,'profile':a.alias}))
        return
    if not a.maintenance_approved:raise RuntimeError('apply/rollback require an approved maintenance window: --maintenance-approved')
    if not a.release_dir:raise RuntimeError('--release-dir is required')
    d=pathlib.Path(a.release_dir).resolve()
    assert d.is_relative_to((ROOT/'runtime/anonymous-api-releases').resolve())
    s=json.loads((d/'state.json').read_text());cfg=json.loads((d/'config.proposed.json').read_text())
    target=pathlib.Path(s['published_config'])
    if a.action=='rollback':
        atomic(target,{'enabled':False})
        args=['docker','compose','-p','ljc-weknora-production','-f',str(d/'compose.before.json'),'up','-d','--no-deps','--force-recreate','app']
        run(args);CANON.write_bytes((d/'compose.before.json').read_bytes())
        s['phase']='rolled_back';write(d/'state.json',s)
        print('Rolled back application/entrypoint; retained API tables, evidence and dedicated viewer.')
        return
    if s['phase']!='prepared':raise RuntimeError('inspect existing release phase before retrying; never blindly repeat apply')
    if app['Image']!=s['old_image'] or inspect('weknora-app:anonymous-api-20261104')['Id']!=s['new_image']:
        raise RuntimeError('image drift; prepare again')
    frontend=inspect('WeKnora-frontend-custom-preview')
    if frontend['NetworkSettings']['Networks'][s['network']]['IPAddress']!=s['frontend_proxy']:raise RuntimeError('proxy IP changed; prepare again')
    version=sql(env,"SELECT version::text || ':' || dirty::text FROM ljc_schema_migrations")
    if version!='4:false':raise RuntimeError('expected clean LJC version4; inspect migration state')
    # Fresh backup immediately before the schema/application cutover.
    with (d/'database.before.dump').open('wb') as f:
        process=subprocess.run(['docker','exec',env['DB_HOST'],'pg_dump','-U',env['DB_USER'],'-d',env['DB_NAME'],'-Fc'],stdout=f,stderr=subprocess.PIPE)
        if process.returncode:raise RuntimeError('database backup failed; no cutover performed')
    (d/'database.before.dump').chmod(0o600)
    with (d/'database.before.dump').open('rb') as f:
        verify=subprocess.run(['docker','exec','-i',env['DB_HOST'],'pg_restore','--list'],stdin=f,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE)
        if verify.returncode:raise RuntimeError('backup archive validation failed; no cutover performed')
    # Migrate only the new additive table and its exact migration version atomically.
    migration=(REPO/'migrations/ljc/postgres/000005_anonymous_api.up.sql').read_text()
    sql(env,"BEGIN; LOCK TABLE ljc_schema_migrations IN EXCLUSIVE MODE; "
        "DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM ljc_schema_migrations WHERE version=4 AND NOT dirty) THEN RAISE EXCEPTION 'migration state changed'; END IF; END $$; "+migration+
        " DELETE FROM ljc_schema_migrations; INSERT INTO ljc_schema_migrations(version,dirty) VALUES(5,false); COMMIT;")
    s['phase']='schema_applied';write(d/'state.json',s)
    viewer=cfg['user_id'];tenant=str(cfg['tenant_id'])
    sql(env,"BEGIN; INSERT INTO users(id,username,email,password_hash,tenant_id,is_active,is_system_admin,can_access_all_tenants,created_at,updated_at) VALUES ('"+viewer+"','anonymous-api-executor','anonymous-api-"+viewer+"@local.invalid','!interactive-login-disabled',"+tenant+",true,false,false,now(),now()); INSERT INTO tenant_members(user_id,tenant_id,role,status,joined_at,created_at,updated_at) VALUES ('"+viewer+"',"+tenant+",'viewer','active',now(),now(),now()); COMMIT;")
    atomic(target,{'enabled':False})
    s['phase']='cutover_started';write(d/'state.json',s)
    run(compose(d)+['up','-d','--no-deps','--force-recreate','app'])
    for attempt in range(60):
        try:
            if requests.get('http://127.0.0.1:18081/health',timeout=2).status_code==200:break
        except requests.RequestException:pass
        time.sleep(1)
    else:raise RuntimeError('new application not ready; run explicit rollback')
    if inspect('WeKnora-app')['Image']!=s['new_image']:raise RuntimeError('runtime image mismatch')
    CANON.write_bytes((d/'compose.api.json').read_bytes())
    atomic(target,cfg)
    s['phase']='deployed_needs_client_acceptance';write(d/'state.json',s)
    print('Deployed; run SDK/curl and existing webpage acceptance from the approved source before declaring live.')
if __name__=='__main__':main()
