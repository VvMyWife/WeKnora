#!/usr/bin/env python3
"""Bounded app/18182 release. Run apply only after explicit restart authorization."""
import json
import subprocess
import sys
import time
import urllib.request
from pathlib import Path

ROOT = Path('/home/ubuntu/RAG/runtime/archive-document-acceptance-20261026/release')
COMPOSE = Path('/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json')
PROJECT = 'ljc-weknora-production'
TARGETS = {'WeKnora-app', 'WeKnora-frontend-custom-preview'}
TAGS = {'app': 'weknora-app:archive-document-20261026', 'frontend': 'weknora-ui-custom:archive-document-20261026'}


def run(args):
    result = subprocess.run(args, capture_output=True, text=True)
    if result.returncode:
        (ROOT/'command-error.log').write_text(result.stdout + result.stderr)
        raise RuntimeError('release command failed; inspect private command-error.log')
    return result.stdout


def containers():
    ids = run(['docker', 'ps', '-q']).split()
    values = json.loads(run(['docker', 'inspect', *ids]))
    return {x['Name'].lstrip('/'): {'id': x['Id'], 'image': x['Image'], 'started': x['State']['StartedAt']} for x in values if x['Name'].lstrip('/') not in TARGETS and not x['Name'].startswith('/ljc-archive-document-')}


def health():
    for _ in range(55):
        try:
            for url in ['http://127.0.0.1:18081/health', 'http://127.0.0.1:18182/']:
                with urllib.request.urlopen(url, timeout=2) as response:
                    assert response.status == 200
            return
        except Exception:
            time.sleep(1)
    raise RuntimeError('app/frontend did not become healthy')


def protected_unchanged(before):
    current = containers()
    # kabel already had 785 automatic restarts before this release. Preserve
    # its container/image identity without treating its existing loop as ours.
    for name, value in before.items():
        actual = current.get(name)
        if name == 'kabel':
            if not actual or any(actual[key] != value[key] for key in ['id', 'image']):
                return False
        elif actual != value:
            return False
    return True


ROOT.mkdir(parents=True, exist_ok=True)
ROOT.chmod(0o700)
statefile = ROOT/'state.json'
phase = sys.argv[1]
if phase == 'prepare':
    if statefile.exists():
        raise RuntimeError('release state exists; inspect status before resuming')
    baseline = COMPOSE.read_bytes()
    config = json.loads(baseline)
    old = {key: config['services'][key]['image'] for key in TAGS}
    expected = {'app': 'sha256:391885b29deae41638a817f831f0da506094022a51e7f26324b34b43125314e1', 'frontend': 'sha256:8242e2acfe8fb4b3720da0b06630150d890568801345900bf2b29673c44847b6'}
    assert old == expected, 'production baseline changed; review before release'
    images = {key: json.loads(run(['docker', 'image', 'inspect', tag]))[0] for key, tag in TAGS.items()}
    revisions = {key: value['Config']['Labels'].get('org.opencontainers.image.revision') for key, value in images.items()}
    assert len(set(revisions.values())) == 1 and all(revisions.values()), 'app/UI revisions differ'
    for key, value in images.items():
        config['services'][key]['image'] = value['Id']
    (ROOT/'compose.before.json').write_bytes(baseline)
    (ROOT/'compose.candidate.json').write_text(json.dumps(config, ensure_ascii=False, indent=2))
    state = {'old_images': old, 'new_images': {key: value['Id'] for key, value in images.items()}, 'source_revision': revisions['app'], 'protected_before': containers(), 'status': 'prepared'}
    statefile.write_text(json.dumps(state, indent=2))
    for file in ROOT.iterdir():
        file.chmod(0o600)
    print('PREPARED', state['source_revision'])
elif phase in ['apply', 'rollback']:
    state = json.loads(statefile.read_text())
    assert protected_unchanged(state['protected_before']), 'protected runtime changed before release'
    if phase == 'apply':
        assert state['status'] == 'prepared', 'already applied; inspect status'
    source = ROOT/('compose.candidate.json' if phase == 'apply' else 'compose.before.json')
    COMPOSE.write_bytes(source.read_bytes())
    COMPOSE.chmod(0o600)
    try:
        output = run(['docker', 'compose', '-p', PROJECT, '-f', str(COMPOSE), 'up', '-d', '--no-deps', 'app', 'frontend'])
        (ROOT/(phase+'.log')).write_text(output)
        health()
        after = containers()
        assert protected_unchanged(state['protected_before']), 'protected runtime changed'
        state['status'] = 'applied' if phase == 'apply' else 'rolled_back'
        state['protected_after'] = after
        statefile.write_text(json.dumps(state, indent=2))
        print(state['status'].upper(), 'app18081/frontend18182 healthy; protected containers unchanged')
    except Exception:
        if phase == 'apply':
            COMPOSE.write_bytes((ROOT/'compose.before.json').read_bytes())
            run(['docker', 'compose', '-p', PROJECT, '-f', str(COMPOSE), 'up', '-d', '--no-deps', 'app', 'frontend'])
            state['status'] = 'rolled_back_after_error'
            statefile.write_text(json.dumps(state, indent=2))
        raise
else:
    state = json.loads(statefile.read_text())
    print(json.dumps({key: state[key] for key in ['status', 'source_revision', 'old_images', 'new_images']}, indent=2))
