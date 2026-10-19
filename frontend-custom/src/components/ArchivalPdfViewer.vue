<script setup lang="ts">
import {computed,nextTick,onBeforeUnmount,onMounted,ref,shallowRef,watch} from 'vue'
import {getDocument,GlobalWorkerOptions,version,type PDFDocumentProxy,type PDFDocumentLoadingTask,type RenderTask} from 'pdfjs-dist'
import workerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url'
import {get} from '@/utils/request'
import {citationPageTarget,citationPageLabel,type CitationLocation} from '@/utils/archivalCitations'

GlobalWorkerOptions.workerSrc=workerUrl
const props=defineProps<{knowledgeId:string;archiveVersionId:string;location:CitationLocation;initialPage?:number}>()
const root=ref<HTMLElement|null>(null),loading=ref(false),error=ref(''),warning=ref('')
const pdf=shallowRef<PDFDocumentProxy|null>(null),current=ref(1),inputPage=ref(1),zoom=ref(1),fullscreen=ref(false)
const pageHeight=ref(840),pageWidth=ref(594),pages=ref<number[]>([])
const sourceLabel=computed(()=>citationPageLabel(props.location))
let task:PDFDocumentLoadingTask|undefined,abort:AbortController|undefined,observer:IntersectionObserver|undefined,resize:ResizeObserver|undefined
let epoch=0,renderEpoch=0,resizeEpoch=0,frame=0,bodyOverflow=''
const renders=new Map<number,RenderTask>(),rendered=new Set<number>(),nearby=new Set<number>()
function element(n:number){return root.value?.querySelector<HTMLElement>(`[data-pdf-page="${n}"]`)||null}
function release(n:number){renders.get(n)?.cancel();renders.delete(n);rendered.delete(n);const host=element(n);if(host)delete host.dataset.pdfRendered;const canvas=host?.querySelector('canvas');if(canvas){canvas.width=0;canvas.height=0;canvas.remove()}}
function resetRenders(){renderEpoch++;for(const n of new Set([...renders.keys(),...rendered]))release(n)}
async function draw(n:number){
 const document=pdf.value,host=element(n),token=epoch,generation=renderEpoch
 if(!document||!host||rendered.has(n)||renders.has(n))return
 rendered.add(n)
 try{
  const page=await document.getPage(n)
  if(token!==epoch||generation!==renderEpoch||!nearby.has(n))return
  const original=page.getViewport({scale:1}),fit=Math.max(200,(root.value?.clientWidth||650)-32)
  const viewport=page.getViewport({scale:fit/original.width*zoom.value})
  const ratio=Math.min(window.devicePixelRatio||1,2,Math.sqrt(8000000/(viewport.width*viewport.height)))
  const canvas=documentOwnerCanvas();canvas.width=Math.ceil(viewport.width*ratio);canvas.height=Math.ceil(viewport.height*ratio)
  canvas.style.width=`${viewport.width}px`;canvas.style.height=`${viewport.height}px`;canvas.setAttribute('aria-label',`PDF 第 ${n} 页`)
  host.style.width=`${viewport.width}px`;host.style.minHeight=`${viewport.height}px`;host.replaceChildren(canvas)
  const rendering=page.render({canvas,viewport,transform:ratio!==1?[ratio,0,0,ratio,0,0]:undefined})
  renders.set(n,rendering);await rendering.promise
  if(token===epoch&&generation===renderEpoch){host.dataset.pdfRendered='true';renders.delete(n)}
 }catch(e:any){if(token===epoch&&generation===renderEpoch&&e?.name!=='RenderingCancelledException'){rendered.delete(n);host.textContent=`第 ${n} 页渲染失败，可重新打开重试`;error.value=e?.message||'PDF 页面渲染失败'}}
 finally{if(token===epoch&&generation===renderEpoch&&!nearby.has(n))release(n)}
}
function documentOwnerCanvas(){return window.document.createElement('canvas')}
function observe(){
 observer?.disconnect();nearby.clear();if(!root.value)return
 observer=new IntersectionObserver(entries=>{for(const entry of entries){const n=Number((entry.target as HTMLElement).dataset.pdfPage);if(entry.isIntersecting){nearby.add(n);void draw(n)}else{nearby.delete(n);release(n)}}},{root:root.value,rootMargin:'500px 0px'})
 root.value.querySelectorAll('[data-pdf-page]').forEach(el=>observer!.observe(el))
}
async function go(n:number){
 if(!Number.isSafeInteger(n)||n<1||n>pages.value.length){warning.value=`页码超出范围：本原件共 ${pages.value.length} 页`;return}
 inputPage.value=n;current.value=n;await nextTick();const el=element(n);if(el&&root.value)root.value.scrollTop=el.offsetTop
}
function onScroll(){
 if(frame)return;frame=requestAnimationFrame(()=>{frame=0;const scroll=root.value?.scrollTop||0;let low=1,high=pages.value.length
  while(low<high){const mid=Math.ceil((low+high)/2);if((element(mid)?.offsetTop||0)<=scroll+40)low=mid;else high=mid-1};current.value=low;inputPage.value=low
 })
}
async function rescale(){
 if(!pdf.value||!root.value)return
 const n=current.value,document=pdf.value,token=epoch,request=++resizeEpoch;resetRenders();const first=await document.getPage(1)
 if(token!==epoch||request!==resizeEpoch||document!==pdf.value||!root.value)return
 const view=first.getViewport({scale:1}),fit=Math.max(200,root.value.clientWidth-32)
 pageWidth.value=fit*zoom.value;pageHeight.value=view.height/view.width*pageWidth.value
 await nextTick();await go(n);observe()
}
async function load(){
 const token=++epoch;abort?.abort();abort=new AbortController();observer?.disconnect();resetRenders();await task?.destroy().catch(()=>{});task=undefined;pdf.value=null;pages.value=[];error.value='';warning.value='';loading.value=true
 if(!props.archiveVersionId){error.value='引用时未绑定可验证的原件版本，不能改用当前映射冒充历史原件';loading.value=false;return}
 try{
  const url=`/api/v1/knowledge/${encodeURIComponent(props.knowledgeId)}/preview?source=archive&archive_version_id=${encodeURIComponent(props.archiveVersionId)}`
  const blob=await get<Blob>(url,{responseType:'blob',signal:abort.signal});if(token!==epoch)return
  const assets=`${import.meta.env.BASE_URL}vendor/pdfjs/${version}/`
  const loadingTask=getDocument({data:new Uint8Array(await blob.arrayBuffer()),cMapUrl:assets+'cmaps/',cMapPacked:true,standardFontDataUrl:assets+'standard_fonts/',wasmUrl:assets+'wasm/',iccUrl:assets+'iccs/'})
  task=loadingTask;loadingTask.onPassword=()=>{error.value='原件是加密 PDF，需要解密后才能预览';void loadingTask.destroy()}
  const document=await loadingTask.promise;if(token!==epoch){await loadingTask.destroy();return};pdf.value=document
  const first=await document.getPage(1);if(token!==epoch)return
  const view=first.getViewport({scale:1}),fit=Math.max(200,(root.value?.clientWidth||650)-32)
  pageWidth.value=fit;pageHeight.value=view.height/view.width*fit;pages.value=Array.from({length:document.numPages},(_,i)=>i+1)
  const target=citationPageTarget(props.location,document.numPages);warning.value=target.warning
  if(!target.warning&&props.initialPage&&props.location.pages.includes(props.initialPage))target.page=props.initialPage
  loading.value=false;await nextTick();await go(target.page);observe()
 }catch(e:any){if(token===epoch&&!abort.signal.aborted)error.value=error.value||`原件预览失败：${e?.message||'原件已删除、版本失效或无访问权限'}`}
 finally{if(token===epoch)loading.value=false}
}
async function toggleFullscreen(){fullscreen.value=!fullscreen.value;if(fullscreen.value){bodyOverflow=window.document.body.style.overflow;window.document.body.style.overflow='hidden'}else window.document.body.style.overflow=bodyOverflow;await nextTick();await rescale()}
watch(()=>[props.knowledgeId,props.archiveVersionId],()=>void load(),{immediate:true})
watch(()=>props.location,()=>{if(pdf.value){const target=citationPageTarget(props.location,pdf.value.numPages);warning.value=target.warning;void go(target.page)}},{deep:true})
onMounted(()=>{resize=new ResizeObserver(()=>{void rescale()});if(root.value)resize.observe(root.value)})
onBeforeUnmount(()=>{epoch++;abort?.abort();observer?.disconnect();resize?.disconnect();resetRenders();void task?.destroy();if(frame)cancelAnimationFrame(frame);if(fullscreen.value)window.document.body.style.overflow=bodyOverflow})
</script>
<template>
 <Teleport to="body" :disabled="!fullscreen">
  <section class="archival-pdf" :class="{'archival-pdf-fullscreen':fullscreen}" aria-label="引用原件 PDF 阅读器">
   <div class="archival-pdf-controls">
    <label>页码 <input v-model.number="inputPage" aria-label="PDF 页码" type="number" min="1" :max="pages.length" @change="go(Number(inputPage))" @keydown.enter="go(Number(inputPage))" /> / {{pages.length}}</label>
    <button type="button" :disabled="zoom<=0.5" @click="zoom=Math.max(0.5,zoom-0.25);rescale()">缩小</button>
    <span>{{Math.round(zoom*100)}}%</span>
    <button type="button" :disabled="zoom>=3" @click="zoom=Math.min(3,zoom+0.25);rescale()">放大</button>
    <button type="button" @click="toggleFullscreen">{{fullscreen?'退出全屏':'全屏'}}</button>
   </div>
   <div class="archival-pdf-source">{{sourceLabel}} <button v-for="p in location.pages" :key="p" type="button" @click="go(p)">{{p}}</button></div>
   <div v-if="warning" class="archival-warning" role="status">{{warning}}</div>
   <div v-if="loading" role="status">正在读取指定版本 PDF…</div>
   <div v-if="error" role="alert" class="archival-warning">{{error}}</div>
   <div ref="root" class="archival-pdf-scroll" @scroll.passive="onScroll">
    <div v-for="n in pages" :key="n" :data-pdf-page="n" class="archival-pdf-page" :style="{width:pageWidth+'px',minHeight:pageHeight+'px'}"><span>第 {{n}} 页 · 按需加载</span></div>
   </div>
  </section>
 </Teleport>
</template>
<style scoped>
.archival-pdf{display:flex;flex-direction:column;border:1px solid #dedbd3;margin-top:14px;background:#faf9f6;min-width:0}
.archival-pdf-controls{display:flex;gap:8px;align-items:center;padding:8px;flex-wrap:wrap;font-size:12px}.archival-pdf-controls input{width:65px}.archival-pdf button{cursor:pointer;border:1px solid #ddd;background:#fff;padding:3px 8px;border-radius:4px}
.archival-pdf-source{font-size:12px;padding:3px 8px}.archival-pdf-source button{margin-left:6px}
.archival-pdf-scroll{position:relative;overflow:auto;height:62vh;min-height:300px;background:#d9d9d6;overscroll-behavior:contain;scrollbar-gutter:stable}
.archival-pdf-page{position:relative;margin:12px auto;background:white;box-shadow:0 1px 4px #888;color:#999;text-align:center;font-size:12px}.archival-pdf-page :deep(canvas){display:block}
.archival-pdf-fullscreen{position:fixed;inset:0;z-index:3000;margin:0;background:#fff}.archival-pdf-fullscreen .archival-pdf-scroll{flex:1;height:0;min-height:0}
</style>
