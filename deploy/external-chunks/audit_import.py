"""Read-only DB audit of a live external import against its original ZIP."""
import argparse
import json
import subprocess
import uuid
import zipfile


def query(sql):
    result = subprocess.run(['docker', 'exec', '-i', 'WeKnora-postgres', 'sh', '-c',
                             'exec psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At'],
                            input=sql, text=True, capture_output=True, check=True)
    return json.loads(result.stdout.strip())


parser = argparse.ArgumentParser()
parser.add_argument('knowledge_id', type=uuid.UUID)
parser.add_argument('package')
args = parser.parse_args()
kid = str(args.knowledge_id)
record = query(f"SELECT row_to_json(r) FROM external_chunk_records r WHERE knowledge_id='{kid}';")
rows = query(f"SELECT coalesce(json_agg(c ORDER BY c.chunk_index),'[]') FROM "
             f"(SELECT id,tenant_id,chunk_index,content,source_content,metadata,index_status FROM chunks "
             f"WHERE knowledge_id='{kid}' AND chunk_type='text' AND deleted_at IS NULL) c;")
with zipfile.ZipFile(args.package) as package:
    expected = [json.loads(x) for x in package.read('chunks.jsonl').decode().splitlines()]
expected = [x for x in expected if x['document_id'] == record['external_document_id']]
assert len(rows) == len(expected), (len(rows), len(expected))
for actual, original in zip(rows, expected):
    assert actual['id'] == str(uuid.uuid5(uuid.NAMESPACE_URL, record['id'] + '\0' + original['id']))
    assert actual['content'] == actual['source_content'] == original['content']
    assert actual['chunk_index'] == original['order']
    assert actual['tenant_id'] == record['tenant_id']
    assert actual['metadata']['external_import']['source_anchors'] == original['source_anchors']
status = query(f"SELECT row_to_json(k) FROM (SELECT parse_status,summary_status,error_message,enable_status "
               f"FROM knowledges WHERE id='{kid}') k;")
print(json.dumps({'knowledge_id': kid, 'unchanged_chunks': len(rows), 'stable_ids': True,
                  'source_anchors': True, 'status': status}, ensure_ascii=False))
