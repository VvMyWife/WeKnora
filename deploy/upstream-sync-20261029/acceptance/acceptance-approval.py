import json,runpy,urllib.request,urllib.error,time
from pathlib import Path
P=Path(__file__).parent;n=runpy.run_path(str(P/'acceptance-api-setup.py'));req=n['req']
u=json.loads((P/'acceptance-second-user.json').read_text());login=req('POST','/auth/login',{'email':u['email'],'password':u['password']});other=login['token']
def foreign(method,path,data=None):
 r=urllib.request.Request(n['BASE']+path,data=json.dumps(data).encode() if data is not None else None,headers={'Content-Type':'application/json','Authorization':'Bearer '+other},method=method)
 try:
  with urllib.request.urlopen(r,timeout=30) as h:return json.load(h)
 except urllib.error.HTTPError as e:
  print('HTTP_FAILURE',e.code,e.read().decode()[:600]);raise
org=req('POST','/organizations',{'name':'LJC approval proof '+str(time.time_ns()),'description':'Isolated acceptance only','require_approval':True})['data']
oid=org['id'];req('PUT','/organizations/'+oid,{'require_approval':True});code=req('POST','/organizations/'+oid+'/invite-code',{})['data']['invite_code']
join=foreign('POST','/organizations/join-request',{'invite_code':code,'role':'editor','message':'Synthetic approval fixture'})['data'];jid=join['id']
try:
 foreign('PUT',f'/organizations/{oid}/join-requests/{jid}/review',{'approved':True})
 raise AssertionError('applicant could approve own request')
except urllib.error.HTTPError as e:assert e.code==403
req('PUT',f'/organizations/{oid}/join-requests/{jid}/review',{'approved':True,'role':'viewer','message':'Acceptance approved as viewer'})
members=req('GET',f'/organizations/{oid}/members')['data']
rows=members if isinstance(members,list) else members.get('items',members.get('members',[]))
assert any(x.get('tenant_id')==login['active_tenant']['id'] and x.get('role')=='viewer' for x in rows),str(rows)
(P/'approval-results.json').write_text(json.dumps({'organization':oid,'applicant_self_review_denied':True,'admin_approval_succeeded':True,'assigned_role_viewer_verified':True},indent=2))
print('APPROVAL_ROLE_BOUNDARY_OK')
