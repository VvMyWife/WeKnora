#!/usr/bin/env python3
"""Read-only boundary and interruption check; resume only explicitly safe jobs."""
import argparse,json,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
def run(*args):return subprocess.check_output(args,text=True).strip()
def main():
 production_state=Path('/home/ubuntu/weknora-fork-migration/20261021-production/state.json')
 if production_state.exists() and json.loads(production_state.read_text()).get('stage') in ('cutover','complete'):
  import importlib.util
  spec=importlib.util.spec_from_file_location('production_selfcheck',ROOT/'scripts/ljc_production_selfcheck.py')
  module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
  print(json.dumps(module.check(),indent=2));return
 p=argparse.ArgumentParser();p.add_argument('--private-dir',type=Path,default=Path('/home/ubuntu/weknora-fork-migration/20261020-upstream'));p.add_argument('--resume-safe-jobs',action='store_true');a=p.parse_args()
 baseline=json.loads((a.private_dir/'baseline-private.json').read_text())
 checks=[]
 for old in baseline['containers']:
  now=json.loads(run('docker','inspect',old['name']))[0]
  assert now['Id']==old['id'] and now['Image']==old['image'],old['name']+' production baseline changed'
  checks.append(old['name']+' unchanged')
 assert run('git','-C','/home/ubuntu/RAG_Weknora','rev-parse','HEAD')==baseline['custom_head'],'Runtime checkout changed'
 ledger=run('docker','exec','WeKnora-postgres','sh','-ec','psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select version,dirty from schema_migrations"')
 assert ledger=='94|f','Production migration ledger changed: '+ledger
 env=dict(x.split('=',1) for x in (a.private_dir/'acceptance.env').read_text().splitlines() if '=' in x and not x.startswith('#'))
 assert env['DB_HOST']=='ljc-upstream-acceptance-postgres' and env['REDIS_ADDR']=='ljc-upstream-acceptance-redis:6379'
 for name in ('app','postgres','redis','neo4j','docreader','official-ui','custom-ui'):
  data=json.loads(run('docker','inspect','ljc-upstream-acceptance-'+name))[0]
  assert data['Config']['Labels'].get('ljc.operation')=='upstream-20261020'
  assert all(m['Type']!='volume' or m['Name'].startswith('ljc-upstream-acceptance-') for m in data['Mounts']),name+' shared volume'
  checks.append(name+': '+data['State']['Status'])
 jobs=[]
 for status in (ROOT/'runtime/upstream/jobs').glob('*.status'):
  state=status.read_text().strip();pidfile=status.with_suffix('.pid');pid=int(pidfile.read_text()) if pidfile.exists() else 0
  proc=Path('/proc')/str(pid)/'cmdline'
  live=proc.exists() and b'ljc_integration_job.sh' in proc.read_bytes() and status.stem.encode() in proc.read_bytes()
  item={'job':status.stem,'state':state,'live':live}
  if state=='running' and not live:
   item['interrupted']=True
   if a.resume_safe_jobs:
    log=status.with_suffix('.resume.log').open('ab')
    subprocess.Popen(['bash',str(ROOT/'scripts/ljc_integration_job.sh'),status.stem],cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
    log.close();item['resume_requested']=True
  jobs.append(item)
 print(json.dumps({'checks':checks,'production_ledger':ledger,'jobs':jobs},indent=2))
if __name__=='__main__':main()
