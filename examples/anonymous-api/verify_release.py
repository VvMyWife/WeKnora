"""Post-deployment acceptance from an approved intranet client; no login or key."""
import argparse
import json
import pathlib
import time
import httpx
from openai import OpenAI

p=argparse.ArgumentParser()
p.add_argument('--base',default='http://10.20.0.15:18182/api/openai/v1')
p.add_argument('--model',default='policy-qa')
p.add_argument('--out',required=True)
p.add_argument('--reference-response-id',help='Optional earlier successful response from this entrypoint')
a=p.parse_args()
h=httpx.Client(timeout=310)
c=OpenAI(base_url=a.base,api_key='not-required',max_retries=0,timeout=310)
report={'checked_at':time.time(),'base':a.base,'checks':{}}
def check(name,v):report['checks'][name]=bool(v);assert v,name
def evidence(response_id,citations):
    r=h.get(a.base+'/responses/'+response_id+'/citations')
    check('manifest_'+response_id,r.status_code==200 and r.json()['citations']==citations)
    for ref in citations:
        d=h.get(a.base.split('/api/openai/v1')[0]+ref['detail_url'])
        check('evidence_'+ref['id'],d.status_code==200 and bool(d.json().get('snapshot')))
        check('no_host_path_'+ref['id'],'/home/ubuntu/' not in d.text and '/data/files/' not in d.text)
        report.setdefault('evidence',[]).append(d.json())
r=h.get(a.base+'/models')
check('no_authorization_header',r.status_code==200 and a.model in [x['id'] for x in r.json()['data']])
q='依据 GB/T 20988，简要概括第三级灾难恢复的数据备份要求，附引用依据。'
start=time.monotonic()
r=c.chat.completions.create(model=a.model,messages=[{'role':'user','content':q}],max_tokens=1024)
content=r.choices[0].message.content
refs=r.model_extra.get('citations',[])
check('sdk_sync',bool(content) and isinstance(refs,list))
state=r.model_extra.get('citation_status',{})
check('sync_citation_status',('[1]' in content) if refs else (state.get('state')=='needs_review' and state.get('unbound_count',0)>0))
check('actual_usage',r.usage is not None and r.usage.total_tokens>0)
report['sync']={'id':r.id,'seconds':time.monotonic()-start,'content':content,'citations':refs,'citation_status':r.model_extra.get('citation_status'),'usage':r.usage.model_dump() if r.usage else None}
evidence(r.id,refs)
stream=c.chat.completions.create(model=a.model,messages=[{'role':'user','content':'依据GB/T 20988，第三级灾难恢复要求多久做一次完全备份？简短回答并附引用。'}],max_tokens=512,stream=True)
draft='';final=None;streamrefs=None;finish=None;streamid=None
for chunk in stream:
    streamid=chunk.id
    if chunk.choices:
        draft+=chunk.choices[0].delta.content or ''
        finish=chunk.choices[0].finish_reason or finish
    extra=chunk.model_extra or {}
    if 'error' in extra:raise RuntimeError(extra['error'])
    if 'final_content' in extra:
        final=extra['final_content'];streamrefs=extra.get('citations',[])
check('sdk_stream_final',bool(final) and isinstance(streamrefs,list) and finish in ['stop','length'] and '<kb ' not in final)
report['stream']={'id':streamid,'draft':draft,'final_content':final,'citations':streamrefs,'finish_reason':finish}
evidence(streamid,streamrefs)
if a.reference_response_id:
    prior=h.get(a.base+'/responses/'+a.reference_response_id+'/citations')
    check('prior_response_persisted',prior.status_code==200)
    report['reference_response']=prior.json()
    evidence(a.reference_response_id,prior.json()['citations'])
check('at_least_one_real_evidence',bool(report.get('evidence')))
history=[{'role':'user','content':'本次接口验收的临时代号是青杉486。'},
         {'role':'assistant','content':'记住了，本次接口验收的临时代号是青杉486。'},
         {'role':'user','content':'我刚才指定的接口验收临时代号是什么？只回答代号。'}]
r=c.chat.completions.create(model=a.model,messages=history,max_tokens=128)
check('multi_turn_history','青杉486' in r.choices[0].message.content)
report['multi']={'id':r.id,'content':r.choices[0].message.content}
check('web_management_still_requires_login',h.get(a.base.split('/api/openai/v1')[0]+'/api/v1/knowledge-bases').status_code in [401,403])
check('private_message_id_denied',h.get(a.base+'/responses/private-web-message/citations').status_code==404)
check('strict_400',h.post(a.base+'/chat/completions',json={'model':a.model,'messages':[{'role':'user','content':'test'}],'tools':[]}).status_code==400)
check('unknown_alias_404',h.post(a.base+'/chat/completions',json={'model':'not-published','messages':[{'role':'user','content':'test'}]}).status_code==404)
pathlib.Path(a.out).write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps({'passed':len(report['checks']),'sync_seconds':report['sync']['seconds'],'citations':len(refs),'out':a.out},ensure_ascii=False))
