#!/usr/bin/env python3
"""Creates data ONLY in the independently named localhost acceptance instance."""
import json,pathlib,subprocess,secrets,time,uuid,hashlib,zipfile,io
import requests
ROOT=pathlib.Path(__file__).resolve().parents[2]
STATE=pathlib.Path('/home/ubuntu/RAG_Weknora/runtime/anonymous-api-acceptance')
BASE='http://127.0.0.1:28481'
assert BASE.endswith(':28481')

def db(sql):
 p=subprocess.run(['docker','exec','weknora-api-test-pg','psql','-XAt','-v','ON_ERROR_STOP=1','-U','weknora','-d','weknora','-c',sql],capture_output=True,text=True)
 if p.returncode:raise RuntimeError(p.stderr[-800:])
 return p.stdout.strip()
def call(s,method,path,**kw):
 r=s.request(method,BASE+'/api/v1'+path,timeout=kw.pop('timeout',120),**kw)
 if r.status_code>=400:raise RuntimeError(method+' '+path+' '+str(r.status_code)+' '+r.text[:400])
 return r.json()
def save(x):
 p=STATE/'fixture.json';p.write_text(json.dumps(x,ensure_ascii=False,indent=2));p.chmod(0o600)
def main():
 for attempt in range(20):
  try:
   if requests.get(BASE+'/health',timeout=2).status_code==200:break
  except requests.RequestException:pass
  time.sleep(1)
 else:raise RuntimeError('isolated app not ready')
 file=STATE/'fixture.json'
 if file.exists():st=json.loads(file.read_text())
 else:st={'email':'api-acceptance-'+uuid.uuid4().hex[:10]+'@example.com','password':secrets.token_urlsafe(24)};save(st)
 s=requests.Session()
 if not st.get('registered'):
  call(s,'POST','/auth/register',json={'username':'API acceptance','email':st['email'],'password':st['password']});st['registered']=True;save(st)
 auth=call(s,'POST','/auth/login',json={'email':st['email'],'password':st['password']});token=auth.get('token') or auth.get('data',{}).get('token');assert token;s.headers['Authorization']='Bearer '+token
 # The isolated tenant starts with no models. Copy provider definitions only, never documents.
 app=json.loads(subprocess.check_output(['docker','inspect','WeKnora-app']))[0];env=dict(x.split('=',1) for x in app['Config']['Env'] if '=' in x)
 q="select json_agg(t) from (select name,type,source,parameters from models where tenant_id=10000 and deleted_at is null and name in ('qwen35-9b','bge-m3','bge-reranker-v2-m3'))t"
 result=subprocess.check_output(['docker','exec',env['DB_HOST'],'psql','-XAt','-U',env['DB_USER'],'-d',env['DB_NAME'],'-c',q],text=True)
 st.setdefault('models',{})
 for m in json.loads(result):
  if m['name'] in st['models']:continue
  m['parameters'].pop('api_key',None);m['parameters'].pop('app_secret',None);m['parameters'].pop('app_id',None)
  out=call(s,'POST','/models',json=m)['data'];st['models'][m['name']]=out['id'];save(st)
 if not st.get('kb'):
  kb=call(s,'POST','/knowledge-bases',json={'name':'独立免密钥API验收','type':'document','description':'仅用于隔离环境接口验收','embedding_model_id':st['models']['bge-m3'],'summary_model_id':st['models']['qwen35-9b'],'extract_config':{'enabled':False},'wiki_config':{'enabled':False},'question_generation_config':{'enabled':False}})['data'];st['kb']=kb['id'];st['tenant']=kb['tenant_id'];save(st)
 if not st.get('batch'):
  text1='独立API验收资料：档案编号为 API-FIXTURE-20261104。检查周期为每27天一次，归档前应核对页数。'
  text2='独立API验收资料要求备份保留3份，分别存放于甲、乙、丙三个测试位置。本资料仅用于接口测试。'
  raw=(text1+'\n\n'+text2).encode();sha=lambda b:hashlib.sha256(b).hexdigest()
  document={'id':'anon-fixture','title':'独立API验收资料','relative_path':'samples/anon.md','version':1,'source_sha256':sha(raw),'parsed_file':'files/anon.md','original_file':'files/anon.txt'}
  units=[];chunks=[]
  for i,text in enumerate([text1,text2]):
   units.append({'id':'u'+str(i),'document_id':'anon-fixture','unit_id':'u'+str(i),'page_index':0,'order':i,'block_type':'text','text':text})
   chunks.append({'id':'c'+str(i),'document_id':'anon-fixture','order':i,'content':text,'retrieval_text':text,'context_text':'','token_count':80,'metadata':{},'source_anchors':[{'unit_id':'u'+str(i)}]})
  jl=lambda rows:''.join(json.dumps(x,ensure_ascii=False)+'\n' for x in rows).encode()
  files={'documents.jsonl':jl([document]),'chunks.jsonl':jl(chunks),'source_units.jsonl':jl(units),'files/anon.md':raw,'files/anon.txt':raw}
  manifest={'format':'weknora-chunks-v1','schema_sha256':sha((ROOT/'internal/externalchunks/manifest.schema.json').read_bytes()),'package_id':str(uuid.uuid4()),'source_system':'anonymous-api-acceptance','dataset_id':st['kb'],'generation_id':'fixture-v1','producer_commit':'acceptance-only','generation_config':{},'counts':{'documents':1,'chunks':2,'source_units':2},'files':{name:{'sha256':sha(b),'size':len(b)} for name,b in files.items()}}
  out=io.BytesIO()
  with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED) as z:
   z.writestr('manifest.json',json.dumps(manifest).encode())
   for name,b in files.items():z.writestr(name,b)
  response=call(s,'POST',f"/knowledge-bases/{st['kb']}/knowledge/external-chunks/batches",files={'file':('fixture.zip',out.getvalue(),'application/zip')})['data'];st['batch']=response['batch']['id'];save(st)
 response=call(s,'GET',f"/knowledge-bases/{st['kb']}/knowledge/external-chunks/batches/{st['batch']}")['data']
 for item in response['documents']:
  if item.get('status') not in ['completed','processing','queued']:
   call(s,'POST',f"/knowledge-bases/{st['kb']}/knowledge/external-chunks/batches/{st['batch']}/documents/{item['id']}/start",json={})
 if not st.get('viewer'):
  st['viewer']=str(uuid.uuid4());save(st)
 if db("SELECT COUNT(*) FROM users WHERE id='"+st['viewer']+"'")=='0':
  db("INSERT INTO users(id,username,email,password_hash,tenant_id,is_active,is_system_admin,can_access_all_tenants,created_at,updated_at) VALUES ('"+st['viewer']+"','anonymous-api-executor','anonymous-api-executor@local.invalid','!interactive-login-disabled',"+str(st['tenant'])+",true,false,false,now(),now()); INSERT INTO tenant_members(user_id,tenant_id,role,status,joined_at,created_at,updated_at) VALUES ('"+st['viewer']+"',"+str(st['tenant'])+",'viewer','active',now(),now(),now());")
 # Explicitly configure the built-in quick-answer model in this isolated tenant.
 db("INSERT INTO custom_agents(id,tenant_id,name,is_builtin,config,created_at,updated_at) VALUES ('builtin-quick-answer',"+str(st['tenant'])+",'API fixture quick answer',true,'"+json.dumps({'agent_mode':'quick-answer','model_id':st['models']['qwen35-9b'],'rerank_model_id':st['models']['bge-reranker-v2-m3'],'temperature':0.1,'thinking':False,'max_completion_tokens':1024,'citation_enabled':True,'memory_enabled':False})+"'::json,now(),now()) ON CONFLICT (id,tenant_id) DO NOTHING;")
 runtime=json.loads((STATE/'runtime.json').read_text())
 config={'enabled':True,'tenant_id':st['tenant'],'user_id':st['viewer'],'allowed_cidrs':['127.0.0.1/32',runtime['gateway']+'/32','10.40.0.34/32'],'profiles':[{'id':'fixture-qa','agent_id':'builtin-quick-answer','knowledge_base_ids':[st['kb']],'max_tokens':1024}]}
 (STATE/'config/anonymous-api.json').write_text(json.dumps(config,ensure_ascii=False,indent=2))
 print(json.dumps({'fixture_kb':st['kb'],'tenant':st['tenant'],'sample_batch':st['batch'],'viewer':st['viewer'],'formal_data_changed':False}))
if __name__=='__main__':main()
