#!/usr/bin/env python3
"""Run the bounded model probe using existing app Redis settings, without logging secrets."""
import json
import os
import pathlib
import re
import subprocess
import threading
import urllib.request

def inspect(name):
    return json.loads(subprocess.check_output(['docker','inspect',name]))[0]

app, redis = inspect('WeKnora-app'), inspect('WeKnora-redis')
config = dict(x.split('=',1) for x in app['Config']['Env'] if '=' in x)
shared = set(app['NetworkSettings']['Networks']) & set(redis['NetworkSettings']['Networks'])
if not shared:
    raise SystemExit('app and Redis have no common Docker network')
network = sorted(shared)[0]
env = dict(os.environ)
env['REDIS_ADDR'] = redis['NetworkSettings']['Networks'][network]['IPAddress'] + ':6379'
env['REDIS_PASSWORD'] = config.get('REDIS_PASSWORD','')
env['REDIS_DB'] = config.get('REDIS_DB','0')
peak = 0
stop = threading.Event()
def metrics():
    global peak
    while not stop.is_set():
        try:
            text = urllib.request.urlopen('http://127.0.0.1:8101/metrics',timeout=2).read().decode()
            for number in re.findall(r'^vllm:num_requests_running(?:\{[^}]*\})?\s+([\d.]+)',text,re.M):
                peak = max(peak,int(float(number)))
        except Exception:
            pass
        stop.wait(.5)
thread = threading.Thread(target=metrics,daemon=True)
thread.start()
try:
    result = subprocess.run(['runtime/enrichment-acceptance'],env=env,text=True,capture_output=True,timeout=210)
finally:
    stop.set();thread.join(3)
output = result.stdout + result.stderr + f'vllm_running_peak={peak}\n'
path = pathlib.Path('runtime/enrichment-live-acceptance.log')
path.write_text(output)
print(output,end='')
if result.returncode or peak < 8:
    raise SystemExit(result.returncode or 1)
