import pytest
from adapters.bge_m3.input_limits import input_counts

class FakeModel:
    max_seq_length=8
    def tokenizer(self,texts,**kwargs):
        assert kwargs['truncation'] is False
        assert kwargs['add_special_tokens'] is True
        return {'input_ids':[[0]*(len(t)+2) for t in texts]}

def test_limits_include_special_tokens_and_do_not_truncate():
    result=input_counts(FakeModel(),['123456','1234567'])
    assert result['input_tokens']==[8,9]
    assert result['over_limit_indices']==[1]
    assert not result['truncation']

def test_empty_rejected():
    with pytest.raises(ValueError):input_counts(FakeModel(),[])
