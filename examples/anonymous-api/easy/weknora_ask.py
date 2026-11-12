#!/usr/bin/env python3
"""Python 3.9+ standard library only. No login, SDK or pip installation."""
import argparse
import json
import sys
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import urljoin
from urllib.request import Request, urlopen

DEFAULT_BASE = 'http://10.20.0.15:18182/api/openai/v1'


def read_stream(lines):
    """The validated final_content replaces deltas; it is never appended to them."""
    draft = []
    final = None
    metadata = {}
    done = False
    for raw in lines:
        line = raw.decode('utf-8').strip() if isinstance(raw, bytes) else raw.strip()
        if not line.startswith('data:'):
            continue
        data = line[5:].strip()
        if data == '[DONE]':
            done = True
            break
        part = json.loads(data)
        if part.get('error'):
            raise RuntimeError(str(part['error'].get('message', '生成失败')))
        metadata['id'] = part.get('id', metadata.get('id'))
        for choice in part.get('choices', []):
            draft.append(choice.get('delta', {}).get('content') or '')
        if 'final_content' in part:
            final = part['final_content']
            for key in ['citations', 'citation_status', 'usage']:
                metadata[key] = part.get(key)
    if not done or final is None:
        raise RuntimeError('流式回答未完整结束，未收到最终校正版正文；请勿把草稿当作完整结果。')
    metadata['choices'] = [{'message': {'role': 'assistant', 'content': final}}]
    return metadata


def request(base, path, body=None, stream=False):
    data = json.dumps(body, ensure_ascii=False).encode('utf-8') if body is not None else None
    req = Request(base.rstrip('/') + path, data=data,
                  headers={'Content-Type': 'application/json; charset=utf-8'})
    # No Authorization header. No automatic retries.
    with urlopen(req, timeout=310) as response:
        return read_stream(response) if stream else json.load(response)


def show(response, base, evidence=False):
    content = response['choices'][0]['message']['content']
    print('\n' + content)
    refs = response.get('citations') or []
    state = response.get('citation_status') or {}
    print('\n回答ID：' + str(response.get('id', '')))
    print('引用状态：' + str(state.get('state', 'unknown')))
    if not refs:
        print('本次没有绑定的引用，不查询原文详情。没有引用不等于答案已核实。')
        return
    print('引用来源：')
    for ref in refs:
        print('  [{}] {}'.format(ref['number'], ref['title']))
        if evidence:
            # Use only the URL actually returned for this response.
            detail_url = urljoin(base + '/', ref['detail_url'])
            if not detail_url.startswith(base.rstrip('/') + '/responses/'):
                raise RuntimeError('引用详情地址不属于本API，已停止查询。')
            with urlopen(detail_url, timeout=30) as detail:
                value = json.load(detail)
            print(value.get('snapshot', '原文暂不可用'))
    if state.get('unbound_count', 0):
        print('注意：仍有 {} 处内容未绑定有效引用。'.format(state['unbound_count']))


def main():
    p = argparse.ArgumentParser(description='WeKnora内网问答：只显示答案，自动处理引用与流式数据')
    p.add_argument('question', nargs='?', help='问题；不填则进入交互问答')
    p.add_argument('--base', default=DEFAULT_BASE)
    p.add_argument('--model', default='policy-qa')
    p.add_argument('--stream', action='store_true', help='接收SSE，完成后统一显示一次校正版正文')
    p.add_argument('--evidence', action='store_true', help='有引用时同时显示引用原文')
    p.add_argument('--max-tokens', type=int, default=1024)
    p.add_argument('--save', help='把本次完整结果保存为JSON文件；交互模式会覆盖为最新一轮')
    args = p.parse_args()
    if not 1 <= args.max_tokens <= 2048:
        p.error('当前policy-qa的max-tokens应在1到2048之间')
    history = []
    if not args.question:
        print('WeKnora多轮问答。输入问题后回车；输入 /new 清空历史，/quit 退出。')
    while True:
        q = args.question or input('\n你：').strip()
        if q.lower() in ['/quit', 'exit', 'quit']:
            break
        if q == '/new':
            history = []
            print('已清空本地历史。')
            continue
        if not q:
            continue
        messages = history + [{'role': 'user', 'content': q}]
        print('正在查询，请等待……', flush=True)
        value = request(args.base, '/chat/completions',
                        {'model': args.model, 'messages': messages, 'max_tokens': args.max_tokens,
                         'stream': args.stream}, stream=args.stream)
        show(value, args.base, args.evidence)
        if args.save:
            Path(args.save).write_text(json.dumps(value, ensure_ascii=False, indent=2), encoding='utf-8')
            print('已保存：' + str(Path(args.save).resolve()))
        history = messages + [value['choices'][0]['message']]
        # The server accepts at most 100 alternating messages including the next user.
        history = history[-98:]
        if args.question:
            break


if __name__ == '__main__':
    try:
        main()
    except HTTPError as exc:
        labels = {400: '参数不支持或输出上限过大', 403: '当前网络不在白名单',
                  404: '地址、模型别名或引用不存在', 429: '请求太多，请等60秒后手动再试',
                  502: '模型生成失败', 503: '服务暂不可用', 504: '请求超过服务端时间限制'}
        print('\n请求失败：HTTP {}，{}'.format(exc.code, labels.get(exc.code, '请检查服务状态')), file=sys.stderr)
        print(exc.read().decode('utf-8', errors='replace'), file=sys.stderr)
        sys.exit(1)
    except (URLError, TimeoutError, RuntimeError, ValueError) as exc:
        print('\n调用失败：' + str(exc), file=sys.stderr)
        sys.exit(1)
    except (KeyboardInterrupt, EOFError):
        print('\n已退出。')
