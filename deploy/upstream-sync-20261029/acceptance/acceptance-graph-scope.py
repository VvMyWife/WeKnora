import runpy,json,urllib.request,urllib.error,time,pathlib,secrets
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
ns=runpy.run_path(str(P/'acceptance-api-setup.py')); req=ns['req'];token=ns['token']
state=json.loads((P/'acceptance-api-state.json').read_text())
base=ns['BASE'];models=state['models']
kbs=req('GET','/knowledge-bases')['data']
kb=next((x for x in kbs if x['name']=='LJC graph extraction acceptance 20261029'),None)
if not kb:
 kb=req('POST','/knowledge-bases',{'name':'LJC graph extraction acceptance 20261029','type':'document','embedding_model_id':models['Embedding'],'summary_model_id':models['KnowledgeQA'],'indexing_strategy':{'vector_enabled':True,'keyword_enabled':True,'wiki_enabled':True,'graph_enabled':True},'extract_config':{'enabled':True,'text':'Extract persons and projects and their manages or reviews relationships.','tags':['manages','reviews'],'nodes':[{'name':'Alice','attributes':['Person']},{'name':'Aurora','attributes':['Project']},{'name':'Bob','attributes':['Person']}],'relations':[{'node1':'Alice','node2':'Aurora','type':'manages'},{'node1':'Bob','node2':'Aurora','type':'reviews'}]},'wiki_config':{'synthesis_model_id':models['KnowledgeQA'],'max_pages_per_ingest':5,'extraction_granularity':'focused','ingest_max_inflight':1,'ingest_map_parallel':1,'ingest_reduce_parallel':1}})['data']
kbID=kb['id']; state['graph_kb']=kbID
listing=req('GET',f'/knowledge-bases/{kbID}/knowledge?page_size=100')['data']
items=listing if isinstance(listing,list) else listing.get('items',[])
ids=[]
for name,text in [('LJC_Aurora.md','# Aurora project\nAlice manages project Aurora. Bob reviews Aurora archives. Aurora stores 17 archive boxes.\n'),('LJC_Orion.md','# Orion project\nCarol manages project Orion. Dave reviews Orion archives. Orion stores 23 books.\n')]:
 doc=next((x for x in items if x.get('file_name')==name),None)
 if not doc:
  boundary='ljc-wiki-acceptance'
  body=(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{name}"\r\nContent-Type: text/markdown\r\n\r\n'+text+f'\r\n--{boundary}--\r\n').encode()
  with urllib.request.urlopen(urllib.request.Request(base+f'/knowledge-bases/{kbID}/knowledge/file',data=body,headers={'Authorization':'Bearer '+token,'Content-Type':'multipart/form-data; boundary='+boundary},method='POST'),timeout=90) as r:doc=json.load(r)['data']
 ids.append(doc['id']);print('WIKI_DOCUMENT',name,doc['id'],flush=True)
state['graph_documents']=ids;(P/'acceptance-api-state.json').write_text(json.dumps(state,indent=2))
for i in range(90):
 pages=req('GET',f'/knowledgebase/{kbID}/wiki/pages?page_size=100')
 graph=req('GET',f'/knowledgebase/{kbID}/wiki/graph?scope=documents&view=union&knowledge_ids={ids[0]}')
 nodes=graph.get('nodes',[])
 print('WIKI_GRAPH',len(nodes),flush=True)
 if nodes:
  (P/'graph-extraction-initial-results.json').write_text(json.dumps({'kb':kbID,'documents':ids,'pages':pages,'graph':graph},ensure_ascii=False,indent=2))
  break
 time.sleep(2)
else:raise RuntimeError('wiki document graph not ready after 180s')
print('WIKI_GRAPH_OK',flush=True)

contexts=[]
for kid in ids:
 for _ in range(180):
  c=req('GET','/knowledge/'+kid+'/archive-context')['data']
  if c['wiki_pages'] and c['graph']['done']>0 and c['graph_data']['relation']:break
  time.sleep(2)
 else:raise AssertionError('document graph not extracted '+str(c))
 nodes=c['graph_data']['node'];edges=c['graph_data']['relation']
 nodeids={x['id'] for x in nodes}
 assert all(x['source_id'] in nodeids and x['target_id'] in nodeids for x in edges)
 foreign='Orion' if kid==ids[0] else 'Aurora'
 assert all(foreign.lower() not in (x.get('name','')+x.get('title','')).lower() for x in nodes+c['wiki_pages'])
 contexts.append(c)
assert {x['slug'] for x in contexts[0]['wiki_pages']}.isdisjoint({x['slug'] for x in contexts[1]['wiki_pages']})
(P/'graph-scope-results.json').write_text(json.dumps({'kb':kbID,'documents':ids,'contexts':contexts,'scope_isolation':True,'relation_endpoints_valid':True},indent=2))
print('REAL_GRAPH_SCOPE_ENDPOINTS_OK',flush=True)
