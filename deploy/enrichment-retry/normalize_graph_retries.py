#!/usr/bin/env python3
"""Pad existing graph failures to 30m; never alter paused parked tasks."""
import argparse
import collections
import json
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--apply', action='store_true')
args = parser.parse_args()
env = dict(x.split('=', 1) for x in json.loads(subprocess.check_output(['docker', 'inspect', 'WeKnora-app']))[0]['Config']['Env'] if '=' in x)
base = ['docker', 'exec', '-e', 'REDISCLI_AUTH=' + env.get('REDIS_PASSWORD', ''), 'WeKnora-redis', 'redis-cli', '-n', env.get('REDIS_DB', '0'), '--raw']

def redis(*commands):
    return subprocess.check_output(base + list(commands), timeout=30)

def fields(data):
    i = 0
    def varint():
        nonlocal i
        n = shift = 0
        while True:
            b = data[i]; i += 1; n |= (b & 127) << shift
            if b < 128: return n
            shift += 7
    out = {}
    while i < len(data):
        key = varint(); number, wire = key >> 3, key & 7
        if wire == 0: out[number] = varint()
        elif wire == 2:
            size = varint(); out[number] = data[i:i+size]; i += size
        elif wire == 1: i += 8
        elif wire == 5: i += 4
        else: raise ValueError('unsupported protobuf wire type')
    return out

# Read field numbers from the checked-in Asynq protobuf definition, rather than
# guessing offsets. Refuse to mutate when its schema has changed.
from pathlib import Path
schema = Path('third_party/asynq/internal/proto/asynq.pb.go').read_text()
import re
def field_number(name):
    return int(re.search(r'\b' + name + r'\s+[^\n]*protobuf:"[^,]+,(\d+),', schema).group(1))
error_field, failed_field = field_number('ErrorMsg'), field_number('LastFailedAt')
stats = collections.Counter()
ids = redis('ZRANGE', 'asynq:{graph}:retry', '0', '-1').decode().splitlines()
script = "if redis.call('HGET',KEYS[1],'state')~='retry' then return 0 end;local s=redis.call('ZSCORE',KEYS[2],ARGV[1]);if s and tonumber(s)<tonumber(ARGV[2]) then redis.call('ZADD',KEYS[2],ARGV[2],ARGV[1]);return 1 end;return 0"
read = "local out={};for _,id in ipairs(ARGV) do out[id]=redis.call('HGET',KEYS[1]..id,'msg') or '' end;return cjson.encode(out)"
for offset in range(0,len(ids),500):
    batch = json.loads(redis('EVAL',read,'1','asynq:{graph}:t:',*ids[offset:offset+500]).decode('latin1'))
    for task_id, value in batch.items():
        if not value: continue
        data = fields(value.encode('latin1'))
        error = data.get(error_field, b'').decode(errors='replace')
        if 'processing is paused' in error:
            stats['paused_untouched'] += 1; continue
        failed_at = data.get(failed_field, 0)
        if not failed_at: continue
        due = failed_at + 1800
        if due <= time.time(): continue
        stats['needs_padding'] += 1
        if args.apply:
            stats['updated'] += int(redis('EVAL', script, '2', 'asynq:{graph}:t:' + task_id, 'asynq:{graph}:retry', task_id, str(due)))
print(json.dumps({'apply': args.apply, 'counts': stats}, ensure_ascii=False))
