<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue';
import { getArchiveDocumentContext, retryArchiveFailures } from '@/api/knowledge-base/index';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { MessagePlugin } from 'tdesign-vue-next';

const props = defineProps<{ knowledgeId: string; view: 'wiki' | 'graph' | 'processing'; canEdit?: boolean }>();
const emit = defineEmits<{ (e: 'locate-chunk', id: string): void }>();
const data = ref<any>(null);
const error = ref('');
const busy = ref(false);
function renderWiki(content: string) {
  const display = (content || '').replace(/\[\[[^|\]]+\|([^\]]+)\]\]/g, '$1').replace(/\[\[([^\]]+)\]\]/g, '$1');
  return DOMPurify.sanitize(marked.parse(display, { async: false }) as string);
}
let epoch = 0;
let timer: ReturnType<typeof setTimeout> | undefined;
async function load() {
  const current = ++epoch;
  const id = props.knowledgeId;
  try {
    const response: any = await getArchiveDocumentContext(id);
    if (current !== epoch || id !== props.knowledgeId) return;
    data.value = response.data;
    error.value = '';
    if (timer) clearTimeout(timer);
    timer = setTimeout(load, 10000);
  } catch { if (current === epoch) error.value = '档案信息读取失败，请刷新重试'; }
}
watch(() => props.knowledgeId, () => { data.value = null; void load(); }, { immediate: true });
onUnmounted(() => { epoch++; if (timer) clearTimeout(timer); });
async function retry() {
  busy.value = true;
  try {
    const result: any = await retryArchiveFailures(props.knowledgeId);
    MessagePlugin.success(result.data?.accepted ? '已提交失败阶段重试；暂停的知识库将等待恢复' : '没有需要新增提交的失败任务，或任务已在队列中');
    await load();
  } catch { MessagePlugin.error('重试提交失败，请稍后再试'); }
  finally { busy.value = false; }
}
const pages = computed(() => [...(data.value?.wiki_pages || [])].sort((a: any, b: any) => Number(b.page_type === 'summary') - Number(a.page_type === 'summary')));
const graphNodes = computed(() => {
  const degree = new Map<string, number>();
  for (const edge of data.value?.graph_data?.relation || []) {
    for (const id of [edge.source_id, edge.target_id]) degree.set(id, (degree.get(id) || 0) + 1);
  }
  return [...(data.value?.graph_data?.node || [])].sort((a: any,b: any) => (degree.get(b.id) || 0) - (degree.get(a.id) || 0)).slice(0, 40).map((node: any, n: number, array: any[]) => ({ ...node, x: 350 + 260 * Math.cos(2 * Math.PI * n / array.length), y: 290 + 230 * Math.sin(2 * Math.PI * n / array.length) }));
});
const graphEdges = computed(() => (data.value?.graph_data?.relation || []).map((edge: any) => ({ ...edge, source: graphNodes.value.find((n: any) => n.id === edge.source_id), target: graphNodes.value.find((n: any) => n.id === edge.target_id) })).filter((edge: any) => edge.source && edge.target));
const stateLabel = computed(() => (({ complete: '已完成', partial: '部分完成，有失败阶段', processing: '处理中' } as Record<string,string>)[data.value?.status as string] || '读取中'));
</script>

<template>
  <section class="archive-panel">
    <header>
      <strong>档案 {{ data?.archive?.title }} · {{ stateLabel }}</strong>
      <span v-if="data">{{ data.archive?.page_count || '—' }} 页 · {{ data.chunks }} 个检索切块</span>
      <button @click="load">刷新</button>
      <button v-if="canEdit" :disabled="busy" @click="retry">{{ busy ? '提交中…' : '重试失败阶段' }}</button>
    </header>
    <p v-if="error" role="alert">{{ error }}</p>
    <p v-else-if="!data">正在读取档案…</p>
    <template v-else-if="view === 'wiki'">
      <p>以下 Wiki 内容均引用当前档案；档案概览优先显示。</p>
      <p v-if="!pages.length">当前档案尚未生成 Wiki，请查看处理记录。</p>
      <details v-for="page in pages" :key="page.slug" :open="page.page_type === 'summary'">
        <summary>{{ page.page_type === 'summary' ? '档案概览' : page.title }}</summary>
        <article class="markdown-body" v-html="renderWiki(page.content || page.summary || '')"></article>
        <button v-for="chunkId in (page.chunk_refs || [])" :key="chunkId" @click="emit('locate-chunk', chunkId)">查看档案证据</button>
      </details>
    </template>
    <template v-else-if="view === 'graph'">
      <p>当前档案图谱 · 已完成 {{ data.graph?.done }} 段 · 失败 {{ data.graph?.failed }} 段 · 进行中 {{ data.graph?.pending }} 段</p>
      <p v-if="data.graph_error">图谱暂时无法读取，请稍后刷新。</p>
      <p v-else-if="!graphNodes.length">当前档案暂无已抽取实体；请查看处理记录。</p>
      <svg v-else viewBox="0 0 700 580" role="img" aria-label="当前档案的实体关系图">
        <g v-for="(edge, index) in graphEdges" :key="index">
          <line :x1="edge.source.x" :y1="edge.source.y" :x2="edge.target.x" :y2="edge.target.y" stroke="#b8c6c0" />
          <text :x="(edge.source.x+edge.target.x)/2" :y="(edge.source.y+edge.target.y)/2" text-anchor="middle" font-size="10">{{ edge.type }}</text>
        </g>
        <g v-for="node in graphNodes" :key="node.id">
          <circle :cx="node.x" :cy="node.y" r="9" fill="#07a870"><title>{{ node.name }}</title></circle>
          <text :x="node.x" :y="node.y-15" text-anchor="middle" font-size="11">{{ node.name.length > 14 ? node.name.slice(0,14)+'…' : node.name }}</text>
        </g>
      </svg>
      <p v-if="(data.graph_data?.node?.length || 0)>40">当前显示前 40 个实体。</p>
      <details v-for="node in graphNodes" :key="node.id"><summary>{{ node.name }}</summary>
        <button v-for="id in (node.chunks || [])" :key="id" @click="emit('locate-chunk', id)">打开档案中的证据</button>
      </details>
    </template>
    <template v-else>
      <p>重试只补做失败阶段，已有档案、切块和成功索引会保留。</p>
      <ul><li v-for="stage in data.stages" :key="stage.name">{{ stage.name }}：{{ stage.status }}</li></ul>
    </template>
  </section>
</template>

<style scoped>
.archive-panel { padding: 16px; min-height: 350px; }
header { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; padding-bottom: 14px; border-bottom: 1px solid #eee; }
button { border: 1px solid #ddd; border-radius: 6px; padding: 6px 12px; background: white; cursor: pointer; }
details { margin-top: 16px; border: 1px solid #eee; padding: 12px; border-radius: 8px; }
summary { cursor: pointer; font-weight: 600; }
article { padding: 12px 0; }
svg { width: 100%; max-height: 650px; }
</style>
