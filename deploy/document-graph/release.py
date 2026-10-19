"""Guarded deployment/rollback. Requires an explicit release directory and action."""
import argparse
import http.client
import json
from pathlib import Path
import socket
import subprocess
import time
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
p = argparse.ArgumentParser()
p.add_argument('directory', type=Path)
p.add_argument('action', choices=['check', 'deploy', 'candidate', 'frontend', 'rollback'])
p.add_argument('--image')
p.add_argument('--frontend-commit')
args = p.parse_args()
folder = args.directory.resolve()
if not folder.is_relative_to(ROOT / 'runtime' / 'document-graph'):
    raise SystemExit('release directory must be inside runtime/document-graph')
release = json.loads((folder / 'release.json').read_text())
before = json.loads((folder / 'app-before.json').read_text())[0]
ui_before = json.loads((folder / 'frontend-before.json').read_text())[0]
original = json.loads((folder / 'original-before.json').read_text())[0]
old_ui_name = 'WeKnora-frontend-custom-before-document-graph-' + folder.name

def run(*cmd):
    return subprocess.check_output(cmd, text=True, cwd=ROOT).strip()

def inspect(name):
    return json.loads(run('docker', 'inspect', name))[0]

def save(name, data):
    target = folder / name
    target.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
    target.chmod(0o600)

def compose(image):
    save('compose.runtime.json', {'services': {'app': {'image': image, 'environment': dict(value.split('=', 1) for value in before['Config']['Env'])}}})
    labels = before['Config']['Labels']
    cmd = ['docker', 'compose', '-p', labels['com.docker.compose.project']]
    for file in labels['com.docker.compose.project.config_files'].split(',') + [str(folder / 'compose.runtime.json')]:
        cmd += ['-f', file]
    return cmd

def validate():
    if inspect('WeKnora-app')['Image'] != release['previous_app_id']:
        raise RuntimeError('app baseline changed; do not overwrite another release')
    if inspect('WeKnora-frontend-custom-preview')['Image'] != release['previous_frontend_id']:
        raise RuntimeError('custom frontend baseline changed')
    if inspect('WeKnora-frontend')['Image'] != original['Image']:
        raise RuntimeError('original frontend baseline changed')
    config = json.loads(run(*compose(release['app_image']), 'config', '--format', 'json'))
    app = config['services']['app']
    if app['environment'] != dict(value.split('=', 1) for value in before['Config']['Env']):
        raise RuntimeError('environment drift')
    old_mounts = {m['Destination']: (m.get('Name') if m['Type'] == 'volume' else m['Source'], not m['RW']) for m in before['Mounts']}
    mounts = {m['target']: (config.get('volumes', {}).get(m['source'], {}).get('name', m['source']), m.get('read_only', False)) for m in app.get('volumes', [])}
    if old_mounts != mounts:
        raise RuntimeError('mount drift')
    old_ports = {(target.split('/')[0], port['HostPort']) for target, ports in before['HostConfig'].get('PortBindings', {}).items() for port in (ports or [])}
    new_ports = {(str(port['target']), str(port['published'])) for port in app.get('ports', [])}
    if old_ports != new_ports:
        raise RuntimeError('port drift')
    print('PASS: runtime environment/mounts preserved; original frontend untouched')

def healthy():
    for _ in range(30):
        try:
            with urllib.request.urlopen('http://127.0.0.1:18081/health', timeout=2) as response:
                if response.status == 200:
                    return
        except Exception:
            pass
        time.sleep(2)
    raise RuntimeError('backend health timeout')

def replace_app(image):
    subprocess.run(compose(image) + ['up', '-d', '--no-deps', 'app'], cwd=ROOT, check=True)
    healthy()

class DockerConnection(http.client.HTTPConnection):
    def connect(self):
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect('/var/run/docker.sock')

def clone_frontend():
    # Preserve complete Docker Config/HostConfig, not just a subset of env/ports.
    config = dict(ui_before['Config'])
    config['Image'] = release['frontend_image']
    config['HostConfig'] = ui_before['HostConfig']
    conn = DockerConnection('localhost')
    conn.request('POST', '/containers/create?name=' + urllib.parse.quote('WeKnora-frontend-custom-preview'), json.dumps(config), {'Content-Type': 'application/json'})
    response = conn.getresponse()
    body = response.read().decode()
    conn.close()
    if response.status != 201:
        raise RuntimeError('frontend create failed: ' + body)
    subprocess.run(['docker', 'start', 'WeKnora-frontend-custom-preview'], check=True)
    for _ in range(15):
        try:
            with urllib.request.urlopen('http://127.0.0.1:18182/login', timeout=2) as page:
                if page.status == 200:
                    return
        except Exception:
            pass
        time.sleep(1)
    raise RuntimeError('custom login page unavailable')

