# WeKnora API 零基础使用说明（Windows / Linux）

照着下面操作即可。不需要登录，不需要申请密钥，也不需要SSH连接服务器。

固定信息：

| 项目 | 填什么 |
| --- | --- |
| 服务器 | `10.20.0.15:18182` |
| API基础地址 | `http://10.20.0.15:18182/api/openai/v1` |
| 模型别名 | `policy-qa`，这里表示已开放的问答配置，不是底层模型名称 |
| 密钥 | 不需要；仅在某些SDK强制要求时填 `not-required` |
| 当前允许的网络 | `10.40.0.0/22`、`10.20.0.0/22`；其他校园/VPN网段需管理员加入 |

**先选对终端：**Linux服务器/Ubuntu/Bash用下面“Bash方法”；Windows PowerShell用后面的“一”。两种语法不要混用，也不要把所有演示命令一起执行。

## Bash方法：在Ubuntu/Linux里直接复制

### 方式A：使用小工具（最省心）

如果你已经在这台服务器的Ubuntu终端里，直接执行：

```bash
cd /home/ubuntu/RAG_Weknora/runtime/worktrees/api-access/examples/anonymous-api/easy
python3 weknora_ask.py '第三级灾难恢复的数据备份要求是什么？请附引用依据。'
```

连续追问就执行下面一条，输入问题回车；输入`/quit`退出：

```bash
python3 weknora_ask.py
```

如果在别的Linux电脑，将ZIP解压到自己的文件夹，进入该文件夹再运行相同Python命令。不需要SSH登录服务器，只需网络可达并在白名单内。

### 方式B：坚持用curl，复制这一整块即可

这段只发起**一次问答**，然后显示正文，并在有引用时自动查询第一条原文。无需手动替换回答ID或引用ID。需要系统已有`curl`和`python3`。

```bash
BASE='http://10.20.0.15:18182/api/openai/v1'

if curl -fsS "$BASE/chat/completions" \
  -H 'Content-Type: application/json' \
  -d '{"model":"policy-qa","messages":[{"role":"user","content":"第三级灾难恢复的数据备份要求是什么？请附引用依据。"}],"stream":false,"max_tokens":1024}' \
  -o answer.json
then
  python3 - <<'PY'
import json
from urllib.request import urlopen
with open('answer.json', encoding='utf-8') as f:
    result = json.load(f)
print(result['choices'][0]['message']['content'])
print('\n回答ID：', result['id'])
refs = result.get('citations') or []
if refs:
    print('\n引用来源：')
    for ref in refs:
        print('[{}] {}'.format(ref['number'], ref['title']))
    with urlopen('http://10.20.0.15:18182' + refs[0]['detail_url'], timeout=30) as response:
        evidence = json.load(response)
    print('\n第一条引用原文：\n' + evidence['snapshot'])
else:
    print('\n本次没有绑定的引用，不查询原文详情。')
PY
fi
```

只改`content`后的中文问题。不要再追加原来那条`curl -N`，否则会额外发起第二个问答。完整JSON保存在当前目录`answer.json`中，下次运行会覆盖该结果文件。

追问推荐直接使用方式A的交互小工具，它会自动携带上一轮历史。单独发`messages:[{"role":"user","content":"请说明依据"}]`不会自动关联前一问。

## 一、先用最简单的方法：PowerShell直接提问

这一种不安装任何依赖。Windows里打开 **PowerShell**，依次复制下面三步。

### 第1步：检查网络和接口

```powershell
$WeKnoraBase = 'http://10.20.0.15:18182/api/openai/v1'
Invoke-RestMethod -Uri "$WeKnoraBase/models" -TimeoutSec 20 | ConvertTo-Json -Depth 5
```

看到包含 `policy-qa` 的内容即可继续。如果报403，说明当前来源不在白名单；如果连接超时，先检查校园网络/VPN。无需在浏览器登录。

### 第2步：发送一个问题

整块复制执行。只改 `content` 后面的中文问题即可。

```powershell
$QuestionBody = @{
    model = 'policy-qa'
    stream = $false
    max_tokens = 1024
    messages = @(
        @{ role = 'user'; content = '第三级灾难恢复的数据备份要求是什么？请附引用依据。' }
    )
} | ConvertTo-Json -Depth 10

$Answer = Invoke-RestMethod `
    -Uri "$WeKnoraBase/chat/completions" `
    -Method Post `
    -ContentType 'application/json; charset=utf-8' `
    -Body ([System.Text.Encoding]::UTF8.GetBytes($QuestionBody)) `
    -TimeoutSec 310

$Answer.choices[0].message.content
```

这时显示的是答案正文。请等待请求结束；首次等待几秒到几十秒都可能发生。`max_tokens` 是输出预算，不是中文字数；当前配置最多2048，先用1024即可。

### 第3步：查看引用和原文（自动判断有没有引用）

```powershell
if (@($Answer.citations).Count -gt 0) {
    $Answer.citations | Select-Object number, title, detail_url | Format-Table -Wrap
    $FirstCitation = @($Answer.citations)[0]
    $EvidenceUrl = 'http://10.20.0.15:18182' + $FirstCitation.detail_url
    $Evidence = Invoke-RestMethod -Uri $EvidenceUrl -TimeoutSec 30
    $Evidence.snapshot
} else {
    Write-Host '本次没有绑定的引用，不查询原文详情。'
}
```

不要自己猜引用ID，也不要直接复制“替换为回答ID”等占位文字去请求。`citations`为空时，就没有可查询的具体脚注。引用不完整、无可信页码的状态会如实保留，不代表已核实整篇答案。

### 可选：在同一个PowerShell窗口追问

下面会把上一轮问题和答案一并发送，形成多轮对话。

```powershell
$PreviousRequest = $QuestionBody | ConvertFrom-Json
$FollowupBody = @{
    model = 'policy-qa'
    stream = $false
    max_tokens = 1024
    messages = @(
        @{ role = 'user'; content = $PreviousRequest.messages[0].content }
        @{ role = 'assistant'; content = $Answer.choices[0].message.content }
        @{ role = 'user'; content = '其中对备份介质的存放有什么要求？' }
    )
} | ConvertTo-Json -Depth 10

