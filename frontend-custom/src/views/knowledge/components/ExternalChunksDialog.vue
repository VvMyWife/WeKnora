<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { get, getDown, post, postUpload } from '@/utils/request';
import { importCounts, type ImportRecord } from '../externalChunks';
const props = defineProps<{visible:boolean;kbId:string}>();
const emit = defineEmits(['update:visible','changed']);
const batchId=ref(''), batches=ref<{id:string;created_at:string}[]>([]), rows=ref<ImportRecord[]>([]);
const busy=ref(false), uploading=ref(false), running=ref(false), stopping=ref(false), error=ref(''), transferred=ref(0), filter=ref('');
const base=computed(()=>`/api/v1/knowledge-bases/${encodeURIComponent(props.kbId)}/knowledge/external-chunks/batches`);
const counts=computed(()=>importCounts(rows.value));
const filtered=computed(()=>rows.value.filter(r=>(r.relative_path+' '+r.error).includes(filter.value)));
const page=ref(1), visibleRows=computed(()=>filtered.value.slice((page.value-1)*25,page.value*25));
let timer:ReturnType<typeof setInterval>|undefined;
async function refresh(){
 if(!batchId.value)return;
 const result=await get<{data:{documents:ImportRecord[]}}>(`${base.value}/${batchId.value}`);
 rows.value=result.data.documents;
}
async function load(){
 try{const r=await get<{data:{id:string;created_at:string}[]}>(base.value);batches.value=r.data;if(!batchId.value&&r.data.length)batchId.value=r.data[0].id;await refresh();}
 catch(e:any){error.value=e.message;}
}
watch(()=>props.visible,async visible=>{
 if(timer)clearInterval(timer);
 if(visible){await load();timer=setInterval(()=>{if(!busy.value&&!uploading.value)void refresh().catch(e=>{error.value=e.message;});},4000);}
});
onBeforeUnmount(()=>{if(timer)clearInterval(timer);stopping.value=true;});
watch(filter,()=>{page.value=1;});
async function pick(event:Event){
 const input=event.target as HTMLInputElement,file=input.files?.[0];input.value='';if(!file)return;
 uploading.value=true;error.value='';transferred.value=0;
 try{const data=new FormData();data.append('file',file);const r=await postUpload(base.value,data,(p:any)=>{if(p.total)transferred.value=Math.round(p.loaded/p.total*100);},{timeout:0});batchId.value=r.data.batch.id;rows.value=r.data.documents;await load();}
 catch(e:any){error.value=e.message;}finally{uploading.value=false;}
}
async function start(failedOnly=false){
 if(running.value)return;running.value=true;stopping.value=false;error.value='';
 const queue=rows.value.filter(r=>r.status==='failed'||(!failedOnly&&r.status==='preview'));
 let cursor=0;
 try{await Promise.all([0,1].map(async()=>{while(!stopping.value&&cursor<queue.length){const item=queue[cursor++];try{await post(`${base.value}/${batchId.value}/documents/${item.id}/start`,{},{timeout:0});}catch(e:any){error.value=e.message;}await refresh();emit('changed');}}));}
 finally{running.value=false;await refresh();emit('changed');}
}
async function downloadMap(){
 try{const blob=await getDown(`${base.value}/${batchId.value}/id-map`);const url=URL.createObjectURL(blob);const a=document.createElement('a');a.href=url;a.download='external-chunks-id-map.json';a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);}
 catch(e:any){error.value=e.message;}
}
function close(){if(running.value||uploading.value){stopping.value=true;error.value='请等待当前上传/提交收尾；已停止发送后续文档。';return;}emit('update:visible',false);}
</script>
<template>
 <t-dialog :visible="visible" header="导入外部分块" width="min(1180px,96vw)" :footer="false" :close-on-overlay-click="false" @close="close">
  <section class="external-chunks">
   <p>上传原 RAG 导出的 weknora-chunks-v1.zip。先预检，再手动执行；直接使用包内分块，不再 OCR、重切或生成图片描述。原件自动接入档案映射。</p>
   <p v-if="error" role="alert" class="error">{{error}}</p>
   <div class="toolbar"><label>选择导出 ZIP <input type="file" accept=".zip" :disabled="uploading||running||busy" @change="pick" /></label><span v-if="uploading">上传 {{transferred}}% · 上传后校验包内数据</span></div>
   <div class="toolbar"><t-select v-model="batchId" :options="batches.map(b=>({label:b.created_at+' · '+b.id.slice(0,8),value:b.id}))" :disabled="running||uploading" placeholder="历史导入批次" @change="refresh" /><t-button variant="text" :disabled="uploading" @click="load">刷新批次</t-button></div>
   <template v-if="batchId">
    <p>共 {{counts.total}} · 全流程完成 {{counts.completed}} · 待导入 {{counts.waiting}} · 处理中 {{counts.active}} · 失败 {{counts.failed}}</p>
    <progress :value="counts.completed" :max="counts.total||1" aria-label="外部分块导入进度" />
    <div class="toolbar"><t-button :disabled="running||uploading||!counts.waiting" @click="start()">确认导入</t-button><t-button :disabled="running||uploading||!counts.failed" variant="outline" @click="start(true)">重试失败文档</t-button><t-button v-if="running" variant="outline" @click="stopping=true">停止继续提交</t-button><t-button variant="text" @click="downloadMap">下载 ID 对照</t-button></div>
    <p>停止提交不会中断已进入后台的任务；后台暂停/恢复使用知识库全库进度面板。刷新或重新打开本窗口可继续，不必重新上传 ZIP。</p>
    <t-input v-model="filter" placeholder="筛选路径或失败原因" clearable />
    <article v-for="item in visibleRows" :key="item.id"><strong>{{item.relative_path}}</strong><span>{{item.status}} · {{item.stage}}</span><small>knowledge_id: {{item.knowledge_id}}</small><p v-if="item.error" class="error">{{item.error}}</p></article>
    <t-pagination v-model="page" :total="filtered.length" :page-size="25" :show-page-size="false" />
   </template>
  </section>
 </t-dialog>
</template>
<style scoped>
.external-chunks{max-height:76vh;overflow:auto;padding:0 8px 12px;line-height:1.6}.toolbar{display:flex;align-items:center;gap:12px;flex-wrap:wrap;margin:12px 0}.toolbar>.t-select__wrap{width:420px}article{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:8px;padding:12px 0;border-bottom:1px solid var(--td-component-border)}article small,article p{grid-column:1/-1}article strong,article p{overflow-wrap:anywhere}.error{color:var(--td-error-color);white-space:pre-wrap}progress{width:100%;accent-color:var(--td-brand-color)}
</style>
