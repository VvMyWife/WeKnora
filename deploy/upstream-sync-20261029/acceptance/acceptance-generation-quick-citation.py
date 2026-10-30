
import runpy,json,urllib.request
from pathlib import Path
P=Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
n=runpy.run_path(str(P/'acceptance-api-setup.py'));s=json.loads((P/'acceptance-api-state.json').read_text())
r=n['req']('POST','/sessions',{'title':'LJC sync generation-time citation acceptance'})
session=r.get('data',r); sid=session['id']
a=n['req']('POST','/agents/builtin-quick-answer/copy',{})['data']
c=a['config'];c['model_id']=s['models']['KnowledgeQA'];c['max_completion_tokens']=8192;c['thinking']=False
n['req']('PUT','/agents/'+a['id'],{'name':'LJC quick citation regression','description':'Isolated model configuration','config':c})
payload={'query':'LJCVERIFY20261020 的 Archive boxes 是多少？请依据检索资料用一句话回答并在事实旁引用来源。','knowledge_base_ids':[s['kb']],'knowledge_ids':[s['document']],'summary_model_id':s['models']['KnowledgeQA'],'agent_enabled':False,'agent_id':a['id'],'disable_title':True,'channel':'web'}
request=urllib.request.Request(n['BASE']+'/knowledge-chat/'+sid,data=json.dumps(payload).encode(),headers={'Authorization':'Bearer '+n['token'],'Content-Type':'application/json'},method='POST')
with urllib.request.urlopen(request,timeout=180) as response: body=response.read().decode()
(P/'generation-time-quick-citation.sse').write_text(body)
data=[]
for line in body.splitlines():
 if line.startswith('data:'):
  try:data.append(json.loads(line[5:].strip()))
  except ValueError:pass
(P/'generation-time-quick-citation-events.json').write_text(json.dumps(data,ensure_ascii=False,indent=2))
print('QA_STREAM_COMPLETE bytes='+str(len(body)),flush=True)
final=next((x.get('data',{}) for x in reversed(data) if x.get('response_type')=='complete'),{})
manifest=final.get('archival_citations',{})
assert '17' in final.get('final_answer',''), 'final answer missing expected fact'
assert len(manifest.get('citations',[]))>0, 'final answer has no bound citations'
assert manifest.get('repair_attempted') is False
result={'session_id':sid,'fact_17':True,'bound_citation_observed':True,'citation_count':len(manifest['citations']),'repair_attempted':manifest['repair_attempted'],'raw_saved_privately':True}
(P/'generation-time-quick-citation-result.json').write_text(json.dumps(result,indent=2))
print('GENERATION_TIME_CITATION_OK',flush=True)