$FollowupAnswer = Invoke-RestMethod `
    -Uri "$WeKnoraBase/chat/completions" `
    -Method Post `
    -ContentType 'application/json; charset=utf-8' `
    -Body ([System.Text.Encoding]::UTF8.GetBytes($FollowupBody)) `
    -TimeoutSec 310

$FollowupAnswer.choices[0].message.content
```

新开PowerShell后这些变量会消失，重新从第1步执行即可。接口不会凭回答ID自动记住历史，多轮必须带上历史文字。

## 二、想长期使用：运行附带的Python小工具

本目录的 `weknora_ask.py` 只使用Python自带模块，**不用pip安装任何包**。需要Python3.9或更新版本。

### 第1步：准备文件夹

把这个README与 `weknora_ask.py` 放在同一个文件夹，例如 `E:\WeKnora_API`。也可以直接解压提供的ZIP到这里。

```powershell
cd E:\WeKnora_API
python --version
```

如果提示找不到python，可以试 `py --version`；后面的命令将 `python` 换成 `py`。如果都没有，先用前面的PowerShell方法，不影响调用API。

在本次开发电脑上，也可以直接使用已有位置，无需复制：

```powershell
cd 'E:\desktop\文档RAG\runtime\worktrees\api-access\examples\anonymous-api\easy'
```

### 第2步：单次提问

```powershell
python -X utf8 weknora_ask.py '第三级灾难恢复的数据备份要求是什么？请附引用依据。'
```

### 第3步：连续追问

```powershell
python -X utf8 weknora_ask.py
```

在 `你：` 后输入问题并回车。下一次输入自动携带当前进程内的历史，不需要手动拼JSON。

- 输入 `/new`：清空本地历史，开始新话题。
- 输入 `/quit`：退出。
- 关闭程序后不自动保留历史。

### 其他常用命令

```powershell
# 回答后显示引用原文；没有引用就自动跳过，不会硬查不存在的脚注。
python -X utf8 weknora_ask.py '数据备份有什么要求？' --evidence

# 使用流式协议；工具会解析data事件，结束后只显示一次校验后的答案。
# 这里不会逐字显示草稿，避免和最后的完整答案重复。
python -X utf8 weknora_ask.py '档案管理有哪些主要要求？' --stream

# 保存完整结果，包含正文、引用和回答ID；文件保存在当前目录。
python -X utf8 weknora_ask.py '数据备份有什么要求？' --save answer.json

# 提高输出预算，当前上限2048。
python -X utf8 weknora_ask.py '请详细说明备份要求。' --max-tokens 2048

# 查看全部参数。
python weknora_ask.py --help
```

## 三、为什么会出现很多data、JSON和404？

当请求里设置 `"stream":true`，服务返回的是SSE流。例如：

```text
data: {"choices":[{"delta":{"content":"档案"}}], ...}
data: {"choices":[{"delta":{"content":"管理"}}], ...}
data: {"final_content":"完整校正版答案", "citations":[], ...}
data: [DONE]
```

这是正常传输格式，不是模型突然输出乱码：

1. `delta.content`：正在生成的文字片段。
2. `final_content`：完成引用校验后的完整答案。程序应使用它**替换草稿**，不能再追加到草稿后面。
3. `citations`：本次可查询的引用；空数组就是本次没有绑定引用。
4. `[DONE]`：本次回答结束，停止读取。

直接用 `curl -N` 会把这些原始事件全部打印出来。只想看到完整答案时，先用第一部分的 `stream=$false`，或者运行附带小工具。

本次粘贴内容有360个data事件，`finish_reason`为stop，最后收到DONE，说明流正常结束。你随后提供的原命令确认：第一条问答后又执行了`curl -N`，所以看到的是第二次独立问答的流。最后两条URL仍含“替换为回答ID/引用ID”等占位文字，原样执行会404。该流式回答的真实引用清单已核实返回200和空数组，没有可查询的具体脚注。不要把404理解为流式回答没有结束。

## 四、常见错误直接照表处理

| 现象 | 处理 |
| --- | --- |
| 连接不上/超时 | 检查校园内网/VPN与10.20.0.15是否可达 |
| 403 | 当前IP不在白名单，找管理员加入实际网段；填密钥无用 |
| 404 | 基础地址和别名须完全照抄；引用只能用返回的detail_url |
| 400 | 暂时只传model、messages、stream、max_tokens；不要加temperature/tools等 |
| 429 | 等60秒后手动重试；同一出口IP共享并发和频次额度 |
| 502/503/504 | 保存错误提示交管理员；不要写无限重试 |
| 没有脚注/引用待核验 | 是引用绑定状态，不要自行编造脚注或页码 |
| 正文出现HTML span | 是引用待核验等标记；普通终端不会像网页那样渲染样式 |

本API只开放配置好的问答资料，不是任意模型代理；没有上传文件或匿名修改知识库功能。所有获准内网调用者共享开放范围，回答ID不代表个人私密会话。

## 五、需要自行检查小工具时

在README与脚本所在文件夹执行；Linux将`python`换成`python3`：

```powershell
python -m unittest -v test_weknora_ask.py
```

测试验证最终答案替换草稿、错误与中断不会伪装成成功、没有引用时不发送原文查询。实际网络测试可使用上面的单次提问命令；会真实调用模型，不自动重试。
