<script setup lang="ts">
import {computed,onBeforeUnmount,ref,watch} from 'vue'
import {getCitationEvidence,getCitationManifest,getCitationOrigin,type CitationDetail} from '@/api/archival-citations'
import {citationPageLabel,citationSourceGroups,type CitationManifest,type CitationSourceGroup} from '@/utils/archivalCitations'
import {safeMarkdownToHTML} from '@/utils/security'
import {copyToClipboard} from '@/utils/clipboard'
import {MessagePlugin} from 'tdesign-vue-next'
import {defineAsyncComponent} from 'vue'
const ArchivalPdfViewer=defineAsyncComponent(()=>import('./ArchivalPdfViewer.vue'))

const props=defineProps<{sessionId:string;messageId:string;complete?:boolean;manifest?:CitationManifest}>()
const enabled=ref(false),manifest=ref<CitationManifest|null>(null),unavailable=ref(0)
const visible=ref(false),loading=ref(false),error=ref(''),detail=ref<CitationDetail|null>(null)
const evidenceElement=ref<HTMLElement|null>(null)
const showPdf=ref(false)
const preferredPage=ref<number|undefined>()
const groups=computed(()=>citationSourceGroups(manifest.value))
const originText=ref(''),originError=ref(''),originLoading=ref(false)
const renderedOrigin=computed(()=>safeMarkdownToHTML(originText.value))
let originRevision=0,manifestRevision=0
let originAbort:AbortController|undefined
let manifestAbort:AbortController|undefined,detailAbort:AbortController|undefined,revision=0
const rendered=computed(()=>detail.value?safeMarkdownToHTML(detail.value.snapshot):'')
const status=computed(()=>!props.complete?'来源绑定校验中':manifest.value?.state==='validation_failed'?'来源校验或证据保存失败，请勿将本回答视为已验证引用':manifest.value?.state==='legacy'?'历史回答：没有版本化引用记录':manifest.value?.state==='answer_changed'?'回答已修改，旧引用不能代表当前回答':manifest.value?.state==='needs_review'?'部分引用待核验':'来源绑定已校验（不代表事实已自动核实）')
async function refresh(){
 manifestAbort?.abort();manifestAbort=new AbortController()
 const ticket=++manifestRevision
 if(!props.sessionId||!props.messageId)return
 try{
  const result=await getCitationManifest(props.sessionId,props.messageId,manifestAbort.signal)
  if(ticket!==manifestRevision)return
  enabled.value=result.enabled!==false;manifest.value=result.data;unavailable.value=result.unavailable_sources||0
  if(props.manifest?.state==='validation_failed'&&result.data?.state==='legacy')manifest.value=props.manifest
 }catch{if(ticket===manifestRevision){enabled.value=!!props.manifest;manifest.value=props.manifest||null}}
}
watch(()=>[props.sessionId,props.messageId,props.complete,props.manifest],()=>{
 if(props.manifest){enabled.value=true;manifest.value=props.manifest}
 void refresh()
},{immediate:true})
watch(()=>[props.sessionId,props.messageId],()=>{revision++;originRevision++;originAbort?.abort();detailAbort?.abort();visible.value=false;detail.value=null;error.value=''})
async function openCitation(id:string,page?:number){
 preferredPage.value=page
 visible.value=true;loading.value=true;error.value='';detail.value=null;showPdf.value=false;originText.value='';originError.value='';originRevision++
 originAbort?.abort();originLoading.value=false
 detailAbort?.abort();detailAbort=new AbortController();const current=++revision
 try{const result=await getCitationEvidence(props.sessionId,props.messageId,id,detailAbort.signal);if(current===revision){detail.value=result.data;showPdf.value=page!==undefined}}
 catch{if(current===revision)error.value='证据不可读取：来源可能已删除、权限已撤销或引用已失效。不会改用其他文件。'}
 finally{if(current===revision)loading.value=false}
}
function openPage(group:CitationSourceGroup,page:number){const c=group.citations.find(c=>c.location.status==='known'&&c.location.pages.includes(page));if(c)void openCitation(c.id,page)}
async function openOrigin(knowledge:string,chunk:string){
 const d=detail.value;if(!d)return;const ticket=++originRevision;originText.value='';originError.value='';originLoading.value=true
 originAbort?.abort();originAbort=new AbortController()
 try{const result=await getCitationOrigin(props.sessionId,props.messageId,d.id,knowledge,chunk,originAbort.signal);if(ticket===originRevision)originText.value=result.data.content}
 catch{if(ticket===originRevision)originError.value='该来源关系已失效、分块已删除或权限已撤销'}
 finally{if(ticket===originRevision)originLoading.value=false}
}
function close(){visible.value=false;revision++;originRevision++;originAbort?.abort();detailAbort?.abort();detail.value=null;showPdf.value=false}
async function copy(text:string){if(await copyToClipboard(text))MessagePlugin.success('已复制');else MessagePlugin.error('复制失败，请手动选择原文复制')}
function copyOriginal(){if(detail.value)void copy(detail.value.snapshot)}
function copyPlain(){if(evidenceElement.value)void copy(evidenceElement.value.innerText)}
function copyReference(){const d=detail.value;if(!d)return;void copy(`[${d.number}] ${d.title}\n${d.wiki_page_id?'Wiki Page':'chunk'}: ${d.chunk_id}\n内容版本: ${d.content_revision<0?'未确认':d.content_revision}\nSHA256: ${d.content_sha256}\n原件版本: ${d.archive_version_id||'未绑定'}\n${citationPageLabel(d.location)}`)}
onBeforeUnmount(()=>{revision++;originRevision++;manifestRevision++;originAbort?.abort();manifestAbort?.abort();detailAbort?.abort()})
defineExpose({openCitation})
</script>
<template>
 <section v-if="enabled && manifest?.state!=='not_applicable'" class="archival-citations" aria-label="回答来源校验">
  <div class="archival-status" role="status">{{status}} <span v-if="unavailable"> · {{unavailable}} 个来源当前不可访问</span></div>
  <details v-if="manifest?.unbound_blocks?.length" class="archival-warning">
   <summary>{{manifest.unbound_blocks.length}} 处引用待核验；不代表知识库没有来源。查看原因</summary>
   <ul><li v-for="(block,index) in manifest.unbound_blocks" :key="index">正文位置 {{block.index+1}}{{block.end_index!==undefined && block.end_index>block.index?`—${block.end_index+1}`:''}}：{{block.reason_code?block.reason:'历史回答未建立有效引用绑定；未自动重新校验或改写'}} <code v-if="block.reason_code">{{block.reason_code}}</code></li></ul>
  </details>
  <div v-if="manifest?.citations?.length" class="archival-footnotes" aria-label="引用脚注">
   <button v-for="c in manifest.citations" :key="c.id" type="button" @click="openCitation(c.id)">[{{c.number}}] {{c.title}}</button>
  </div>
  <details v-if="groups.length" class="archival-directory">
   <summary>答案来源目录 · {{groups.length}} 个来源版本 · {{manifest?.citations.length}} 个证据片段</summary>
   <table><thead><tr><th>文件／原件版本</th><th>实际引用页码</th><th>脚注与分块</th></tr></thead>
    <tbody><tr v-for="g in groups" :key="g.key">
     <td>{{g.title}}<div class="archival-source-version">{{g.derived?'派生来源，未验证原始页':g.version?`原件版本：${g.version}`:'未提供原件版本'}}</div></td>
     <td><button v-for="p in g.pages" :key="p" type="button" @click="openPage(g,p)">第 {{p}} 页</button><div v-if="g.withoutPages.length">未提供页码：脚注 {{g.withoutPages.join('、')}}</div></td>
     <td><button v-for="c in g.citations" :key="c.id" type="button" @click="openCitation(c.id)">[{{c.number}}] {{c.wiki_page_id?'Wiki 页面':'分块'}} {{c.chunk_id}} · 内容版本 {{c.content_revision<0?'未确认':c.content_revision}}</button></td>
    </tr></tbody>
   </table>
  </details>
 </section>
 <t-drawer :visible="visible" attach="body" :footer="false" size="min(860px, 94vw)" header="引用证据" @close="close">
  <nav v-if="manifest?.citations?.length" class="archival-actions" aria-label="切换引用证据"><button v-for="c in manifest.citations" :key="c.id" type="button" @click="openCitation(c.id)">脚注 {{c.number}}</button></nav>
  <div v-if="loading" role="status">正在读取引用时的证据快照…</div>
  <div v-else-if="error" role="alert" class="archival-warning">{{error}}</div>
  <div v-else-if="detail" class="archival-evidence">
   <h3>[{{detail.number}}] {{detail.title}}</h3>
   <div>{{citationPageLabel(detail.location)}}</div>
   <div>内容版本：{{detail.content_revision<0?'未确认':detail.content_revision}} · {{detail.snapshot_at}}</div>
   <div v-if="detail.source_kind==='derived'" class="archival-warning">Wiki／图谱等派生证据，不能冒充 PDF 直接证据。</div>
   <div v-if="detail.source_kind==='excerpt'" class="archival-warning">模型获得的是摘录，以下展示该次提供的证据文本，内容版本未确认。</div>
   <div v-if="detail.current_changed" class="archival-warning">当前内容与引用文本不完全相同。下方保留引用时的快照，不会悄悄换成新版。</div>
   <div class="archival-actions">
    <button type="button" @click="copyOriginal">复制原文</button>
    <button type="button" @click="copyPlain">复制纯文本／表格</button>
    <button type="button" @click="copyReference">复制引用信息</button>
    <button v-if="detail.archive_version_id && detail.source_kind!=='derived'" type="button" @click="showPdf=!showPdf">{{showPdf?'收起原件':'预览对应原件页面'}}</button>
   </div>
   <div ref="evidenceElement" class="archival-evidence-text markdown-content" v-html="rendered" />
   <details v-if="detail.origins?.length"><summary>派生来源关系（不等同于直接原文证据）</summary>
    <ul><li v-for="origin in detail.origins" :key="origin.knowledge_id">{{origin.title}} · 文档 {{origin.knowledge_id}}
     <div><button v-for="chunk in origin.chunk_ids" :key="chunk" type="button" @click="openOrigin(origin.knowledge_id,chunk)">查看当前来源分块 {{chunk}}</button></div>
    </li></ul>
    <div v-if="originLoading" role="status">读取当前来源分块…</div><div v-if="originError" role="alert">{{originError}}</div>
    <div v-if="originText"><strong>以下是当前来源内容，不是引用时的证据快照</strong><button type="button" @click="copy(originText)">复制当前来源</button><div class="archival-evidence-text" v-html="renderedOrigin" /></div>
   </details>
   <div v-if="!detail.archive_version_id" class="archival-warning">引用时未绑定原件版本，不能自动使用后来上传的文件。</div>
   <ArchivalPdfViewer v-if="showPdf && detail.archive_version_id" :key="detail.id" :knowledge-id="detail.knowledge_id" :archive-version-id="detail.archive_version_id" :location="detail.location" :initial-page="preferredPage" />
  </div>
 </t-drawer>
