#!/usr/bin/env python3
"""Pinned frontend-only promotion; preserves every other container and setting."""
import argparse
import copy
import json
import os
from pathlib import Path
import subprocess
import time
import urllib.request

COMPOSE = Path('/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json')
TARGET = 'WeKnora-frontend-custom-preview'
BASELINE = 'sha256:79ab0a372e3858ac37dbabc10156e0e81b3a340e06cfde0107f56a200a3e0c79'

def run(*args):
    return subprocess.check_output(args, text=True).strip()

def containers():
    ids = run('docker', 'ps', '-aq').splitlines()
    rows = json.loads(run('docker', 'inspect', *ids))
    return {c['Name'].lstrip('/'): (c['Id'], c['Image']) for c in rows}

def up():
    run('docker', 'compose', '-f', str(COMPOSE), 'up', '-d', '--no-deps', 'frontend')

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--image', required=True)
    parser.add_argument('--apply', action='store_true')
    args = parser.parse_args()
    if not args.image.startswith('sha256:') or len(args.image) != 71:
        raise SystemExit('Use an immutable image ID')
    assert run('docker', 'image', 'inspect', '--format', '{{.Id}}', args.image) == args.image
    before = containers()
    if before[TARGET][1] != BASELINE:
        raise SystemExit('Baseline changed or release already applied; inspect rather than replay')
    original = COMPOSE.read_bytes()
    config = json.loads(original)
    candidate = copy.deepcopy(config)
    candidate['services']['frontend']['image'] = args.image
    comparison = copy.deepcopy(candidate)
    comparison['services']['frontend']['image'] = config['services']['frontend']['image']
    assert comparison == config
    if not args.apply:
        print(json.dumps({'frontend_before': BASELINE, 'frontend_after': args.image, 'protected': len(before)-1}))
        return
    release = COMPOSE.parent / ('citation-notice-ui-' + args.image[7:19])
    release.mkdir(mode=0o700, exist_ok=False)
    backup = release / 'compose.before.json'
    backup.write_bytes(original)
    os.chmod(backup, 0o600)
    staged = release / 'compose.after.json'
    staged.write_text(json.dumps(candidate, indent=2) + '\n')
    os.chmod(staged, 0o600)
    run('docker', 'compose', '-f', str(staged), 'config', '--quiet')
    assert COMPOSE.read_bytes() == original and containers() == before
    try:
        COMPOSE.write_bytes(staged.read_bytes())
        up()
        run('docker', 'exec', TARGET, 'nginx', '-t')
        for attempt in range(20):
            try:
                with urllib.request.urlopen('http://127.0.0.1:18182/', timeout=3) as response:
                    assert response.status == 200
                break
            except Exception:
                if attempt == 19:
                    raise
                time.sleep(1)
        after = containers()
        assert after[TARGET][1] == args.image
        for name, state in before.items():
            if name != TARGET:
                assert after.get(name) == state, f'Protected container changed: {name}'
        result = {'status': 'promoted', 'image': args.image, 'baseline': BASELINE,
                  'protected_unchanged': len(before)-1, 'rollback_compose': str(backup)}
        (release / 'result.json').write_text(json.dumps(result, indent=2) + '\n')
        print(json.dumps(result))
    except Exception:
        COMPOSE.write_bytes(original)
        up()
        raise

if __name__ == '__main__':
    main()
