import runpy,json,pathlib,urllib.request,time
ns=runpy.run_path('/home/ubuntu/weknora-fork-migration/20261029-upstream/acceptance-api-setup.py')
req,P,token=ns['req'],ns['P'],ns['token']
state=json.loads((P/'acceptance-api-state.json').read_text())
kb=state['kb']
listing=req('GET',f'/knowledge-bases/{kb}/knowledge?page=1&page_size=100')['data']
items=listing if isinstance(listing,list) else listing.get('items',[])
name='LJC_ACCEPTANCE_INDEXED_20261020.md'
doc=next((x for x in items if x.get('file_name')==name or x.get('title')==name),None)
if not doc:
 boundary='ljc-acceptance-file-boundary'
 content='# Integration acceptance indexed retrieval\n\nLJCVERIFY20261020 is an isolated verification marker.\n\nProject Aurora is managed by Alice. Bob reviews the archive.\n\n| Item | Quantity |\n|---|---|\n| Archive boxes | 17 |\n| Scan pages | 42 |\n'
 body=(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{name}"\r\nContent-Type: text/markdown\r\n\r\n'+content+f'\r\n--{boundary}--\r\n').encode()
 request=urllib.request.Request(ns['BASE']+f'/knowledge-bases/{kb}/knowledge/file',data=body,headers={'Authorization':'Bearer '+token,'Content-Type':'multipart/form-data; boundary='+boundary},method='POST')
 with urllib.request.urlopen(request,timeout=90) as r:doc=json.load(r)['data']
state['document']=doc['id']
(P/'acceptance-api-state.json').write_text(json.dumps(state,indent=2))
print('DOCUMENT',doc['id'],flush=True)
for i in range(180):
 doc=req('GET',f'/knowledge/{doc["id"]}')['data']
 status=doc.get('parse_status',doc.get('status'))
 print('STATUS',status,flush=True)
 if status in ('completed','failed'):break
 time.sleep(2)
assert status=='completed',str(doc)
checks={}
for mode,flags in [('keyword',{'disable_vector_match':True}),('vector',{'disable_keywords_match':True}),('hybrid',{}),('rerank',{'rerank':{'enabled':True,'model_id':state['models']['Rerank'],'top_k':5,'threshold':-100}})]:
 result=req('POST',f'/knowledge-bases/{kb}/hybrid-search',{'query_text':'LJCVERIFY20261020','match_count':5,**flags})
 hits=result.get('data',[])
 assert hits,mode+' has no hits'
 assert any('LJCVERIFY20261020' in h.get('content','') for h in hits),mode+' missing fixture'
 checks[mode]={'hits':len(hits),'first_chunk':hits[0].get('id')}
 print('SEARCH_OK',mode,len(hits),flush=True)
(P/'realfile-results.json').write_text(json.dumps({'document':doc['id'],'status':status,'search':checks},indent=2))
