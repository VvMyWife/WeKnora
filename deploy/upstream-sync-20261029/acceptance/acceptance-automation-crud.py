import pathlib,json
P=pathlib.Path('/home/ubuntu/weknora-fork-migration/20261029-upstream')
exec((P/'acceptance-api-setup.py').read_text().split("state=json.loads")[0])
trigger={'type':'once','at':'2030-01-01T10:00:00+08:00','timezone':'Asia/Shanghai'}
check=req('POST','/automations/validate-trigger',{'trigger':trigger})
name='LJC_AUTOMATION_ACCEPTANCE_20261021'
rows=req('GET','/automations').get('data',[])
if isinstance(rows,dict):rows=rows.get('items',[])
a=next((r for r in rows if r['name']==name),None)
if not a:a=req('POST','/automations',{'name':name,'instruction':'Isolated acceptance fixture. Do not contact external services.','trigger':trigger,'access_level':'read_only','session_mode':'per_run','approval_timeout_minutes':60})['data']
id=a['id']
req('POST',f'/automations/{id}/pause',{})
a=req('GET',f'/automations/{id}')['data']
assert a['enabled'] is False,a
req('PUT',f'/automations/{id}',{'description':'CRUD and paused-state acceptance; never executed'})
a=req('GET',f'/automations/{id}')['data']
assert not a['enabled'] and a['description'].startswith('CRUD'),a
(P/'automation-crud-results.json').write_text(json.dumps({'id':id,'created':True,'updated':True,'paused':True,'never_executed':True,'trigger_validation':check,'cleanup_pending':True},indent=2))
print('AUTOMATION_CREATED_PAUSED_UPDATED_OK',id)
