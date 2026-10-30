import runpy,json,urllib.request,urllib.error,secrets,pathlib
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream');ns=runpy.run_path(str(P/'acceptance-api-setup.py'));req=ns['req'];base=ns['BASE'];state=json.loads((P/'acceptance-api-state.json').read_text())
path=P/'acceptance-second-user.json'
if path.exists():user=json.loads(path.read_text())
else:
 user={'username':'ljc_isolation','email':'ljc-isolation@example.test','password':secrets.token_urlsafe(24)}
 req('POST','/auth/register',user);path.write_text(json.dumps(user));path.chmod(0o600)
login=req('POST','/auth/login',{'email':user['email'],'password':user['password']})
token=login['token'];assert login['active_tenant']['id']!=state['tenant']
checks={}
for name,path in [('kb',f'/knowledge-bases/{state["kb"]}'),('document',f'/knowledge/{state["document"]}'),('admin','/system/admin/list')]:
 try:
  urllib.request.urlopen(urllib.request.Request(base+path,headers={'Authorization':'Bearer '+token}),timeout=20)
  raise RuntimeError('isolation failed '+name)
 except urllib.error.HTTPError as e:
  assert e.code in (403,404),name+' '+str(e.code);checks[name]=e.code
print('WORKSPACE_ISOLATION_OK',checks,flush=True)
models=req('GET','/models')['data']
for model in models:
 p=model.get('parameters',{});assert not p.get('api_key') and not p.get('app_secret'),'model returned credential'
checks['model_secrets_redacted']=True
(P/'permissions-results.json').write_text(json.dumps(checks,indent=2))
print('MODEL_SECRET_REDACTION_OK',flush=True)
