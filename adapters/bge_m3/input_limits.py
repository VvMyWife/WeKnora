"""Tokenizer-based limits. Never silently truncate an imported evidence chunk."""
def input_counts(model, texts):
    if not texts or len(texts) > 1024 or any(not isinstance(text, str) for text in texts):
        raise ValueError('input must contain 1..1024 strings')
    if sum(len(text.encode('utf-8')) for text in texts) > 16 * 1024 * 1024:
        raise ValueError('input exceeds adapter request limit: 16 MiB')
    maximum = int(model.max_seq_length)
    encoded = model.tokenizer(texts, truncation=False, add_special_tokens=True,
                              return_attention_mask=False, return_token_type_ids=False)
    counts = [len(ids) for ids in encoded['input_ids']]
    return {'max_input_tokens': maximum, 'input_tokens': counts,
            'over_limit_indices': [i for i, count in enumerate(counts) if count > maximum],
            'truncation': False}
