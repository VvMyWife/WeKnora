import json, urllib.request, urllib.error, pathlib, time, sys
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
env=dict(x.split('=',1) for x in (P/'acceptance.env').read_text().splitlines() if '=' in x and not x.startswith('#'))
BASE='http://127.0.0.1:28281/api/v1'
token=''
def req(method,path,data=None):
 b=json.dumps(data).encode() if data is not None else None
 h={'Content-Type':'application/json'}
 if token:h['Authorization']='Bearer '+token
 try:
  with urllib.request.urlopen(urllib.request.Request(BASE+path,data=b,headers=h,method=method),timeout=90) as r:return json.load(r)
 except urllib.error.HTTPError as e:raise RuntimeError(str(e.code)+' '+e.read().decode()[:1000])
login=req('POST','/auth/login',{'email':env['ACCEPTANCE_USER_EMAIL'],'password':env['ACCEPTANCE_USER_PASSWORD']})
token=login['token']
state=json.loads((P/'acceptance-api-state.json').read_text()) if (P/'acceptance-api-state.json').exists() else {}
state.update({'tenant':login['active_tenant']['id'],'models':{}})
print('LOGIN_OK tenant',state['tenant'],flush=True)
existing=req('GET','/models').get('data',[])
for name,typ,url in [('qwen35-9b','KnowledgeQA','http://10.20.0.15:8101/v1'),('bge-m3','Embedding','http://10.20.0.15:18082/v1'),('bge-reranker-v2-m3','Rerank','http://10.20.0.15:18083')]:
 m=next((m for m in existing if m['name']==name and m['type']==typ),None)
 if not m:
  params={'base_url':url,'api_key':'acceptance-local','interface_type':'openai','provider':'generic','max_concurrency':1}
  if typ=='Embedding':params['embedding_parameters']={'dimension':1024,'supports_dimension_override':True}
  m=req('POST','/models',{'name':name,'type':typ,'source':'remote','parameters':params})['data']
 state['models'][typ]=m['id']
 print('MODEL_OK',name,flush=True)
kbs=req('GET','/knowledge-bases').get('data',[])
kb=next((k for k in kbs if k['name']=='LJC upstream acceptance 20261020'),None)
if not kb:
 kb=req('POST','/knowledge-bases',{'name':'LJC upstream acceptance 20261020','type':'document','embedding_model_id':state['models']['Embedding'],'summary_model_id':state['models']['KnowledgeQA'],'chunking_config':{'chunk_size':512,'chunk_overlap':50},'indexing_strategy':{'vector_enabled':True,'keyword_enabled':True,'graph_enabled':False,'wiki_enabled':False}})['data']
state['kb']=kb['id']
(P/'acceptance-api-state.json').write_text(json.dumps(state,indent=2))
(P/'acceptance-api-state.json').chmod(0o600)
print('KB_OK',state['kb'],flush=True)
