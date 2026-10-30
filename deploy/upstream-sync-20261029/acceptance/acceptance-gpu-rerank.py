import runpy,json,pathlib,urllib.request
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
n=runpy.run_path(str(P/'acceptance-api-setup.py'));r=n['req'];s=json.loads((P/'acceptance-api-state.json').read_text())
models=r('GET','/models')['data'];model=next(x for x in models if x['id']==s['models']['Rerank'])
params=model['parameters'];params['base_url']='http://10.20.0.15:18083';params['api_key']='acceptance-local'
r('PUT',f'/models/{model["id"]}',{'name':model['name'],'type':model['type'],'source':model['source'],'parameters':params})
result=r('POST',f'/knowledge-bases/{s["kb"]}/hybrid-search',{'query_text':'Aurora archive boxes','match_count':5,'rerank':{'enabled':True,'model_id':model['id'],'top_k':3,'threshold':-100}})
print('RERANK_FIELDS',list(result),flush=True)
print('RERANK_METADATA',{k:v for k,v in result.items() if k!='data'},flush=True)
(P/'gpu-rerank-results.json').write_text(json.dumps(result,ensure_ascii=False,indent=2))
assert result.get('data'),'reranked results missing'
assert result['meta']['rerank']['applied'] is True and result['meta']['rerank']['outcome']=='ok',result.get('meta')
print('GPU_RERANK_REQUEST_OK',flush=True)
