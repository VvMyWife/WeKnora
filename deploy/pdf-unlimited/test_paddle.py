#!/usr/bin/env python3
import base64,json,time,urllib.request
from pathlib import Path
root=Path(__file__).resolve().parents[2]
report=[]
for name,count in [('single.pdf',1),('sixteen.pdf',16),('long24.pdf',24),('blank-page.pdf',3),('scan.pdf',1)]:
 started=time.monotonic()
 data={'file':base64.b64encode((root/'runtime/pdf-fixtures'/name).read_bytes()).decode(),'fileType':0,'useDocOrientationClassify':False,'useDocUnwarping':False,'useSealRecognition':False,'useChartRecognition':False,'restructurePages':True,'mergeTables':True,'relevelTitles':True}
 req=urllib.request.Request('http://127.0.0.1:9960/layout-parsing',data=json.dumps(data).encode(),headers={'Content-Type':'application/json'})
 with urllib.request.urlopen(req) as response:body=json.load(response)
 (root/'runtime'/('paddle-'+name+'.json')).write_text(json.dumps(body,ensure_ascii=False))
 result=body['result'];pages=result['layoutParsingResults'];info=result['dataInfo']
 assert info['numPages']==count and info['processedPages']==count and len(pages)==count,(name,info,len(pages))
 assert sorted(p['prunedResult']['page_index'] for p in pages)==list(range(count))
 text='\n'.join(p['markdown']['text'] for p in pages)
 assert f'ZEBRA{count:02d}PAGE{count:02d}' in text.replace(' ',''),(name,'last-page marker not recognized')
 row={'file':name,'pages':count,'seconds':round(time.monotonic()-started,2),'last_page_marker':True};report.append(row)
 print(json.dumps(row),flush=True)
 (root/'runtime/paddle-acceptance.json').write_text(json.dumps(report,indent=2))
