<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { listArchiveBatches, getArchiveBatch, createArchiveBatch, uploadArchiveItem, skipArchiveItem, getArchiveSource, setArchiveSource, type ArchiveBatch, type ArchiveItem, type ArchiveTarget, type ArchiveVersion } from '@/api/archive-mapping';
import { relativeOriginalPath, commonParsedRoot, archiveCounts, runArchiveUploads } from '../archiveMapping';
const props=defineProps<{visible:boolean; kbId:string}>();
const emit=defineEmits<{(e:'update:visible',v:boolean):void;(e:'changed'):void}>();
const batches=ref<ArchiveBatch[]>([]), items=ref<ArchiveItem[]>([]), targets=ref<ArchiveTarget[]>([]);
const batchId=ref(''), parsedRoot=ref(''), files=ref(new Map<string,File>()), replace=ref<Record<string,boolean>>({});
const busy=ref(false), uploading=ref(false), stopping=ref(false), filter=ref(''), page=ref(1), error=ref('');
const versionKnowledge=ref(''), versions=ref<ArchiveVersion[]>([]), activeVersion=ref(''), storageBytes=ref(0);
const selectedCount=computed(()=>files.value.size), counts=computed(()=>archiveCounts(items.value));
const filtered=computed(()=>items.value.filter(i=>!filter.value || (i.relative_path+' '+i.error).includes(filter.value)));
const rows=computed(()=>filtered.value.slice((page.value-1)*25,page.value*25));
const targetOptions=computed(()=>targets.value.map(t=>({label:t.path,value:t.id})));
const statusNames:Record<string,string>={ready:'待上传',unmatched:'未匹配',ambiguous:'多目标',conflict:'已有原件',bound:'已关联',skipped:'已跳过',failed:'失败'};
let epoch=0;
const remember=()=>localStorage.setItem(`archive-mapping:${props.kbId}`,batchId.value);
async function refreshBatch() {
  if(!batchId.value)return;
  const kb=props.kbId, id=batchId.value;const r=await getArchiveBatch(kb,id);
  if(kb===props.kbId && id===batchId.value){items.value=r.data.items;parsedRoot.value=r.data.batch.parsed_root;remember();}
}
async function open(){
  const token=++epoch,kb=props.kbId;busy.value=true;error.value='';
  try{
    const r=await listArchiveBatches(kb);if(token!==epoch)return;
    batches.value=r.data.batches;targets.value=r.data.targets;parsedRoot.value=commonParsedRoot(targets.value.map(t=>t.path));
    const saved=localStorage.getItem(`archive-mapping:${kb}`);batchId.value=batches.value.find(b=>b.id===saved)?.id || batches.value[0]?.id || '';
    if(batchId.value)await refreshBatch();else items.value=[];
  }catch(e:any){if(token===epoch)error.value=e.message || '无法读取映射批次';}finally{if(token===epoch)busy.value=false;}
}
watch(()=>[props.visible,props.kbId],()=>{if(props.visible)void open();else epoch++;},{immediate:true});
watch(filter,()=>page.value=1);
function pick(event:Event){
  const input=event.target as HTMLInputElement;const next=new Map<string,File>();
  for(const f of Array.from(input.files || []))next.set(relativeOriginalPath(f),f);
  files.value=next;input.value='';
}
async function plan(){
  busy.value=true;error.value='';
  try{const r=await createArchiveBatch(props.kbId,parsedRoot.value,Array.from(files.value,([relative_path,f])=>({relative_path,file_size:f.size})));batchId.value=r.data.batch.id;items.value=r.data.items;batches.value.unshift(r.data.batch);page.value=1;remember();}
  catch(e:any){error.value=e.message;}finally{busy.value=false;}
}
async function send(item:ArchiveItem){
  const f=files.value.get(item.relative_path);if(!f){item.error='请重新选择包含此相对路径的原件目录';return;}
  try{await uploadArchiveItem(props.kbId,batchId.value,item,f,!!replace.value[item.id]);item.status='bound';item.error='';}
  catch(e:any){item.status='failed';item.error=e.message || '上传失败';}
}
async function upload(all:boolean,item?:ArchiveItem){
  uploading.value=true;stopping.value=false;
  try{
    const queue=all ? items.value.filter(i=>i.status!=='bound' && i.status!=='skipped' && i.knowledge_id && files.value.has(i.relative_path) && (i.status!=='conflict' || replace.value[i.id])) : item ? [item] : [];
    await runArchiveUploads(queue,send,()=>stopping.value);await refreshBatch();emit('changed');
  }catch(e:any){error.value=e.message;}finally{uploading.value=false;}
}
async function skip(item:ArchiveItem){try{await skipArchiveItem(props.kbId,batchId.value,item.id);item.status='skipped';item.error='用户跳过';}catch(e:any){error.value=e.message;}}
async function loadVersions(){
  if(!versionKnowledge.value)return;const id=versionKnowledge.value;
  try{const r=await getArchiveSource(id);if(id===versionKnowledge.value){versions.value=r.data.versions;activeVersion.value=r.data.active?.id || '';storageBytes.value=r.data.storage_bytes;}}
  catch(e:any){error.value=e.message;}
}
async function activate(id:string){busy.value=true;try{await setArchiveSource(versionKnowledge.value,id);await loadVersions();emit('changed');}catch(e:any){error.value=e.message;}finally{busy.value=false;}}
function close(){if(uploading.value){stopping.value=true;MessagePlugin.info('正在停止发送新文件，等待当前上传完成后即可关闭');return;}emit('update:visible',false);}
</script>
<template>
  <t-dialog :visible="visible" header="档案映射" width="min(1180px, 96vw)" :footer="false" :close-on-overlay-click="false" @close="close">
    <section class="archive-mapping" :aria-busy="busy || uploading">
      <p>原件仅存储，不进行 OCR、切块或模型处理。检索仍使用现有 MD；预览与引用优先打开关联原件。按相对目录和同名文件匹配，不含最外层目录名。</p>
      <p v-if="error" role="alert" class="error">{{ error }}</p>
      <div class="toolbar">
        <label>选择原件目录 <input type="file" webkitdirectory multiple :disabled="busy || uploading" @change="pick" /></label>
        <span>已选择 {{ selectedCount }} 个文件（刷新后需重新选择目录继续）</span>
      </div>
      <div class="toolbar">
        <label>MD 匹配根目录 <t-input v-model="parsedRoot" placeholder="如 政策标准库_MD；留空表示知识库根目录" :disabled="busy || uploading" /></label>
        <t-button :disabled="!selectedCount || busy || uploading" @click="plan">新建匹配预览</t-button>
        <t-select v-model="batchId" :disabled="busy || uploading" placeholder="历史批次" :options="batches.map(b=>({label:b.created_at+' · '+b.id.slice(0,8),value:b.id}))" @change="refreshBatch" />
      </div>
      <template v-if="batchId">
        <div class="toolbar">
          <strong>共 {{ counts.total }} · 已关联 {{ counts.bound }} · 跳过 {{ counts.skipped }} · 失败 {{ counts.failed }} · 待处理 {{ counts.pending }}</strong>
          <t-button :disabled="uploading || busy || !selectedCount" @click="upload(true)">上传已匹配原件</t-button>
          <t-button v-if="uploading" variant="outline" @click="stopping=true">停止继续发送</t-button>
          <t-button variant="text" :disabled="uploading" @click="refreshBatch">刷新批次</t-button>
        </div>
        <progress :value="counts.bound+counts.skipped" :max="counts.total || 1" aria-label="档案映射进度" />
        <t-input v-model="filter" placeholder="按相对路径或失败原因筛选" clearable />
        <div class="items">
          <article v-for="item in rows" :key="item.id">
            <div><strong>{{ item.relative_path }}</strong> <small>{{ statusNames[item.status] || item.status }}</small><p v-if="item.error" class="error">{{ item.error }}</p></div>
            <t-select v-model="item.knowledge_id" filterable :disabled="uploading || item.status==='bound'" :options="targetOptions" placeholder="选择当前知识库目标 MD" />
            <label v-if="item.status!=='bound'"><input v-model="replace[item.id]" type="checkbox" :disabled="uploading" />允许替换已有原件（保留旧版本）</label>
            <div v-if="item.status!=='bound'" class="toolbar"><t-button size="small" :disabled="uploading || busy || !item.knowledge_id || !files.has(item.relative_path)" @click="upload(false,item)">上传并关联</t-button><t-button size="small" variant="text" :disabled="uploading" @click="skip(item)">跳过</t-button></div>
          </article>
        </div>
        <t-pagination v-model="page" :total="filtered.length" :page-size="25" :show-page-size="false" />
      </template>
      <details><summary>原件版本与回退</summary>
        <t-select v-model="versionKnowledge" filterable :options="targetOptions" placeholder="选择目标 MD 查看原件版本" @change="loadVersions" />
        <p>保留版本占用 {{ (storageBytes/1024/1024).toFixed(2) }} MB；解除关联不会删除原件。删除知识条目时清理原件。</p>
        <div v-for="v in versions" :key="v.id" class="toolbar"><span>{{ v.file_name }} · {{ v.created_at }} · {{ v.sha256.slice(0,12) }} {{ v.id===activeVersion ? '（当前）' : '' }}</span><t-button size="small" :disabled="busy || uploading || v.id===activeVersion" @click="activate(v.id)">设为当前原件</t-button></div>
        <t-button v-if="activeVersion" variant="outline" :disabled="busy || uploading" @click="activate('')">解除关联，预览恢复 MD</t-button>
      </details>
    </section>
  </t-dialog>
</template>
<style scoped>
.archive-mapping{max-height:76vh;overflow:auto;padding:0 8px 12px;line-height:1.6}.toolbar{display:flex;gap:12px;align-items:center;flex-wrap:wrap;margin:12px 0}.toolbar>label{min-width:300px}.toolbar>.t-select__wrap{width:340px}.items article{display:grid;grid-template-columns: minmax(0,1fr) minmax(220px,1fr);gap:8px 16px;padding:12px 0;border-bottom:1px solid var(--td-component-border)}.items strong{overflow-wrap:anywhere}.items small{margin-left:8px}.error{color:var(--td-error-color);white-space:pre-wrap;overflow-wrap:anywhere}progress{width:100%;accent-color:var(--td-brand-color)}details{margin-top:20px}summary{cursor:pointer;font-weight:600}
</style>
