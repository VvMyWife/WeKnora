#!/usr/bin/env python3
"""Repair only housekeeping-failed, parsed documents owning live graph work."""
import argparse
import datetime
import json
import pathlib
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--apply', action='store_true')
args = parser.parse_args()
kb = '8575ca20-5e7e-48c4-9fc0-786d979fdf5d'
env = dict(x.split('=', 1) for x in json.loads(subprocess.check_output(['docker','inspect','WeKnora-app']))[0]['Config']['Env'] if '=' in x)
base = ['docker','exec','-e','REDISCLI_AUTH='+env.get('REDIS_PASSWORD',''),'WeKnora-redis','redis-cli','-n',env.get('REDIS_DB','0'),'--raw']
paused = subprocess.check_output(base+['GET',f'weknora:processing:{{10000:{kb}}}:paused'],text=True).strip()
if paused != '1':
    raise SystemExit('refusing counter reconciliation while the target KB is not paused')
script = r'''local ids={};for _,suffix in ipairs({'pending','active'}) do for _,id in ipairs(redis.call('LRANGE',KEYS[1]..suffix,0,-1)) do ids[id]=true end end
for _,suffix in ipairs({'retry','scheduled'}) do for _,id in ipairs(redis.call('ZRANGE',KEYS[1]..suffix,0,-1)) do ids[id]=true end end
local out={};for id,_ in pairs(ids) do local m=redis.call('HGET',KEYS[1]..'t:'..id,'msg') or '';local k=string.match(m,'"knowledge_id":"([^"]+)"');local a=string.match(m,'"attempt":(%d+)') or '0';if k then local key=k..':'..a;out[key]=(out[key] or 0)+1 end end;return cjson.encode(out)'''
counts = json.loads(subprocess.check_output(base+['EVAL',script,'1','asynq:{graph}:'],text=True))
def sql(query):
    return subprocess.check_output(['docker','exec','WeKnora-postgres','psql','-X','-A','-t','-v','ON_ERROR_STOP=1','-U','weknora','-d','weknora','-c',query],text=True)
query = f"""SELECT row_to_json(t) FROM (
 SELECT k.id,k.pending_subtasks_count,k.parse_status,k.error_message,
 (SELECT max(attempt) FROM knowledge_processing_spans s WHERE s.knowledge_id=k.id) AS attempt
 FROM knowledges k WHERE k.knowledge_base_id='{kb}' AND k.deleted_at IS NULL
 AND k.parse_status='failed' AND k.summary_status='completed'
 AND k.error_message LIKE 'task stuck in processing %'
 AND EXISTS(SELECT 1 FROM knowledge_processing_spans s WHERE s.knowledge_id=k.id AND s.name='docreader' AND s.status='done')
 AND EXISTS(SELECT 1 FROM knowledge_processing_spans s WHERE s.knowledge_id=k.id AND s.name='embedding' AND s.status='done')
 )t"""
manifest = []
for line in sql(query).splitlines():
    row=json.loads(line);n=counts.get(f"{row['id']}:{row['attempt']}",0)
    if n:
        row['restored_pending']=n;manifest.append(row)
folder=pathlib.Path('backups/enrichment-20261016');folder.mkdir(parents=True,exist_ok=True)
path=folder/('counter-reconciliation-'+datetime.datetime.now().strftime('%Y%m%dT%H%M%S')+('.apply.json' if args.apply else '.dry-run.json'))
path.write_text(json.dumps(manifest,ensure_ascii=False,indent=2));path.chmod(0o640)
if args.apply:
    for row in manifest:
        # IDs originate from PostgreSQL, nevertheless enforce a narrow UUID alphabet.
        import re
        if not re.fullmatch('[0-9a-f-]{36}',row['id']): raise ValueError('invalid knowledge ID')
        sql(f"""UPDATE knowledges SET parse_status='finalizing',pending_subtasks_count={row['restored_pending']},error_message='',updated_at=NOW()
        WHERE id='{row['id']}' AND parse_status='failed' AND error_message LIKE 'task stuck in processing %'
        AND (SELECT max(attempt) FROM knowledge_processing_spans WHERE knowledge_id='{row['id']}')={row['attempt']};""")
print(json.dumps({'apply':args.apply,'documents':len(manifest),'pending':sum(r['restored_pending'] for r in manifest),'manifest':str(path)},ensure_ascii=False))