def restore_frontend():
    names = run('docker', 'ps', '-a', '--format', '{{.Names}}').splitlines()
    if old_ui_name not in names:
        return
    if 'WeKnora-frontend-custom-preview' in names:
        subprocess.run(['docker', 'stop', 'WeKnora-frontend-custom-preview'], check=True)
        subprocess.run(['docker', 'rename', 'WeKnora-frontend-custom-preview', 'WeKnora-frontend-custom-document-graph-stopped-' + str(int(time.time()))], check=True)
    subprocess.run(['docker', 'rename', old_ui_name, 'WeKnora-frontend-custom-preview'], check=True)
    subprocess.run(['docker', 'start', 'WeKnora-frontend-custom-preview'], check=True)

if args.action == 'check':
    validate()
elif args.action == 'candidate':
    if not args.image or not args.frontend_commit:
        raise RuntimeError('candidate requires --image and --frontend-commit')
    if run('git', '-C', 'frontend-custom', 'rev-parse', 'HEAD') != args.frontend_commit:
        raise RuntimeError('frontend source revision does not match HEAD')
    if inspect('WeKnora-app')['Image'] != release['app_image_id']:
        raise RuntimeError('expected backend is not running')
    save('frontend-candidate-before-' + str(int(time.time())) + '.json', release)
    release.update(frontend_image=args.image, frontend_image_id=inspect(args.image)['Id'], frontend_commit=args.frontend_commit)
    save('release.json', release)
    print('Frontend candidate recorded; no service changed')
elif args.action == 'rollback':
    if inspect('WeKnora-app')['Image'] not in (release['app_image_id'], release['previous_app_id']):
        raise RuntimeError('a different release is active; refusing rollback')
    if inspect('WeKnora-frontend-custom-preview')['Image'] not in (release['frontend_image_id'], release['previous_frontend_id']):
        raise RuntimeError('a different custom frontend release is active; refusing rollback')
    restore_frontend()
    replace_app(release['previous_app_id'])
    save('rollback.json', {'time': time.time(), 'app_image': inspect('WeKnora-app')['Image']})
elif args.action == 'frontend':
    if inspect('WeKnora-app')['Image'] != release['app_image_id']:
        raise RuntimeError('expected backend is not running')
    if inspect('WeKnora-frontend-custom-preview')['Image'] != release['previous_frontend_id']:
        raise RuntimeError('custom baseline changed')
    if inspect('WeKnora-frontend')['Image'] != original['Image']:
        raise RuntimeError('original frontend changed')
    try:
        subprocess.run(['docker', 'stop', 'WeKnora-frontend-custom-preview'], check=True)
        subprocess.run(['docker', 'rename', 'WeKnora-frontend-custom-preview', old_ui_name], check=True)
        clone_frontend()
        save('frontend-deployed.json', {'time': time.time(), 'frontend_image': inspect('WeKnora-frontend-custom-preview')['Image'], 'frontend_commit': release['frontend_commit'], 'backend_unchanged': True})
        print('Custom frontend deployed; backend not restarted')
    except Exception:
        restore_frontend()
        raise
else:
    validate()
    try:
        replace_app(release['app_image'])
        subprocess.run(['docker', 'stop', 'WeKnora-frontend-custom-preview'], check=True)
        subprocess.run(['docker', 'rename', 'WeKnora-frontend-custom-preview', old_ui_name], check=True)
        clone_frontend()
        if inspect('WeKnora-frontend')['Image'] != original['Image']:
            raise RuntimeError('original frontend changed')
        save('deployed.json', {'time': time.time(), 'app_image': inspect('WeKnora-app')['Image'], 'frontend_image': inspect('WeKnora-frontend-custom-preview')['Image'], 'backend_commit': release['backend_commit'], 'frontend_commit': release['frontend_commit']})
        print('Deployed; previous images and frontend container retained')
    except Exception:
        restore_frontend()
        replace_app(release['previous_app_id'])
        raise
