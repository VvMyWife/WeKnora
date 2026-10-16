"""Live adapter acceptance, without truncation or large vector dumps."""
import json
import urllib.request
import urllib.error


def post(path, texts):
    request = urllib.request.Request('http://127.0.0.1:18082/v1/' + path,
                                     data=json.dumps({'input': texts, 'model': 'bge-m3'}).encode(),
                                     headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(request, timeout=120) as response:
        return json.load(response)


limits = post('token-counts', ['档案检索', '表格完整行与原件溯源'])
vectors = post('embeddings', ['档案检索', '表格完整行与原件溯源'])
assert len(vectors['data']) == 2
assert all(len(x['embedding']) == 1024 for x in vectors['data'])
assert limits['max_input_tokens'] > 0 and not limits['over_limit_indices']
long_text = '档案 ' * (limits['max_input_tokens'] + 10)
try:
    post('embeddings', [long_text])
    raise AssertionError('oversized input silently accepted')
except urllib.error.HTTPError as error:
    detail = json.load(error)
    assert error.code == 400 and detail['detail']['code'] == 'EMBEDDING_INPUT_TOO_LONG'
print(json.dumps({'short_vectors': 2, 'dimensions': 1024, 'limits': limits,
                  'oversized_input': 'rejected_without_truncation'}, ensure_ascii=False))
