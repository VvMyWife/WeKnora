"""Run with the persistent examples/anonymous-api/.venv interpreter."""
import argparse,json,os
from openai import OpenAI
import httpx
p=argparse.ArgumentParser();p.add_argument('--base',default=os.getenv('WEKNORA_API_BASE','http://10.20.0.15:18182/api/openai/v1'));p.add_argument('--model',default='policy-qa');p.add_argument('--question',default='第三级灾难恢复的数据备份要求是什么？');p.add_argument('--stream',action='store_true');p.add_argument('--max-tokens',type=int,default=1024);args=p.parse_args()
c=OpenAI(base_url=args.base,api_key='not-required',timeout=310,max_retries=0)
print('models:',[m.id for m in c.models.list().data])
if args.stream:
 response_id=None;citations=[];final=None
 for part in c.chat.completions.create(model=args.model,messages=[{'role':'user','content':args.question}],stream=True,max_tokens=args.max_tokens):
  response_id=part.id
  if part.choices and part.choices[0].delta.content:print(part.choices[0].delta.content,end='',flush=True)
  extra=part.model_extra or {}
  if 'final_content' in extra:final=extra['final_content'];citations=extra.get('citations',[]);print('\nCitation state:',extra.get('citation_status'))
 print('\nFinal content:',final)
else:
 response=c.chat.completions.create(model=args.model,messages=[{'role':'user','content':args.question}],max_tokens=args.max_tokens)
 print(response.choices[0].message.content);print('usage:',response.usage)
 response_id=response.id;citations=(response.model_extra or {}).get('citations',[]);print('Citation state:',(response.model_extra or {}).get('citation_status'))
print('citations:',json.dumps(citations,ensure_ascii=False,indent=2))
with httpx.Client(timeout=30) as http:
 manifest=http.get(args.base+'/responses/'+response_id+'/citations');manifest.raise_for_status();print('manifest:',manifest.json())
 if citations:
  detail=http.get(args.base+'/responses/'+response_id+'/citations/'+citations[0]['id']);detail.raise_for_status();print('first quote:',detail.json()['snapshot'])
# Multi-turn: send the supplied user/assistant history, never a server conversation ID.
# c.chat.completions.create(model=args.model,messages=[{'role':'user','content':'问题一'}, {'role':'assistant','content':'回答一'}, {'role':'user','content':'追问'}])
