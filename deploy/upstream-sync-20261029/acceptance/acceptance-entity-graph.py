import runpy,json,urllib.request,urllib.error,time,pathlib,secrets,subprocess
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
ns=runpy.run_path(str(P/'acceptance-api-setup.py'));req=ns['req'];token=ns['token'];base=ns['BASE']
state=json.loads((P/'acceptance-api-state.json').read_text()); kb=state['kb']
def upload(name,data,config=None):
 boundary='ljc-scan-acceptance';body=b''
 if config:body+=f'--{boundary}\r\nContent-Disposition: form-data; name="process_config"\r\n\r\n{json.dumps(config)}\r\n'.encode()
 body+=f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{name}"\r\nContent-Type: application/octet-stream\r\n\r\n'.encode()+data+f'\r\n--{boundary}--\r\n'.encode()
 with urllib.request.urlopen(urllib.request.Request(base+f'/knowledge-bases/{kb}/knowledge/file',data=body,headers={'Authorization':'Bearer '+token,'Content-Type':'multipart/form-data; boundary='+boundary},method='POST'),timeout=90) as r:return json.load(r)['data']
results={}
def wait(id,wanted,limit=120):
 for _ in range(limit):
  doc=req('GET',f'/knowledge/{id}')['data'];s=doc.get('parse_status')
  print('DOC_STATUS',id,s,flush=True)
  if s in wanted:return doc
  if s=='failed' and 'failed' not in wanted:raise RuntimeError(json.dumps(doc,ensure_ascii=False)[:1000])
  time.sleep(2)
 raise RuntimeError('task still incomplete')

state=json.loads((P/'acceptance-api-state.json').read_text())
name='LJC entity graph acceptance 20261020'
kbs=req('GET','/knowledge-bases')['data']
found=next((x for x in kbs if x['name']==name),None)
if found is None:
 found=req('POST','/knowledge-bases',{'name':name,'type':'document','embedding_model_id':state['models']['Embedding'],'summary_model_id':state['models']['KnowledgeQA'],'chunking_config':{'chunk_size':512,'chunk_overlap':50},'indexing_strategy':{'vector_enabled':True,'keyword_enabled':True,'graph_enabled':True,'wiki_enabled':False},'extract_config':{'enabled':True,'text':'Extract named people and projects and the manages relationship from the document. Keep original names.','tags':['person','project'],'nodes':[{'name':'Alice','attributes':['person']},{'name':'Aurora','attributes':['project']}],'relations':[{'node1':'Alice','node2':'Aurora','type':'manages'}]}})['data']
kb=found['id']
state['entity_kb']=kb
docs=state.get('entity_documents',[])
if not docs:
 for title,txt in [('Aurora','# Aurora project\nAlice manages the Aurora project. Alice supervises the Aurora archive. The Aurora archive contains 17 boxes.\n'),('Orion','# Orion project\nBob manages the Orion project. Bob supervises the Orion archive. The Orion archive contains 23 boxes.\n')]:
  doc=upload('LJC_ENTITY_'+title+'.md',txt.encode(),{'summary_enabled':False})
  docs.append(doc['id'])
  state['entity_documents']=docs
  (P/'acceptance-api-state.json').write_text(json.dumps(state,indent=2))
for id in docs:wait(id,{'completed'},120)
for attempt in range(150):
 a=req('POST',f'/knowledge-bases/{kb}/graph-search',{'query':'Aurora'})
 b=req('POST',f'/knowledge-bases/{kb}/graph-search',{'query':'Orion'})
 (P/'entity-graph-observed.json').write_text(json.dumps({'kb':kb,'documents':docs,'aurora':a,'orion':b},indent=2))
 an=a['data'].get('entities',[]);bn=b['data'].get('entities',[])
 print('GRAPH_COUNTS',len(an),len(bn),flush=True)
 if an and bn:
  assert a['data']['relationships'] and b['data']['relationships']
  for data,docid in [(a['data'],docs[0]),(b['data'],docs[1])]:
   assert all(x['knowledge_id']==docid for x in data['entities']+data['source_chunks']),data
   chunkids={x['id'] for x in data['source_chunks']}
   assert all(set(x['source_chunk_ids'])<=chunkids for x in data['entities']+data['relationships']),data
  assert {x['id'] for x in an}.isdisjoint(x['id'] for x in bn)
  (P/'entity-graph-results.json').write_text(json.dumps({'kb':kb,'documents':docs,'aurora':a,'orion':b,'source_isolation':True},indent=2))
  print('ENTITY_GRAPH_CREATED_OK',flush=True)
  break
 time.sleep(2)
else:raise RuntimeError('entity graph did not populate')
