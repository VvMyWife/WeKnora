#!/usr/bin/env python3
"""Independent real-API fixture; never touches a formal KB or reprocesses it.

setup creates one clearly named sample KB and imports pre-chunked test text +
the existing 16-page test PDF. chat runs normal/agent QA against that KB only.
Credentials are prompted, never stored in the report or repository.
"""
import argparse,getpass,hashlib,io,json,pathlib,time,uuid,zipfile
import requests

p=argparse.ArgumentParser()
p.add_argument('action',choices=['setup','normal','agent','inspect','verify'])
p.add_argument('--base',default='http://127.0.0.1:18081')
p.add_argument('--email',required=True)
p.add_argument('--state',type=pathlib.Path,required=True)
p.add_argument('--pdf',type=pathlib.Path)
p.add_argument('--schema',type=pathlib.Path)
args=p.parse_args()
s=requests.Session()
login=s.post(args.base+'/api/v1/auth/login',json={'email':args.email,'password':getpass.getpass('Login password: ')},timeout=30)
login.raise_for_status();auth=login.json();token=auth.get('token') or auth.get('data',{}).get('token')
if not token:raise RuntimeError('login returned no token')
s.headers['Authorization']='Bearer '+token
def api(method,path,**kw):
    r=s.request(method,args.base+'/api/v1'+path,timeout=kw.pop('timeout',120),**kw)
    if r.status_code>=400:raise RuntimeError(f'{method} {path}: {r.status_code} {r.text[:500]}')
    return r.json()
