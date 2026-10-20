#!/usr/bin/env python3
"""Isolated OCR fault fixture: healthy preflight followed by deliberate HTTP 500."""
import json,subprocess
from http.server import BaseHTTPRequestHandler,ThreadingHTTPServer
gateway=subprocess.check_output(['docker','network','inspect','-f','{{(index .IPAM.Config 0).Gateway}}','ljc-upstream-validate'],text=True).strip()
class Handler(BaseHTTPRequestHandler):
 def do_GET(self):
  body=json.dumps({'errorCode':0,'errorMsg':'Healthy','status':'ok'}).encode()
  self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)
 def do_POST(self):
  length=int(self.headers.get('Content-Length','0'))
  if length>8*1024*1024:self.send_error(413);return
  self.rfile.read(length)
  body=b'{"errorCode":500,"errorMsg":"Intentional acceptance fault"}'
  self.send_response(500);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)
if __name__=='__main__':
 ThreadingHTTPServer((gateway,18084),Handler).serve_forever()
