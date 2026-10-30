import runpy,json,time,urllib.request,subprocess
from pathlib import Path
P=Path(__file__).parent
n=runpy.run_path(str(P/'acceptance-api-setup.py'));req=n['req'];state=json.loads((P/'acceptance-api-state.json').read_text())
fixtures=Path('/home/ubuntu/RAG_Weknora-upstream-20261029/runtime/upstream-20261029/fixtures')
settings=req('GET','/tenants/kv/default-settings')['data']
desired=dict(settings,chat_model_id=state['models']['KnowledgeQA'],utility_model_id=state['models']['KnowledgeQA'],embedding_model_id=state['models']['Embedding'],rerank_model_id=state['models']['Rerank'])
req('PUT','/tenants/kv/default-settings',desired)
saved=req('GET','/tenants/kv/default-settings')['data']
assert all(saved[k]==desired[k] for k in ['chat_model_id','utility_model_id','embedding_model_id','rerank_model_id'])
models=req('GET','/models')['data']
vlm=next((m for m in models if m['name']=='qwen35-9b' and m['type']=='VLM'),None)
if not vlm:
 vlm=req('POST','/models',{'name':'qwen35-9b','type':'VLM','source':'remote','parameters':{'base_url':'http://10.20.0.15:8101/v1','api_key':'acceptance-local','provider':'generic','interface_type':'openai','max_concurrency':1,'extra_config':{'max_tokens':'8192'}}})['data']
kb=req('POST','/knowledge-bases',{'name':'Isolated formats OCR 20261029','type':'document','embedding_model_id':state['models']['Embedding'],'summary_model_id':state['models']['KnowledgeQA'],'chunking_config':{'chunk_size':512,'chunk_overlap':50},'indexing_strategy':{'vector_enabled':True,'keyword_enabled':True,'graph_enabled':False,'wiki_enabled':False},'vlm_config':{'enabled':True,'model_id':vlm['id']},'image_processing_config':{'model_id':vlm['id'],'image_pipeline':'default','image_pipeline_params':{'default.ocr_enabled':True,'default.caption_enabled':False}}})['data']
def upload(file):
 boundary='ljc-20261029-proof'
 body=(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{file.name}"\r\nContent-Type: application/octet-stream\r\n\r\n').encode()+file.read_bytes()+f'\r\n--{boundary}--\r\n'.encode()
 r=urllib.request.Request(n['BASE']+f'/knowledge-bases/{kb["id"]}/knowledge/file',data=body,headers={'Authorization':'Bearer '+n['token'],'Content-Type':'multipart/form-data; boundary='+boundary},method='POST')
 with urllib.request.urlopen(r,timeout=90) as h:return json.load(h)['data']
docs={name:upload(fixtures/name)['id'] for name in ['table.csv','table.xlsx','vector.pptx','scan.pdf']}
results={}
for name,kid in docs.items():
 for _ in range(150):
  d=req('GET','/knowledge/'+kid)['data'];status=d.get('parse_status')
  if status in ['completed','failed']:break
  time.sleep(2)
 assert status=='completed',(name,d.get('error_message'))
 results[name]={'id':kid,'parse_status':status}
 print('FORMAT_PARSE_OK',name,flush=True)
scan=docs['scan.pdf'];vector=docs['vector.pptx']
def chunks(kid):
 statement="SELECT COALESCE(json_agg(json_build_object('type',chunk_type,'content',content)), '[]') FROM chunks WHERE knowledge_id='"+kid+"' AND deleted_at IS NULL"
 raw=subprocess.check_output(['docker','exec','ljc-sync-20261029-postgres','psql','-U','postgres','-d','ljc_sync_fresh','-Atc',statement],text=True)
 return json.loads(raw)
for _ in range(150):
 sc=chunks(scan)
 if any(x['type']=='image_ocr' and '17' in (x['content'] or '') for x in sc):break
 time.sleep(2)
else:raise AssertionError('scanned PDF OCR has no expected fact')
vc=chunks(vector)
assert any('resource://' in (x['content'] or '') or x['type'] in ['image_ocr','image_caption'] for x in vc),'PPTX lost vector media'
for name in ['table.csv','table.xlsx']:
 cs=chunks(docs[name]);assert any('17' in (x['content'] or '') for x in cs),name+' lost table fact'
results.update(default_settings_roundtrip=True,scan_ocr_fact_17=True,pptx_vector_media_preserved=True)
(P/'formats-defaults-results.json').write_text(json.dumps(results,indent=2))
print('FORMATS_DEFAULTS_OCR_OK',flush=True)
