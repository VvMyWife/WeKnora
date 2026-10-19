"""Start an isolated custom UI on port 18184, using the live backend read APIs."""
import argparse
import http.client
import json
from pathlib import Path
import socket
import subprocess

p = argparse.ArgumentParser()
p.add_argument('directory', type=Path)
p.add_argument('image')
args = p.parse_args()
before = json.loads((args.directory / 'frontend-before.json').read_text())[0]
config = dict(before['Config'])
config['Image'] = args.image
host = dict(before['HostConfig'])
host['PortBindings'] = {'80/tcp': [{'HostIp': '0.0.0.0', 'HostPort': '18184'}]}
host['RestartPolicy'] = {'Name': 'no', 'MaximumRetryCount': 0}
config['HostConfig'] = host

class Connection(http.client.HTTPConnection):
    def connect(self):
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect('/var/run/docker.sock')

conn = Connection('localhost')
conn.request('POST', '/containers/create?name=WeKnora-frontend-custom-document-graph-qa', json.dumps(config), {'Content-Type': 'application/json'})
response = conn.getresponse()
body = response.read().decode()
if response.status != 201:
    raise SystemExit('preview create failed: ' + body)
subprocess.run(['docker', 'start', 'WeKnora-frontend-custom-document-graph-qa'], check=True)
print('Preview: http://10.20.0.15:18184 (no backend restart)')
