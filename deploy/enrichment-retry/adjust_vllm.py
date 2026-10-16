#!/usr/bin/env python3
"""Persist exactly one vLLM change; run as root after ubuntu login."""
import datetime
import pathlib
import re
import shutil

path = pathlib.Path('/root/kaifang/vllm/docker-compose-scu.yml')
text = path.read_text()
match = re.search(r'^  scu-qwen35-9b:\s*\n(?:(?!^  [A-Za-z0-9_-]+:).)*', text, re.M | re.S)
if not match:
    raise SystemExit('Qwen3.5-9B service not found')
block = match.group()
if len(re.findall(r'--max-num-seqs\s+\d+', block)) != 1:
    raise SystemExit('expected exactly one max-num-seqs in target service')
old = re.search(r'--max-num-seqs\s+(\d+)', block).group(1)
if old not in ('8', '12'):
    raise SystemExit('unexpected current max-num-seqs: ' + old)
if old == '12':
    print('already configured: max-num-seqs=12')
else:
    backup = path.with_name(path.name + '.before-enrichment-' + datetime.datetime.now().strftime('%Y%m%dT%H%M%S'))
    shutil.copy2(path, backup)
    changed = re.sub(r'(--max-num-seqs\s+)8\b', r'\g<1>12', block, count=1)
    path.write_text(text[:match.start()] + changed + text[match.end():])
    print('backup=' + str(backup))
    print('configured: max-num-seqs=12; other text unchanged')