def save(state):
    args.state.parent.mkdir(parents=True,exist_ok=True)
    args.state.write_text(json.dumps(state,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    args.state.chmod(0o600)
def sha(b):return hashlib.sha256(b).hexdigest()
if args.action=='setup':
    if args.state.exists():raise RuntimeError('state already exists; do not create another fixture')
    pdf=args.pdf.read_bytes();schema=args.schema.read_bytes()
    bases=api('GET','/knowledge-bases')['data']
    if isinstance(bases,dict):bases=bases.get('items',bases.get('knowledge_bases',[]))
    template=next(k for k in bases if k.get('embedding_model_id') and k.get('summary_model_id'))
    config={k:template[k] for k in ['embedding_model_id','summary_model_id','vector_store_id','storage_backend_id','storage_provider_config'] if template.get(k)}
    config.update(name='档案脚注验收-'+time.strftime('%Y%m%d-%H%M%S'),type='document',description='独立功能验收样例；不参与正式业务',question_generation_config={'enabled':False},wiki_config={'enabled':False},extract_config=None)
    kb=api('POST','/knowledge-bases',json=config)['data']
    state={'knowledge_base_id':kb['id'],'model_id':config['summary_model_id'],'created_at':time.time()};save(state)
    document={'id':'citation-qa-doc','title':'档案脚注物理页验收','relative_path':'samples/citation-qa.md','version':1,'source_sha256':sha(pdf),'parsed_file':'files/citation-qa.md','original_file':'files/citation-qa.pdf'}
    chunks=[];units=[]
    for index,page in enumerate([2,3]):
        text=f'FULLPAGE TEST {page} OF 16'
        units.append({'id':f'unit-{page}','document_id':document['id'],'unit_id':f'unit-{page}','page_index':page-1,'order':index,'block_type':'text','text':text})
        chunks.append({'id':f'chunk-{page}','document_id':document['id'],'order':index,'content':text,'retrieval_text':text,'token_count':16,'metadata':{},'source_anchors':[{'unit_id':f'unit-{page}','page_index':page-1}]})
    def jsonl(rows):return ''.join(json.dumps(r,ensure_ascii=False)+'\n' for r in rows).encode()
    files={'documents.jsonl':jsonl([document]),'chunks.jsonl':jsonl(chunks),'source_units.jsonl':jsonl(units),'files/citation-qa.md':b'FULLPAGE TEST 2 OF 16\n\nFULLPAGE TEST 3 OF 16\n','files/citation-qa.pdf':pdf}
    manifest={'format':'weknora-chunks-v1','schema_sha256':sha(schema),'package_id':str(uuid.uuid4()),'source_system':'archival-citation-acceptance','dataset_id':state['knowledge_base_id'],'generation_id':'physical-pages-v1','producer_commit':'acceptance-only','generation_config':{},'counts':{'documents':1,'chunks':2,'source_units':2},'files':{name:{'sha256':sha(b),'size':len(b)} for name,b in files.items()}}
    out=io.BytesIO()
    with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED) as z:
        z.writestr('manifest.json',json.dumps(manifest).encode())
        for name,b in files.items():z.writestr(name,b)
    result=api('POST',f"/knowledge-bases/{kb['id']}/knowledge/external-chunks/batches",files={'file':('citation-qa.zip',out.getvalue(),'application/zip')})['data']
    state.update(batch_id=result['batch']['id'],record_id=result['documents'][0]['id']);save(state)
    api('POST',f"/knowledge-bases/{kb['id']}/knowledge/external-chunks/batches/{state['batch_id']}/documents/{state['record_id']}/start",json={})
    print(json.dumps(state,ensure_ascii=False),flush=True)
else:
    state=json.loads(args.state.read_text(encoding='utf-8'))
    if args.action=='inspect':
        result=api('GET',f"/knowledge-bases/{state['knowledge_base_id']}/knowledge/external-chunks/batches/{state['batch_id']}")
        print(json.dumps(result,ensure_ascii=False)[:6000])
    elif args.action=='verify':
        checks=[]
        for mode in ['normal','agent']:
            if mode not in state:continue
            turn=state[mode]
            messages=turn['messages']['data']
            if isinstance(messages,dict):messages=messages.get('messages',messages.get('items',[]))
            message=next(m for m in messages if m['role']=='assistant' and m.get('is_completed'))
            path=f"/messages/{turn['session_id']}/{message['id']}/citations"
            manifest=api('GET',path)['data']
            complete=next(e for e in reversed(turn['events']) if e.get('response_type')=='complete')['data']
            assert complete['final_answer']==message['content']
            assert manifest==complete['archival_citations']
            assert manifest['citations'],'no actual citation to verify'
            for ref in manifest['citations']:
                detail=api('GET',path+'/'+ref['id'])['data']
                assert sha(detail['snapshot'].encode())==ref['content_sha256']
                r=s.get(args.base+f"/api/v1/knowledge/{ref['knowledge_id']}/preview",params={'source':'archive','archive_version_id':ref['archive_version_id']},timeout=30)
                assert r.status_code==200 and sha(r.content)==ref['archive_sha256']
                assert requests.get(args.base+'/api/v1'+path+'/'+ref['id'],timeout=10).status_code in [401,403]
            checks.append({'mode':mode,'message_id':message['id'],'evidence_count':len(manifest['citations']),'state':manifest['state'],'pages':[c['location']['pages'] for c in manifest['citations']],'canonical_stream_matches_history':True,'snapshots_and_original_hashes_match':True,'unauthenticated_evidence_denied':True})
        state['verified']=checks;save(state);print(json.dumps(checks,ensure_ascii=False))
    else:
        session=api('POST','/sessions',json={'title':f'档案引用验收-{args.action}'})['data']
        mode='knowledge-chat' if args.action=='normal' else 'agent-chat'
        request={'query':'请仅依据这份测试档案，逐条列出其中的两个 FULLPAGE TEST 标记原文，每条都用知识库脚注注明来源。不要补充其他信息。','knowledge_base_ids':[state['knowledge_base_id']],'summary_model_id':state['model_id'],'agent_enabled':args.action=='agent','disable_title':True}
        if args.action=='agent':request['agent_id']='builtin-smart-reasoning'
        response=s.post(args.base+f"/api/v1/{mode}/{session['id']}",json=request,stream=True,timeout=(15,240))
        if response.status_code>=400:raise RuntimeError(f'{response.status_code} {response.text[:500]}')
        events=[]
        for line in response.iter_lines():
            if not line.startswith(b'data:'):continue
            try:event=json.loads(line[5:].strip())
            except ValueError:continue
            events.append(event)
            if event.get('response_type') in ['error','complete']:print(json.dumps(event,ensure_ascii=False)[:3500],flush=True)
        state[args.action]={'session_id':session['id'],'events':events};save(state)
        messages=api('GET',f"/messages/{session['id']}/load")
        state[args.action]['messages']=messages;save(state)
        print('saved real conversation: '+session['id'],flush=True)
