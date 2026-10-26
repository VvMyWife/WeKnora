<script setup lang="ts">
import { ref, watch } from 'vue';
import { listKnowledgeFiles } from '@/api/knowledge-base';
import ArchiveDocumentPanel from './ArchiveDocumentPanel.vue';
const props = defineProps<{ knowledgeBaseId: string; view: 'wiki' | 'graph'; canEdit?: boolean }>();
const emit = defineEmits<{ (e: 'open-source-doc', id: string, chunkId?: string): void }>();
const rows = ref<any[]>([]);
const selected = ref<any>(null);
const page = ref(1);
const total = ref(0);
const keyword = ref('');
const error = ref('');
let epoch = 0;
async function load() {
  const current = ++epoch;
  try {
    const response: any = await listKnowledgeFiles(props.knowledgeBaseId, { page: page.value, page_size: 30, keyword: keyword.value, sort_by: 'file_name', sort_order: 'asc' });
    if (current !== epoch) return;
    rows.value = response.data || [];
    total.value = response.total || 0;
    error.value = '';
  } catch { if (current === epoch) error.value = '档案列表读取失败，请重试'; }
}
watch(() => props.knowledgeBaseId, () => { selected.value = null; page.value = 1; void load(); }, { immediate: true });
function search() { page.value = 1; void load(); }
</script>
<template>
  <section class="archive-library">
    <template v-if="!selected">
      <header><strong>按档案查看{{ view === 'wiki' ? ' Wiki' : '图谱' }}</strong><span>共 {{ total }} 份文档</span></header>
      <form @submit.prevent="search"><input v-model="keyword" aria-label="搜索档案名称" placeholder="搜索档案名称" /><button>搜索</button><button type="button" @click="load">刷新</button></form>
      <p v-if="error" role="alert">{{ error }}</p>
      <p v-if="!rows.length && !error">暂无档案。</p>
      <div class="archive-grid">
        <button v-for="row in rows" :key="row.id" class="archive-card" @click="selected = row">
          <strong>{{ row.title || row.file_name }}</strong><span>{{ row.folder_path || '知识库根目录' }}</span>
          <span v-if="row.custom_metadata?.['档案页数']">{{ row.custom_metadata['档案页数'] }} 页 · {{ row.custom_metadata['检索切块数'] }} 个检索切块</span>
          <span>{{ row.parse_status === 'completed' ? '已完成' : row.parse_status === 'failed' ? '存在失败阶段' : '处理中' }}</span>
        </button>
      </div>
      <footer><button :disabled="page <= 1" @click="page--; load()">上一页</button><span>第 {{ page }} 页</span><button :disabled="page * 30 >= total" @click="page++; load()">下一页</button></footer>
    </template>
    <template v-else>
      <header><button @click="selected = null">返回档案列表</button><strong>{{ selected.title || selected.file_name }}</strong><button @click="emit('open-source-doc', selected.id)">完整档案与切块</button></header>
      <ArchiveDocumentPanel :key="selected.id" :knowledge-id="selected.id" :view="view" :can-edit="canEdit" @locate-chunk="(chunkId) => emit('open-source-doc', selected.id, chunkId)" />
    </template>
  </section>
</template>
<style scoped>
.archive-library { padding: 20px; overflow: auto; height: 100%; }
header, form, footer { display: flex; align-items: center; flex-wrap: wrap; gap: 12px; margin-bottom: 18px; }
input, button { padding: 8px 12px; border: 1px solid #ddd; border-radius: 6px; background: white; }
button { cursor: pointer; } button:disabled { cursor: default; opacity: .5; }
.archive-grid { display: grid; grid-template-columns: repeat(auto-fill,minmax(210px,1fr)); gap: 14px; margin-bottom: 18px; }
.archive-card { display: flex; flex-direction: column; align-items: start; gap: 10px; text-align: left; padding: 18px; }
.archive-card:hover { border-color: #07a870; }
.archive-card span { color: #666; font-size: 12px; overflow-wrap: anywhere; }
</style>
