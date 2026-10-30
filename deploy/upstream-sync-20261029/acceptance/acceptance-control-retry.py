import runpy,json,time,urllib.request,uuid,subprocess
from pathlib import Path
P=Path(__file__).parent
n=runpy.run_path(str(P/'acceptance-api-setup.py')); req=n['req']
s=json.loads((P/'acceptance-api-state.json').read_text()); kb=s['kb']; kid=s['document']
uuid.UUID(kid)
def sql(q):
 return subprocess.check_output(['docker','exec','ljc-sync-20261029-postgres','psql','-U','postgres','-d','ljc_sync_fresh','-Atc',q],text=True).strip()
def digest():
 return sql("SELECT md5(COALESCE(string_agg(id||content,'' ORDER BY id),'')) FROM chunks WHERE knowledge_id='"+kid+"' AND chunk_type='text' AND deleted_at IS NULL")
before=digest()
control='/knowledge-bases/'+kb+'/knowledge/processing-control'
paused=req('PUT',control,{'paused':True})['data']; assert paused['requested_paused']
try:
 one=req('POST','/knowledge/'+kid+'/retry-failed',{'stage':'summary'})['data']
 two=req('POST','/knowledge/'+kid+'/retry-failed',{'stage':'summary'})['data']
 rows=sql("SELECT count(*),count(DISTINCT logical_key) FROM enrichment_recoveries WHERE knowledge_id='"+kid+"'")
 print('RETRY_ACCEPTED',one,two,'ROWS',rows,flush=True)
 counts=[int(x) for x in rows.split('|')]; assert counts[0]==counts[1] and counts[0]>0
finally:req('PUT',control,{'paused':False})
for _ in range(180):
 d=req('GET','/knowledge/'+kid)['data']
 if d.get('summary_status')=='completed':break
 time.sleep(2)
else:raise AssertionError('summary retry did not complete: '+str(d.get('summary_status')))
assert before==digest(),'summary retry modified text chunks'
new=req('POST','/knowledge-bases',{'name':'LJC control acceptance '+str(time.time_ns()),'type':'document','embedding_model_id':s['models']['Embedding'],'summary_model_id':s['models']['KnowledgeQA'],'indexing_strategy':{'vector_enabled':True,'keyword_enabled':True}})['data']['id']
c='/knowledge-bases/'+new+'/knowledge/processing-control'
req('PUT',c,{'paused':True})
def upload(label):
 b='ljc-control-fixture';text='# '+label+'\nThere are 17 archive boxes.\n'
 body=(f'--{b}\r\nContent-Disposition: form-data; name="file"; filename="{label}.md"\r\nContent-Type: text/markdown\r\n\r\n'+text+f'\r\n--{b}--\r\n').encode()
 r=urllib.request.Request(n['BASE']+'/knowledge-bases/'+new+'/knowledge/file',data=body,headers={'Authorization':'Bearer '+n['token'],'Content-Type':'multipart/form-data; boundary='+b},method='POST')
 with urllib.request.urlopen(r,timeout=30) as h:return json.load(h)['data']['id']
cancel=upload('cancel-proof');resume=upload('resume-proof')
time.sleep(3)
assert req('GET','/knowledge/'+resume)['data']['parse_status']!='completed'
assert req('POST','/knowledge/'+cancel+'/cancel-parse',{})['data']['parse_status']=='cancelled'
assert req('POST','/knowledge/'+cancel+'/cancel-parse',{})['data']['parse_status']=='cancelled'
req('PUT',c,{'paused':False})
for _ in range(180):
 d=req('GET','/knowledge/'+resume)['data']
 if d['parse_status']=='completed':break
 time.sleep(2)
else:raise AssertionError('resume failed '+d['parse_status'])
assert req('GET','/knowledge/'+cancel)['data']['parse_status']=='cancelled'
out={'summary_retry_complete':True,'duplicate_retry_unique_logical_keys':True,'text_chunks_unchanged':True,'pause_resume_complete':True,'cancel_idempotent_and_stays_cancelled':True,'kb':new,'resumed_doc':resume,'cancelled_doc':cancel}
(P/'control-retry-results.json').write_text(json.dumps(out,indent=2))
print('CONTROL_RETRY_OK',flush=True)
