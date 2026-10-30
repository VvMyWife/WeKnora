import runpy,json,urllib.request,hashlib,zipfile,io,time,pathlib
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
state=json.loads((P/'acceptance-api-state.json').read_text())
ns=runpy.run_path(str(P/'acceptance-api-setup.py'))
req,token=ns['req'],ns['token']; kb=state['kb'];kid=state['document']
def multipart(path,name,data,fields={}):
 boundary='ljc-multipart-acceptance'
 body=b''
 for k,v in fields.items():body+=f'--{boundary}\r\nContent-Disposition: form-data; name="{k}"\r\n\r\n{v}\r\n'.encode()
 body+=f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{name}"\r\nContent-Type: application/octet-stream\r\n\r\n'.encode()+data+f'\r\n--{boundary}--\r\n'.encode()
 request=urllib.request.Request(ns['BASE']+path,data=body,headers={'Authorization':'Bearer '+token,'Content-Type':'multipart/form-data; boundary='+boundary},method='POST')
 with urllib.request.urlopen(request,timeout=90) as r:return json.load(r)
def pdf(text):
 stream=f'BT /F1 12 Tf 50 750 Td ({text}) Tj ET'.encode()
 objects=[b'<< /Type /Catalog /Pages 2 0 R >>',b'<< /Type /Pages /Kids [3 0 R] /Count 1 >>',b'<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>',b'<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>',b'<< /Length '+str(len(stream)).encode()+b' >>\nstream\n'+stream+b'\nendstream']
 out=b'%PDF-1.4\n';offsets=[0]
 for i,o in enumerate(objects,1):offsets.append(len(out));out+=str(i).encode()+b' 0 obj\n'+o+b'\nendobj\n'
 xref=len(out);out+=b'xref\n0 6\n0000000000 65535 f \n'+b''.join(f'{o:010} 00000 n \n'.encode() for o in offsets[1:])+b'trailer << /Size 6 /Root 1 0 R >>\nstartxref\n'+str(xref).encode()+b'\n%%EOF\n'
 return out
results={}
original=pdf('LJC archive original revision one. Archive boxes 17, pages 42.')
name='LJC_ACCEPTANCE_INDEXED_20261020.pdf'
batch=req('POST',f'/knowledge-bases/{kb}/knowledge/archive-mapping/batches',{'parsed_root':'','files':[{'relative_path':name,'file_size':len(original)}]})['data']
item=batch['items'][0]
assert item['knowledge_id']==kid,(item,kid)
uploaded=multipart(f'/knowledge-bases/{kb}/knowledge/archive-mapping/batches/{batch["batch"]["id"]}/items/{item["id"]}/file',name,original,{'knowledge_id':kid,'replace':'true'})['data']
for route in ('preview','download'):
 request=urllib.request.Request(ns['BASE']+f'/knowledge/{kid}/{route}?source=archive',headers={'Authorization':'Bearer '+token})
 with urllib.request.urlopen(request,timeout=30) as r:
  data=r.read()
  assert hashlib.sha256(data).digest()==hashlib.sha256(original).digest(),route
  assert 'no-store' in r.headers.get('Cache-Control',''),r.headers
print('ARCHIVE_OK preview download hash no-store',flush=True)
results['archive']={'knowledge_id':kid,'version_id':uploaded['id'],'bytes':len(original)}
md=b'# External acceptance\nLJCEXTERNAL20261020 archive source has 23 books.\n'
files={'files/parsed.md':md}
def jl(x):return (json.dumps(x)+'\n').encode()
files['documents.jsonl']=jl({'id':'acceptance-doc','title':'LJC external acceptance','relative_path':'external/acceptance.md','version':1,'source_sha256':hashlib.sha256(md).hexdigest(),'parsed_file':'files/parsed.md'})
files['chunks.jsonl']=jl({'id':'acceptance-chunk','document_id':'acceptance-doc','order':0,'content':md.decode(),'retrieval_text':md.decode(),'token_count':20,'source_anchors':[{'unit_id':'u1'}]})
files['source_units.jsonl']=jl({'id':'acceptance-doc:u1','document_id':'acceptance-doc','unit_id':'u1','page_index':0,'order':0,'block_type':'text','text':md.decode()})
manifest={'format':'weknora-chunks-v1','schema_sha256':hashlib.sha256(pathlib.Path('/home/ubuntu/RAG_Weknora-upstream-20261029/internal/externalchunks/manifest.schema.json').read_bytes()).hexdigest(),'package_id':'ljc-acceptance-20261020','source_system':'acceptance','dataset_id':'isolated','generation_id':'g1','producer_commit':'a0041910','generation_config':{},'counts':{'documents':1,'chunks':1,'source_units':1},'files':{n:{'sha256':hashlib.sha256(v).hexdigest(),'size':len(v)} for n,v in files.items()}}
files['manifest.json']=json.dumps(manifest).encode()
data=io.BytesIO()
with zipfile.ZipFile(data,'w') as z:
 for n,v in files.items():z.writestr(n,v)
package=data.getvalue()
path=f'/knowledge-bases/{kb}/knowledge/external-chunks/batches'
a=multipart(path,'acceptance.zip',package)['data']; b=multipart(path,'acceptance.zip',package)['data']
assert a['batch']['id']==b['batch']['id'],'duplicate preflight generated another batch'
batchid=a['batch']['id']; record=a['documents'][0]['id']
if a['documents'][0].get('status') not in ('completed','processing','finalizing'):
 req('POST',f'{path}/{batchid}/documents/{record}/start',{})
for _ in range(180):
 details=req('GET',f'{path}/{batchid}')['data']
 rows=details.get('documents',[])
 status=rows[0]['status']
 print('EXTERNAL_STATUS',status,flush=True)
 if status in ('completed','failed'):break
 time.sleep(2)
assert status=='completed',details
hits=req('POST',f'/knowledge-bases/{kb}/hybrid-search',{'query_text':'LJCEXTERNAL20261020','match_count':5})['data']
assert any('LJCEXTERNAL20261020' in x.get('content','') for x in hits),hits
results['external']={'batch_id':batchid,'record_id':record,'status':status,'duplicate_preflight_same_batch':True}
(P/'archive-external-results.json').write_text(json.dumps(results,indent=2))
print('EXTERNAL_OK import duplicate-preflight retrieval',flush=True)
