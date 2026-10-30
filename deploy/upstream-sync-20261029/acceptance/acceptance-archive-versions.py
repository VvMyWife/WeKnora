import pathlib,json,hashlib,urllib.request
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
source=(P/'acceptance-archive-external.py').read_text().split('results={}')[0]

exec(source)
before=req('GET',f'/chunks/{kid}?page_size=100')
one=pdf('LJC archive original revision one. Archive boxes 17, pages 42.')
two=pdf('LJC archive original revision two. Archive boxes 17, pages 42. Version tracking verified.')
name='LJC_ACCEPTANCE_INDEXED_20261020.pdf'
info=req('GET',f'/knowledge/{kid}/archive-source')['data']
active=info.get('active')
if not active or active['sha256']!=hashlib.sha256(two).hexdigest():
 b=req('POST',f'/knowledge-bases/{kb}/knowledge/archive-mapping/batches',{'parsed_root':'','files':[{'relative_path':name,'file_size':len(two)}]})['data']
 item=b['items'][0]
 v=multipart(f'/knowledge-bases/{kb}/knowledge/archive-mapping/batches/{b["batch"]["id"]}/items/{item["id"]}/file',name,two,{'knowledge_id':kid,'replace':'true'})['data']
info=req('GET',f'/knowledge/{kid}/archive-source')['data']
assert len(info['versions'])>=2,info
old=next(v for v in info['versions'] if v['sha256']==hashlib.sha256(one).hexdigest())
request=urllib.request.Request(ns['BASE']+f'/knowledge/{kid}/download?source=archive&archive_version_id={old["id"]}',headers={'Authorization':'Bearer '+token})
with urllib.request.urlopen(request,timeout=30) as r:assert r.read()==one
assert before==req('GET',f'/chunks/{kid}?page_size=100'),'mapping version changed parsed chunks'
(P/'archive-version-results.json').write_text(json.dumps({'versions':len(info['versions']),'historical_download_verified':True,'parsed_chunks_unchanged':True},indent=2))
print('ARCHIVE_VERSION_HISTORY_OK',len(info['versions']))
