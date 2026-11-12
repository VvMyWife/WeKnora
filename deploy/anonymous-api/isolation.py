#!/usr/bin/env python3
"""Build/start only independently named acceptance services. No production mutations."""
import argparse,json,pathlib,subprocess,secrets,shutil,os,time
REPO=pathlib.Path(__file__).resolve().parents[2]
STATE=pathlib.Path('/home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance')
STATE.mkdir(parents=True,exist_ok=True,mode=0o700)
def cmd(args,check=True):
 p=subprocess.run(args,capture_output=True,text=True)
 if check and p.returncode:raise RuntimeError('operation failed: '+args[0]+' '+(p.stderr or p.stdout)[-1500:])
 return p

def inspect(name):
 p=cmd(['docker','inspect',name],False)
 return json.loads(p.stdout)[0] if p.returncode==0 else None

def start(name,args):
 old=inspect(name)
 if old:
  if old['Config']['Labels'].get('ljc.acceptance')!='anonymous-api':raise RuntimeError('unowned container: '+name)
  desired=inspect(args[-1])
  if desired and old['Image']!=desired['Id']:
   cmd(['docker','rm','-f',name]);cmd(['docker','run','-d','--name',name,'--label','ljc.acceptance=anonymous-api',*args])
  elif not old['State']['Running']:cmd(['docker','start',name])
 else:cmd(['docker','run','-d','--name',name,'--label','ljc.acceptance=anonymous-api',*args])

def main():
 app=inspect('WeKnora-app');assert app
 runtime=STATE/'runtime.json'
 if runtime.exists():s=json.loads(runtime.read_text())
 else:
  s={'pg_password':secrets.token_urlsafe(30),'jwt_secret':secrets.token_urlsafe(48),'base_image':app['Image']};runtime.write_text(json.dumps(s));runtime.chmod(0o600)
 if s['base_image']!=app['Image']:raise RuntimeError('production changed since acceptance setup; review baseline')
 ctx=STATE/'image';ctx.mkdir(exist_ok=True)
 binary=REPO/'runtime/WeKnora.anonymous-api'
 if not binary.is_file():raise RuntimeError('build binary first')
 shutil.copy2(binary,ctx/'WeKnora')
 migration_dest=ctx/'ljc'
 assert migration_dest.resolve().is_relative_to(STATE.resolve())
 if migration_dest.exists():shutil.rmtree(migration_dest)
 shutil.copytree(REPO/'migrations/ljc',migration_dest)
 tag='weknora-app:anonymous-api-20261104'
 base_tag='weknora-app:anonymous-api-base-20261104'
 cmd(['docker','image','tag',s['base_image'],base_tag])
 build=cmd(['docker','build','--build-arg','BASE_IMAGE='+base_tag,'--label','org.opencontainers.image.revision='+cmd(['git','-C',str(REPO),'rev-parse','HEAD']).stdout.strip(),'-f',str(REPO/'deploy/anonymous-api/Dockerfile'),'-t',tag,str(ctx)])
 (STATE/'image-build.log').write_text(build.stdout+build.stderr)
 if cmd(['docker','network','inspect','weknora-api-test-net'],False).returncode:cmd(['docker','network','create','weknora-api-test-net'])
 net=json.loads(cmd(['docker','network','inspect','weknora-api-test-net']).stdout)[0];gateway=net['IPAM']['Config'][0]['Gateway']
 for sub in ['files','tmp','config']:(STATE/sub).mkdir(exist_ok=True)
 cfg=STATE/'config/anonymous-api.json'
 if not cfg.exists():cfg.write_text(json.dumps({'enabled':False}))
 start('weknora-api-test-pg',['--network','weknora-api-test-net','--network-alias','api-pg','-e','POSTGRES_USER=weknora','-e','POSTGRES_PASSWORD='+s['pg_password'],'-e','POSTGRES_DB=weknora','--memory','2g','docker.m.daocloud.io/paradedb/paradedb:v0.25.11-pg17'])
 start('weknora-api-test-redis',['--network','weknora-api-test-net','--network-alias','api-redis','--memory','256m','redis:7.0-alpine'])
 for attempt in range(60):
  if cmd(['docker','exec','weknora-api-test-pg','pg_isready','-U','weknora','-d','weknora'],False).returncode==0:break
  time.sleep(1)
 else:raise RuntimeError('isolated PostgreSQL is not ready')
 env=dict(x.split('=',1) for x in app['Config']['Env'] if '=' in x)
 env.update(AUTO_MIGRATE='true',DB_HOST='api-pg',DB_PORT='5432',DB_USER='weknora',DB_PASSWORD=s['pg_password'],DB_NAME='weknora',DB_DRIVER='postgres',REDIS_ADDR='api-redis:6379',REDIS_PASSWORD='',REDIS_DB='0',NEO4J_ENABLE='false',JWT_SECRET=s['jwt_secret'],APP_EXTERNAL_URL='http://127.0.0.1:28481',WEKNORA_ANONYMOUS_API_CONFIG='/app/api-config/anonymous-api.json',LLM_DEBUG_LOG='false',LOG_LEVEL='info',DISABLE_REGISTRATION='false',STORAGE_TYPE='local',LOCAL_STORAGE_BASE_DIR='/data/files')
 args=['--network','weknora-api-test-net','-p','127.0.0.1:28481:8080','--memory','4g','--cpus','4','--add-host','host.docker.internal:host-gateway']
 for k,v in env.items():args+=['-e',k+'='+v]
 # Keep source config read-only; no storage/Redis/Postgres/queue is shared.
 config=next(m['Source'] for m in app['Mounts'] if m['Destination']=='/app/config')
 args+=['-v',config+':/app/config:ro','-v',str(STATE/'config')+':/app/api-config:ro','-v',str(STATE/'files')+':/data/files','-v',str(STATE/'tmp')+':/tmp/docreader',tag]
 start('weknora-api-acceptance',args)
 s['gateway']=gateway;s['tag']=tag;runtime.write_text(json.dumps(s));runtime.chmod(0o600)
 print(json.dumps({'candidate':'http://127.0.0.1:28481','network_gateway':gateway,'image':tag,'production_changed':False}))
if __name__=='__main__':main()
