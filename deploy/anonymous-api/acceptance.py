#!/usr/bin/env python3
"""SDK acceptance against the isolated fixture only; writes a local evidence report."""
import argparse
import json
import pathlib
import time
import httpx
from openai import OpenAI

p = argparse.ArgumentParser()
p.add_argument('--base', default='http://127.0.0.1:28481/api/openai/v1')
p.add_argument('--out', required=True)
a = p.parse_args()
assert ':28481/' in a.base or ':28482/' in a.base, 'acceptance fixture ports only'
c = OpenAI(base_url=a.base, api_key='not-required', max_retries=0, timeout=310)
http = httpx.Client(timeout=310)
report = {'base': a.base, 'started_at': time.time(), 'checks': {}}
def check(name, value):
    report['checks'][name] = bool(value)
    assert value, name
def chat(messages, **kw):
    start = time.monotonic()
    r = c.chat.completions.create(model='fixture-qa', messages=messages, max_tokens=1024, **kw)
    return r, time.monotonic()-start

for attempt in range(60):
    try:
        if http.get(a.base+'/models').status_code==200:break
    except httpx.RequestError:pass
    time.sleep(1)
else:raise RuntimeError('acceptance API not ready')

check('models', 'fixture-qa' in [x.id for x in c.models.list().data])
q = '根据独立API验收资料，检查周期和备份保留份数分别是多少？'
r, elapsed = chat([{'role':'user','content':q}])
text = r.choices[0].message.content
citations = r.model_extra.get('citations', [])
check('real_answer', '27' in text and '3' in text)
check('inline_footnotes', len(citations)>0 and '[1]' in text)
report['sync'] = {'id':r.id, 'seconds':elapsed, 'content':text,
                  'citations':citations, 'usage':r.usage.model_dump() if r.usage else None}
check('usage', r.usage is not None and r.usage.total_tokens>0)
base = a.base.split('/api/openai/v1')[0]
cl = http.get(f'{a.base}/responses/{r.id}/citations')
check('no_auth_citation_list', cl.status_code==200 and cl.json()['citations']==citations)
details=[]
for citation in citations:
    d = http.get(base+citation['detail_url'])
    check('snapshot_'+citation['id'], d.status_code==200 and bool(d.json().get('snapshot')))
    details.append(d.json())
report['details']=details
check('private_web_id_denied', http.get(a.base+'/responses/private-web-message/citations').status_code==404)
check('management_not_anonymous', http.get(base+'/api/v1/knowledge-bases').status_code in [401,403,404])
bad=http.post(a.base+'/chat/completions',json={'model':'fixture-qa','messages':[{'role':'user','content':'q'}],'tools':[]})
check('unsupported_parameter_400',bad.status_code==400)
stream,elapsed=chat([{'role':'user','content':'根据独立API验收资料，备份应分别放在哪里？'}],stream=True)
draft=''; final=None; manifest=None; finish=None
for chunk in stream:
    if chunk.choices:
        draft += chunk.choices[0].delta.content or ''
        finish=chunk.choices[0].finish_reason or finish
    extra=chunk.model_extra or {}
    if 'error' in extra: raise RuntimeError(extra['error'])
    if 'final_content' in extra:
        final=extra['final_content'];manifest=extra.get('citations',[])
check('sdk_stream_final', bool(final) and finish in ['stop','length'] and '<kb ' not in final)
check('stream_evidence', bool(manifest) and '甲' in final and '丙' in final)
report['stream']={'draft':draft,'final_content':final,'citations':manifest,'finish_reason':finish}
# An invented prior user detail is intentionally absent from the sample library.
# This verifies the supplied history reaches the current turn rather than rerunning it.
history=[{'role':'user','content':'我给这次验收起的代号是蓝松729，请记住。'},
         {'role':'assistant','content':'好的，这次验收的代号是蓝松729。'},
         {'role':'user','content':'我刚才给这次验收起的代号是什么？'}]
multi,elapsed=chat(history)
check('multi_turn_history', '蓝松729' in multi.choices[0].message.content)
report['multi']={'content':multi.choices[0].message.content,'seconds':elapsed,'id':multi.id}
report['finished_at']=time.time()
pathlib.Path(a.out).write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps({'passed':len(report['checks']), 'sync_seconds':report['sync']['seconds'],
                  'citations':len(citations),'report':a.out},ensure_ascii=False))
