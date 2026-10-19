<template>
  <section class="document-graph-toolbar" aria-label="图谱档案范围">
    <div class="scope-row">
      <button :aria-pressed="scope === 'kb'" @click="emit('scope', 'kb')">全库概览</button>
      <button :aria-pressed="scope === 'documents'" @click="selectDocuments">按档案</button>
      <button v-if="scope === 'documents'" @click="open">选择档案（{{ documents.length }}/50）</button>
      <select v-if="scope === 'documents'" :value="view" aria-label="档案图谱展示方式" @change="emit('view', ($event.target as HTMLSelectElement).value as 'union' | 'connections')">
        <option value="union">合并展示并突出关联</option>
        <option value="connections">只看档案间关联</option>
      </select>
    </div>
    <details v-if="scope === 'documents' && documents.length">
      <summary>已选 {{ documents.length }} 份档案 · {{ meta?.returned || 0 }}/{{ meta?.total || 0 }} 个节点{{ meta?.truncated ? '（已截取，可搜索或展开）' : '' }}</summary>
      <ul class="coverage-list"><li v-for="doc in documents" :key="doc.id" :title="documentGraphPath(doc)">
        {{ documentGraphLabel(doc) }}：{{ coverage(doc.id)?.returned || 0 }}/{{ coverage(doc.id)?.total || 0 }} 节点
      </li></ul>
    </details>
    <small v-if="scope === 'documents'">紫色外圈：共享来源；紫色连线：已有 Wiki 跨来源链接，并非已证实的业务关系。</small>
    <Teleport to="body">
    <div v-if="visible" class="selector-backdrop" @click.self="visible = false">
      <section class="selector-dialog" role="dialog" aria-modal="true" aria-label="选择图谱档案">
        <h3>选择图谱档案</h3>
        <p>从当前知识库所有文件夹选择，最多 50 份。已映射原件显示原件名称；无 Wiki 来源时显示 0 个节点。</p>
        <form class="scope-row" @submit.prevent="page = 1; load()">
          <input v-model="keyword" placeholder="搜索档案名称" aria-label="搜索档案名称" />
          <button type="submit">搜索</button><span>已选 {{ draft.length }}/50</span>
        </form>
        <p v-if="error" role="alert">{{ error }}</p>
        <div class="document-options" :aria-busy="loading">
          <p v-if="loading">正在加载档案…</p>
          <p v-else-if="!rows.length">没有匹配的档案</p>
          <label v-for="doc in rows" :key="doc.id" class="document-option">
            <input type="checkbox" :checked="draft.some(item => item.id === doc.id)" :disabled="!draft.some(item => item.id === doc.id) && draft.length >= 50" @change="toggle(doc)" />
            <span><strong>{{ documentGraphLabel(doc) }}</strong><small>{{ documentGraphPath(doc) }}</small><small v-if="doc.archive_source">解析文件：{{ doc.file_name || doc.title }} · {{ doc.folder_path || '/' }}</small></span>
          </label>
        </div>
        <footer class="scope-row">
          <button :disabled="page <= 1 || loading" @click="page--; load()">上一页</button><span>第 {{ page }} 页</span>
          <button :disabled="!hasNext || loading" @click="page++; load()">下一页</button>
          <button @click="draft = []">清空选择</button><button @click="visible = false">取消</button>
          <button :disabled="!draft.length || loading" @click="apply">应用选择</button>
        </footer>
      </section>
    </div>
    </Teleport>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { listKnowledgeFiles } from '@/api/knowledge-base'
import type { WikiGraphMeta } from '@/api/wiki'
import { documentGraphLabel, documentGraphPath, type GraphDocument } from './documentGraphScope'
const props = defineProps<{ kbId: string; scope: 'kb' | 'documents'; documents: GraphDocument[]; view: 'union' | 'connections'; meta?: WikiGraphMeta }>()
const emit = defineEmits<{ scope: [value: 'kb' | 'documents']; documents: [value: GraphDocument[]]; view: [value: 'union' | 'connections'] }>()
const visible = ref(false), loading = ref(false), error = ref(''), keyword = ref(''), page = ref(1), hasNext = ref(false)
const rows = ref<GraphDocument[]>([]), draft = ref<GraphDocument[]>([])
let seq = 0
const coverage = (id: string) => props.meta?.documents?.find(doc => doc.knowledge_id === id)
function open() { draft.value = [...props.documents]; keyword.value = ''; page.value = 1; visible.value = true; void load() }
function selectDocuments() { emit('scope', 'documents'); if (!props.documents.length) open() }
function toggle(doc: GraphDocument) { const index = draft.value.findIndex(item => item.id === doc.id); if (index >= 0) draft.value.splice(index, 1); else if (draft.value.length < 50) draft.value.push(doc) }
function apply() { emit('documents', [...draft.value]); visible.value = false }
async function load() {
  const token = ++seq, kb = props.kbId
  loading.value = true; error.value = ''
  try {
    const res: any = await listKnowledgeFiles(kb, {page:page.value, page_size:50, keyword:keyword.value.trim(), include_archive_source:true})
    if (token !== seq || kb !== props.kbId) return
    const data = res?.data
    rows.value = Array.isArray(data) ? data : (data?.data || data?.knowledges || res?.knowledges || [])
    const total = res?.total ?? data?.total
    hasNext.value = typeof total === 'number' ? page.value * 50 < total : rows.value.length === 50
  } catch { if (token === seq) { error.value = '档案列表加载失败，请重试'; rows.value = [] } }
  finally { if (token === seq) loading.value = false }
}
watch(() => props.kbId, () => { ++seq; visible.value = false; rows.value = []; draft.value = []; loading.value = false })
</script>

<style scoped>
.document-graph-toolbar { position:absolute;left:16px;top:16px;z-index:30;max-width:calc(100% - 260px);background:var(--td-bg-color-container,#fff);padding:10px 12px;border:1px solid var(--td-component-border,#ddd);border-radius:8px;box-shadow:var(--td-shadow-1);font-size:12px }
.scope-row { display:flex;gap:8px;align-items:center;flex-wrap:wrap }
button, select, input { font:inherit;padding:5px 8px;border:1px solid var(--td-component-border,#ddd);border-radius:4px;background:var(--td-bg-color-container,#fff);color:var(--td-text-color-primary,#222) }
button { cursor:pointer } button[aria-pressed=true] { color:#0052d9;border-color:#0052d9 } button:disabled { opacity:.45;cursor:default }
small { display:block;color:var(--td-text-color-secondary,#666);margin-top:5px } details {margin-top:8px} summary {cursor:pointer}
.coverage-list {max-height:130px;overflow:auto;padding-left:20px}
.selector-backdrop {position:fixed;inset:0;background:#0006;display:flex;align-items:center;justify-content:center;z-index:2000}
.selector-dialog {width:min(760px,90vw);background:var(--td-bg-color-container,#fff);padding:24px;border-radius:10px;box-shadow:0 10px 40px #0003}
.document-options {height:min(420px,50vh);overflow:auto;margin:16px 0;border-block:1px solid #ddd}
.document-option {display:flex;align-items:flex-start;gap:12px;padding:10px 4px;border-bottom:1px solid #eee;cursor:pointer}.document-option span {min-width:0;overflow-wrap:anywhere}
@media(max-width:900px){.document-graph-toolbar{max-width:calc(100% - 32px)}}
</style>
