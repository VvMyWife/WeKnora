from pathlib import Path
import runpy,subprocess,json,time,sys
ROOT=Path('/home/ubuntu/RAG_Weknora-upstream-20261029')
m=runpy.run_path(str(ROOT/'deploy/upstream-sync-20261029/migration_proof.py'))
P=m['PRIVATE'];NAME=m['NAME'];NET=m['NET'];out=m['out'];sql=m['sql'];db='ljc_sync_rollback128'
assert not (P/'rollback-proof.json').exists(),'Already complete; inspect evidence'
assert m['inspect'](NAME)['Config']['Labels']['ljc.operation']=='upstream-20261029'
existing=sql('postgres',"SELECT count(*) FROM pg_database WHERE datname='"+db+"'")=='1'
assert not existing or '--resume-dirty-check' in sys.argv,'Existing proof database; inspect first'
baseline_counts=m['counts']('ljc_sync_current128');baseline_hashes=m['fingerprints']('ljc_sync_current128')
if not existing:m['run'](['docker','exec',NAME,'createdb','-T','ljc_sync_current128','-U','postgres',db])
assert sql(db,"SELECT version FROM schema_migrations")=='134'
if existing:assert sql(db,'SELECT dirty FROM schema_migrations')=='t'
else:sql(db,'UPDATE schema_migrations SET dirty=true')
env=P/'rollback-migration.env';env.write_text((P/'current128-migration.env').read_text().replace('/ljc_sync_current128?','/'+db+'?'));env.chmod(0o600)
cmd=['docker','run','--rm','--network',NET,'--label','ljc.operation=upstream-20261029','--env-file',str(env),'-v',str(ROOT)+':/src','-v','/home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod:ro','-w','/src','weknora-builder:upstream-20261020','/src/runtime/upstream-20261029/ljc-migrate','-mode','up']
r=subprocess.run(cmd,capture_output=True,text=True)
(P/'dirty-migration-rejection.log').write_text(r.stdout+r.stderr)
assert r.returncode!=0 and 'dirty' in (r.stdout+r.stderr).lower(),'Dirty ledger was not refused'
assert sql(db,'SELECT dirty FROM schema_migrations')=='t'
assert m['counts'](db)==baseline_counts and m['fingerprints'](db)==baseline_hashes
print('DIRTY_LEDGER_REFUSED_WITHOUT_DATA_CHANGES',flush=True)
# Rehearse snapshot restoration, not a migration force or schema down.
m['run'](['docker','exec',NAME,'dropdb','-U','postgres',db])
m['run'](['docker','exec',NAME,'createdb','-T','template0','-U','postgres',db])
sql(db,'CREATE EXTENSION vector WITH SCHEMA public')
with (P/'production-128-snapshot.dump').open('rb') as h:
 m['run'](['docker','exec','-i','-e','PGOPTIONS=-c maintenance_work_mem=512MB -c max_parallel_maintenance_workers=0',NAME,'pg_restore','--no-owner','--no-privileges','--exit-on-error','-U','postgres','-d',db],stdin=h)
assert sql(db,"SELECT version::text||':'||dirty::text FROM schema_migrations")=='128:false'
assert sql(db,"SELECT version::text||':'||dirty::text FROM ljc_schema_migrations")=='4:false'
assert m['counts'](db)==baseline_counts and m['fingerprints'](db)==baseline_hashes
(P/'rollback-proof.json').write_text(json.dumps({'database':db,'dirty_ledger_refused':True,'dirty_attempt_data_unchanged':True,'snapshot_restore_official':'128:false','snapshot_restore_personal':'4:false','protected_counts_and_content_preserved':True,'production_modified':False},indent=2))
print('ROLLBACK_SNAPSHOT_RESTORE_128_PLUS_4_OK',flush=True)
