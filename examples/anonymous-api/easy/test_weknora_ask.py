import contextlib
import io
import json
import unittest
from unittest.mock import patch
from weknora_ask import read_stream, show, DEFAULT_BASE


def event(value):
    return 'data: ' + json.dumps(value, ensure_ascii=False) + '\n'


class ClientTests(unittest.TestCase):
    def test_final_replaces_draft(self):
        value = read_stream([
            event({'id': 'r', 'choices': [{'delta': {'content': '未校验草稿'}}]}),
            event({'id': 'r', 'choices': [{'delta': {}, 'finish_reason': 'stop'}],
                   'final_content': '最终答案[1]', 'citations': [{'number': 1}],
                   'citation_status': {'state': 'bound'}}),
            'data: [DONE]\n', 'curl error unrelated to this SSE stream'])
        self.assertEqual(value['choices'][0]['message']['content'], '最终答案[1]')

    def test_unfinished_stream_is_not_success(self):
        with self.assertRaises(RuntimeError):
            read_stream([event({'choices': [{'delta': {'content': 'partial'}}]})])

    def test_error_is_not_success(self):
        with self.assertRaises(RuntimeError):
            read_stream([event({'error': {'message': 'failed'}}), 'data: [DONE]'])

    def test_no_citations_means_no_detail_request(self):
        with patch('weknora_ask.urlopen') as network, contextlib.redirect_stdout(io.StringIO()):
            show({'id': 'r', 'choices': [{'message': {'content': 'answer'}}], 'citations': []},
                 DEFAULT_BASE, evidence=True)
            network.assert_not_called()


if __name__ == '__main__':
    unittest.main()
