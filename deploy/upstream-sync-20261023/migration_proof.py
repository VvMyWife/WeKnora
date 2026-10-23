
from pathlib import Path
import subprocess as sp, json, secrets, time, os, hashlib
ROOT=Path(__file__).resolve().parents[2]
PRIVATE=Path('/home/ubuntu/weknora-fork-migration/20261023-upstream')
NAME='ljc-sync-20261023-postgres'; NET='ljc-sync-20261023'; LABEL='upstream-20261023'
def run(args,**kw): return sp.run(args,check=True,**kw)
def out(args): return sp.check_output(args,text=True).strip()
def inspect(name): return json.loads(out(['docker','inspect',name]))[0]
def sql(db,statement):
 return out(['docker','exec',NAME,'psql','-X','-A','-t','-v','ON_ERROR_STOP=1','-U','postgres','-d',db,'-c',statement])
def counts(db):
 result={}
 for table in ['knowledges','chunks','wiki_pages','archive_mapping_versions','archive_mapping_links','external_chunk_records','message_citation_evidence','enrichment_recoveries']:
  if sql(db,"SELECT to_regclass('public."+table+"') IS NOT NULL")=='t':
   result[table]=int(sql(db,'SELECT count(*) FROM '+table))
 return result
def main():
 assert not (PRIVATE/'migration-proof.json').exists(),'Proof already complete; inspect result instead of replaying'
 assert not sp.run(['docker','inspect',NAME],stdout=sp.DEVNULL,stderr=sp.DEVNULL).returncode==0,'Existing proof container: inspect before resume'
 PRIVATE.mkdir(mode=0o700,exist_ok=True); os.chmod(PRIVATE,0o700)
 source=inspect('WeKnora-postgres-20261021')
 assert source['State']['Running']
 password=secrets.token_hex(24)
 env=PRIVATE/'proof-postgres.env';env.write_text('POSTGRES_USER=postgres\nPOSTGRES_PASSWORD='+password+'\nPOSTGRES_DB=postgres\n');env.chmod(0o600)
 run(['docker','network','create','--label','ljc.operation='+LABEL,NET],stdout=sp.DEVNULL)
 run(['docker','volume','create','--label','ljc.operation='+LABEL,NAME],stdout=sp.DEVNULL)
 run(['docker','run','-d','--name',NAME,'--network',NET,'--label','ljc.operation='+LABEL,'--env-file',str(env),'--shm-size','4g','-v',NAME+':/var/lib/postgresql/data',source['Image']],stdout=sp.DEVNULL)
 target=inspect(NAME);assert target['Config']['Labels']['ljc.operation']==LABEL
 assert target['HostConfig']['PortBindings'] in ({},None)
 assert all(m['Source'] not in {s['Source'] for s in source['Mounts']} for m in target['Mounts'])
 for _ in range(60):
  if sp.run(['docker','exec',NAME,'pg_isready','-U','postgres'],stdout=sp.DEVNULL,stderr=sp.DEVNULL).returncode==0:break
  time.sleep(1)
 else: raise RuntimeError('Proof PG not ready')
 snapshot=PRIVATE/'production-128-snapshot.dump'
 assert not snapshot.exists(),'Snapshot already exists: inspect before retry'
 with snapshot.open('wb') as handle:
  run(['docker','exec','WeKnora-postgres-20261021','sh','-ec','pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc'],stdout=handle)
 snapshot.chmod(0o600)
 scenarios=[('fresh',None),('current128',snapshot),('legacy94',Path('/home/ubuntu/weknora-fork-migration/20261020-upstream/database.dump'))]
 results={}
 for key,dump in scenarios:
  db='ljc_sync_'+key
  run(['docker','exec',NAME,'createdb','-T','template0','-U','postgres',db])
  if dump:
   sql(db,'CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;')
   with dump.open('rb') as handle:
    run(['docker','exec','-i','-e','PGOPTIONS=-c maintenance_work_mem=512MB -c max_parallel_maintenance_workers=0',NAME,'pg_restore','--no-owner','--no-privileges','--exit-on-error','-U','postgres','-d',db],stdin=handle)
  else:
   sql(db,'CREATE EXTENSION IF NOT EXISTS vector; CREATE EXTENSION IF NOT EXISTS pg_search;')
  before=counts(db); before_version=sql(db,"SELECT CASE WHEN to_regclass('schema_migrations') IS NULL THEN 0 ELSE 1 END")
  envpath=PRIVATE/(key+'-migration.env')
  envpath.write_text('LJC_MIGRATION_DSN=postgres://postgres:'+password+'@'+NAME+':5432/'+db+'?sslmode=disable\n')
  envpath.chmod(0o600)
  cmd=['docker','run','--rm','--network',NET,'--label','ljc.operation='+LABEL,'--env-file',str(envpath),'-v',str(ROOT)+':/src','-v','/home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod:ro','-w','/src','weknora-builder:upstream-20261020','/src/runtime/upstream-20261023/ljc-migrate']
  if key=='legacy94':run(cmd+['-mode','bootstrap-legacy'],stdout=(PRIVATE/(key+'-bootstrap.log')).open('w'))
  for index in (1,2):
   with (PRIVATE/(key+'-up-'+str(index)+'.log')).open('w') as handle:run(cmd+['-mode','up'],stdout=handle)
  official=sql(db,'SELECT version::text || chr(58) || dirty::text FROM schema_migrations')
  custom=sql(db,'SELECT version::text || chr(58) || dirty::text FROM ljc_schema_migrations')
  assert official=='133:false',official
  assert custom=='4:false',custom
  after=counts(db)
  if dump:assert before==after,(key,before,after)
  results[key]={'official':official,'custom':custom,'counts_before':before,'counts_after':after,'repeated_up':'passed'}
  (PRIVATE/('migration-'+key+'-result.json')).write_text(json.dumps(results[key],indent=2))
  print('MIGRATION_PROOF_OK',key,official,custom,flush=True)
 result={'target':'bc1ade36','production_modified':False,'container':NAME,'results':results}
 (PRIVATE/'migration-proof.json').write_text(json.dumps(result,indent=2))
 print('ALL_MIGRATION_PROOFS_PASSED',flush=True)
if __name__=='__main__':
 import sys
 if '--status' in sys.argv:
  result=PRIVATE/'migration-proof.json'
  print(result.read_text() if result.exists() else json.dumps({'complete':False,'private_dir':str(PRIVATE)}))
 else:main()
