#!/usr/bin/env python3
"""Prepare private configuration for a fresh dedicated acceptance stack; no production writes."""
import argparse
import json
import os
from pathlib import Path
import secrets
import subprocess

TARGET = 'ljc-upstream-acceptance-postgres'
LABEL = 'upstream-20261020'
ROLE = 'ljc_acceptance'


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--private-dir', type=Path, default=Path('/home/ubuntu/weknora-fork-migration/20261020-upstream'))
    args = parser.parse_args()
    private = args.private_dir.resolve()
    private.mkdir(mode=0o700, parents=True, exist_ok=True)
    env_path = private / 'acceptance.env'
    if env_path.exists():
        values = dict(line.split('=', 1) for line in env_path.read_text().splitlines() if '=' in line and not line.startswith('#'))
        if values.get('DB_HOST') != TARGET or values.get('DB_NAME') != ROLE or values.get('DB_USER') != ROLE:
            raise SystemExit('Existing environment does not target the dedicated acceptance database')
    else:
        password, redis_password, graph_password = [secrets.token_hex(24) for _ in range(3)]
        values = {
            'DB_DRIVER':'postgres', 'DB_HOST':TARGET, 'DB_PORT':'5432',
            'DB_USER':ROLE, 'DB_PASSWORD':password, 'DB_NAME':ROLE,
            'POSTGRES_USER':ROLE, 'POSTGRES_PASSWORD':password, 'POSTGRES_DB':ROLE,
            'LJC_MIGRATION_DSN':f'postgres://{ROLE}:{password}@{TARGET}:5432/{ROLE}?sslmode=disable',
            'AUTO_MIGRATE':'false', 'AUTO_RECOVER_DIRTY':'false', 'RETRIEVE_DRIVER':'postgres',
            'ARCHIVAL_CITATIONS_ENABLED':'true',
            'REDIS_ADDR':'ljc-upstream-acceptance-redis:6379', 'REDIS_PASSWORD':redis_password,
            'REDIS_DB':'0', 'REDIS_PREFIX':'ljc-acceptance:', 'WEKNORA_REDIS_NAMESPACE':'ljc-upstream-20261020',
            'NEO4J_ENABLE':'true', 'NEO4J_URI':'bolt://ljc-upstream-acceptance-neo4j:7687',
            'NEO4J_USERNAME':'neo4j', 'NEO4J_PASSWORD':graph_password, 'NEO4J_AUTH':'neo4j/'+graph_password,
            'DOCREADER_ADDR':'ljc-upstream-acceptance-docreader:50051', 'DOCREADER_TRANSPORT':'grpc',
            'STORAGE_TYPE':'local', 'LOCAL_STORAGE_BASE_DIR':'/data/files',
            'JWT_SECRET':secrets.token_hex(32), 'SYSTEM_AES_KEY':secrets.token_hex(16), 'SYSTEM_SIGNING_KEY':secrets.token_hex(32),
            'GIN_MODE':'release', 'LOG_LEVEL':'info', 'TZ':'Asia/Shanghai', 'WEKNORA_LANGUAGE':'zh-CN',
            'WEKNORA_BOOTSTRAP_SYSTEM_ADMIN_EMAIL':'ljc-acceptance@example.test',
            'WEKNORA_SANDBOX_DOCKER_ENABLED':'false', 'LANGFUSE_ENABLED':'false',
            'SSRF_WHITELIST':'10.20.0.15,localhost,127.0.0.1',
            'ACCEPTANCE_USER_EMAIL':'ljc-acceptance@example.test', 'ACCEPTANCE_USER_PASSWORD':secrets.token_urlsafe(24),
        }
        descriptor = os.open(env_path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(descriptor, 'w') as out:
            out.write(''.join(f'{key}={value}\n' for key, value in values.items()))
    if 'RETRIEVE_DRIVER' not in values:
        with env_path.open('a') as out:
            out.write('RETRIEVE_DRIVER=postgres\n')
        values['RETRIEVE_DRIVER'] = 'postgres'
    if 'ARCHIVAL_CITATIONS_ENABLED' not in values:
        with env_path.open('a') as out:
            out.write('ARCHIVAL_CITATIONS_ENABLED=true\n')
        values['ARCHIVAL_CITATIONS_ENABLED'] = 'true'
    if values.get('RETRIEVE_DRIVER') != 'postgres':
        raise SystemExit('Acceptance requires its dedicated PostgreSQL retrieval engine')
    # PostgreSQL starts separately with its own volume and database owner.
    # Never mutate the restored migration-proof instances or a production server.
    running = subprocess.run(['docker', 'inspect', '-f', '{{json .Config.Labels}}', TARGET], text=True, capture_output=True)
    if running.returncode == 0:
        labels = json.loads(running.stdout)
        if labels.get('ljc.operation') != LABEL:
            raise SystemExit('Refusing a database container outside this integration')
    print(json.dumps({'private_environment':str(env_path), 'database':ROLE, 'container':TARGET, 'production_modified':False}))


if __name__ == '__main__':
    main()