</template>
<style>
.archival-citations{font-size:12px;color:#72756f;margin:16px 0;padding:12px 0;border-top:1px solid #e8e6df}
.archival-status{line-height:1.7}.archival-warning,.citation-unbound{color:#a85d1b;background:#fff7e9;padding:3px 5px;border-radius:4px}
.citation-unbound{font-size:11px;white-space:nowrap}
.archival-footnotes,.archival-actions{display:flex;gap:8px;flex-wrap:wrap;margin:12px 0}
.archival-footnotes button,.archival-actions button{cursor:pointer;border:1px solid #dedbd3;border-radius:5px;padding:5px 9px;background:#faf9f5;color:#3e5545}
.archival-evidence{font-size:13px;line-height:1.8}.archival-evidence-text{overflow-wrap:anywhere;overflow-x:auto;margin-top:16px;border-top:1px solid #eee;padding-top:12px}
.archival-evidence-text table{border-collapse:collapse;max-width:100%;display:block;overflow:auto}.archival-evidence-text th,.archival-evidence-text td{border:1px solid #ddd;padding:6px 10px;min-width:70px}
.citation-archival{font-size:11px!important;vertical-align:super;cursor:pointer;color:#247445!important;padding:0 3px!important;min-width:0!important}
.citation-archival:focus-visible{outline:2px solid #247445;outline-offset:2px}
.archival-directory{margin-top:10px;overflow-x:auto}.archival-directory summary{cursor:pointer;color:#34573b}.archival-directory table{width:100%;border-collapse:collapse;margin-top:10px}.archival-directory th,.archival-directory td{padding:8px;border:1px solid #e5e2db;text-align:left;vertical-align:top;overflow-wrap:anywhere}.archival-directory button{display:inline-block;margin:2px;cursor:pointer;background:#fff;border:1px solid #ddd;border-radius:3px;padding:3px 6px;color:#346548}.archival-source-version{font-size:11px;color:#888;margin-top:4px}
</style>
