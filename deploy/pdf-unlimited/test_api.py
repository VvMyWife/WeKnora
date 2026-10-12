#!/usr/bin/env python3
"""End-to-end checks using an existing valid session, kept only in memory.
Creates/resumes one independent test KB; does not reprocess business documents.
Run with /home/ubuntu/RAG/runtime/eval-venv/bin/python.
"""
import json,subprocess,shlex,time
from pathlib import Path
import requests
root=Path(__file__).resolve().parents[2]
state_path=root/'runtime/api-acceptance-state.json'
def sql(q):
 return subprocess.check_output(['docker','exec','WeKnora-postgres','sh','-c','psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tA -c '+shlex.quote(q)],text=True).strip()
token=sql("select token from auth_tokens where token_type='access_token' and expires_at>now() and not is_revoked order by created_at desc limit 1")
assert token,'No active application session; sign in once before testing'
s=requests.Session();s.headers['Authorization']='Bearer '+token
base='http://127.0.0.1:18081/api/v1'
def api(method,path,**kw):
 response=s.request(method,base+path,timeout=60,**kw)
 if not response.ok: raise RuntimeError(f'{method} {path}: HTTP {response.status_code}: {response.text[:300]}')
 return response.json()
state=json.loads(state_path.read_text()) if state_path.exists() else {'documents':{}}
def save():state_path.write_text(json.dumps(state,indent=2,ensure_ascii=False))
if 'kb' not in state:
 source=api('GET','/knowledge-bases/d620fda5-6c61-4569-88c8-34a71a6e8449')['data']
 cfg={k:source[k] for k in ['embedding_model_id','vector_store_id','storage_backend_id','storage_provider_config','chunking_config'] if k in source}
 cfg.update(name='PDF全页上线验收 20261010',type='document',description='独立合成测试文件，不包含历史业务样例',summary_model_id=source['summary_model_id'],indexing_strategy={'vector_enabled':True,'keyword_enabled':True,'wiki_enabled':False,'graph_enabled':False})
 cfg['chunking_config']['parser_engine_rules']=[{'file_types':['pdf'],'engine':'paddleocr_vl'}]
 state['kb']=api('POST','/knowledge-bases',json=cfg)['data']['id'];save()
# Exercise recovery only for this independent test KB (an earlier run left
# summary unset intentionally); the business knowledge base remains read-only.
current=api('GET','/knowledge-bases/'+state['kb'])['data']
if not current.get('summary_model_id'):
 source=api('GET','/knowledge-bases/d620fda5-6c61-4569-88c8-34a71a6e8449')['data']
 current['summary_model_id']=source['summary_model_id']
 api('PUT','/knowledge-bases/'+state['kb'],json={'name':current['name'],'description':current.get('description',''),'config':current})
for name,ident in state['documents'].items():
 doc=api('GET','/knowledge/'+ident)['data']
 if doc.get('summary_status')=='failed' and doc.get('parse_status')=='completed':
  api('POST','/knowledge/'+ident+'/regenerate-summary',json={})
print('Test KB:',state['kb'],flush=True)
for name in ['single.pdf','sixteen.pdf','blank-page.pdf','scan.pdf','corrupt.pdf','encrypted.pdf']:
 if name in state['documents']:continue
 with (root/'runtime/pdf-fixtures'/name).open('rb') as f:
  result=api('POST',f"/knowledge-bases/{state['kb']}/knowledge/file",files={'file':(name,f,'application/pdf')},data={'enable_multimodel':'false','metadata':json.dumps({'folder_path':'failures' if name in ['corrupt.pdf','encrypted.pdf'] else 'valid'})})
 state['documents'][name]=result['data']['id'];save()
 print('Uploaded',name,flush=True)
for folder in ['valid','failures']:
 ids=[ident for name,ident in state['documents'].items() if (name in ['corrupt.pdf','encrypted.pdf'])==(folder=='failures')]
 api('POST','/knowledge/folder',json={'kb_id':state['kb'],'knowledge_ids':ids,'folder_path':folder})
# No execution timeout: keep waiting, while saving observable progress for resumption.
last=None
while True:
 overview=api('GET',f"/knowledge-bases/{state['kb']}/knowledge/overview")['data']
 signature=json.dumps(overview['parse'],sort_keys=True)
 if signature!=last:print('Progress:',signature,flush=True);last=signature
 (root/'runtime/api-overview.json').write_text(json.dumps(overview,indent=2,ensure_ascii=False))
 if not any(overview['parse'].get(k,0) for k in ['pending','processing','finalizing']) and not any(overview['summary'].get(k,0) for k in ['pending','processing']):break
 time.sleep(3)
failures=api('GET',f"/knowledge-bases/{state['kb']}/knowledge/failures")
assert failures['page_size']==20 and failures['total']==2,failures
assert len(api('GET',f"/knowledge-bases/{state['kb']}/knowledge/failures?page_size=1")['data'])==1
assert api('GET',f"/knowledge-bases/{state['kb']}/knowledge/failures?stage=docreader")['total']==2
assert api('GET',f"/knowledge-bases/{state['kb']}/knowledge/failures?stage=embedding")['total']==0
for name,ident in state['documents'].items():
 doc=api('GET','/knowledge/'+ident)['data']
 expected='failed' if name in ['corrupt.pdf','encrypted.pdf'] else 'completed'
 assert doc['parse_status']==expected,(name,doc['parse_status'],doc.get('error_message'))
 if expected=='failed':assert int(sql(f"select count(*) from chunks where knowledge_id='{ident}' and deleted_at is null"))==0
 else:
  meta=doc['metadata'];meta=json.loads(meta) if isinstance(meta,str) else meta
  assert meta['pdf_page_integrity']=='verified',(name,meta)
marker='ZEBRA16PAGE16'
ident=state['documents']['sixteen.pdf']
assert int(sql(f"select count(*) from chunks where knowledge_id='{ident}' and content like '%{marker}%' and deleted_at is null"))>0
search=api('POST',f"/knowledge-bases/{state['kb']}/hybrid-search",json={'query_text':marker,'match_count':20,'vector_threshold':0,'keyword_threshold':0})
assert marker in json.dumps(search), 'last-page marker was not retrievable'
report={'last_page_searchable':True,'knowledge_base_id':state['kb'],'overview':overview,'failures':failures,'last_page_in_indexed_chunks':True}
(root/'runtime/api-acceptance.json').write_text(json.dumps(report,indent=2,ensure_ascii=False))
print('PASS: complete pages, indexed last page, invalid PDFs rejected before chunks, failure counts/filter/pagination',flush=True)
