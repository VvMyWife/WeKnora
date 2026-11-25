# PDF 物理页溯源

原生 RAG 切块导出和 WeKnora 自己上传 PDF 均可提供物理页引用。外部分块包使用零基 `page_index`，WeKnora 引用与界面使用从 1 起算的 PDF 物理页。版本绑定依据原件字节的 SHA-256。空白页保留位置，跨页块保存实际来源页。旧资料沿用现有原件预览。

## 原生 RAG

```bash
cd /home/ubuntu/RAG
git status --short --branch
.venv/bin/pytest -q tests/test_contextual.py tests/test_chunker.py tests/test_folder_input.py tests/test_table_version_export.py tests/test_weknora_export.py
.venv/bin/pytest -q tests/test_folder_ingest.py
./scripts/ragctl.sh ingest --help
```

指定已有 OCR JSON 与 Markdown 的目录，先 `--dry-run` 核对实际选中的 PDF，再在新 dataset 上验证。

```bash
read -r -p 'OCR 输出目录 ' INPUT_DIR
DATASET='物理页验收'
OUTPUT_DIR='/home/ubuntu/RAG/runtime/results/物理页验收'
./scripts/ragctl.sh ingest --input-dir "$INPUT_DIR" --dataset "$DATASET" --output-dir "$OUTPUT_DIR" --dry-run
./scripts/ragctl.sh ingest --input-dir "$INPUT_DIR" --dataset "$DATASET" --output-dir "$OUTPUT_DIR" --export-weknora
find "$OUTPUT_DIR" -name weknora-chunks-v1.zip -type f
```

新包标记 `generation_config.pdf_page_mapping=physical-v1`。来源校验使用 OCR 记录的原件 SHA-256，或对应 `source_uri` 的当前字节与已归档原件一致。OCR 页列表须包含空白页并与 PDF 页数一致。只给能证明来源的 PDF chunk 输出页锚点。非 PDF 的原有来源锚点继续导出。无页锚点的文档可按外部分块包现有流程导入。

## WeKnora 双路径验收

```bash
cd /home/ubuntu/RAG_Weknora
p="$PWD/runtime/pdf-page-provenance"
git status --short --branch
git log -6 --oneline
ss -Hlt '( sport = :28281 or sport = :28282 )'
docker run --rm --pull never --network none -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" \
  -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:release-20261112-tests \
  go test ./internal/archivemapping ./internal/archivalcitations ./internal/externalchunks ./internal/sourceloc -count=1
bash "$p/test-ui.sh"
sha=$(cat "$p/build-source.sha")
git diff --exit-code "$sha" HEAD -- internal frontend-custom migrations
backend=$(cat "$p/backend-source.sha")
git diff --exit-code "$backend" HEAD -- internal migrations
bash "$p/build-app.sh" "$backend"
bash "$p/build-ui.sh" "$sha"
docker start ljc-20261125-proof-pg ljc-20261125-embedding
docker compose -f "$p/compose.candidate.json" up -d --pull never
curl -fsS --retry 20 --retry-all-errors --retry-delay 2 http://127.0.0.1:28281/health
python3 "$p/candidate-test.py"
```

候选测试脚本使用两个隔离知识库，分别上传样例 PDF 和导入外部分块 ZIP。打开验收日志记录的问答，核对第 3 页缩略图、空白第 2 页、点击原地放大和原件校验。短文跨页形成同一 chunk 时，缩略图包含该块实际覆盖的来源页。

## 发布、自检与回退

确认候选服务验收后停止，发布脚本检查生产活动任务和原数据库账本。部署前后 Compose 和日志在运行目录中保存。

```bash
cd /home/ubuntu/RAG_Weknora
p=runtime/pdf-page-provenance
docker compose -f "$p/compose.candidate.json" stop
docker stop ljc-20261125-proof-pg ljc-20261125-embedding
nohup bash -c 'python3 -u runtime/pdf-page-provenance/deploy.py; echo $? > runtime/pdf-page-provenance/deploy.status' > "$p/deploy.log" 2>&1 < /dev/null &
echo "发布 PID=$!"
tail -n 20 "$p/deploy.log"
docker inspect WeKnora-app WeKnora-frontend-custom-preview --format '{{.Name}} {{.Image}}'
curl -fsS http://127.0.0.1:18081/health
curl -fsSI http://127.0.0.1:18182/
```

中断后先核对容器镜像、日志及 `deploy.status`，再决定是否需要恢复。回退仍使用同一数据库。

```bash
p=runtime/pdf-page-provenance
docker compose -p ljc-weknora-production -f "$p/compose.before.json" config --quiet
docker compose -p ljc-weknora-production -f "$p/compose.before.json" up -d --pull never --no-deps app frontend
curl -fsS --retry 20 --retry-all-errors --retry-delay 2 http://127.0.0.1:18081/health
curl -fsSI http://127.0.0.1:18182/
install -m 600 "$p/compose.before.json" runtime/production/compose.json.rollback
mv runtime/production/compose.json.rollback runtime/production/compose.json
```
