// Development-only fixture. Vite's production entry remains index.html.
import {createApp,h,ref} from 'vue'
import TDesign from 'tdesign-vue-next'
import 'tdesign-vue-next/es/style/index.css'
import i18n from '../src/i18n'
import ArchivalCitations from '../src/components/ArchivalCitations.vue'
import {preprocessCitationTags} from '../src/utils/citationMarkdown'
import {sanitizeMarkdownHTML} from '../src/utils/security'
const session='11111111-1111-4111-8111-111111111111',message='22222222-2222-4222-8222-222222222222'
createApp({components:{ArchivalCitations},setup(){
 const panel=ref<InstanceType<typeof ArchivalCitations>|null>(null)
 const html=sanitizeMarkdownHTML(preprocessCitationTags(`<kb doc="隔离样例" chunk_id="33333333-3333-4333-8333-333333333333" citation_id="${'a'.repeat(64)}" footnote="1" />`))
 const open=(event:MouseEvent)=>{const el=event.target instanceof Element?event.target.closest('[data-citation-id]'):null;if(el)panel.value?.openCitation(el.getAttribute('data-citation-id')||'')}
 return ()=>h('main',{style:'max-width:900px;margin:40px auto;font-family:sans-serif'},[
  h('h1','档案引用隔离验收'),h('p',{onClick:open},['有依据的档案说明 ',h('span',{innerHTML:html})]),
  h(ArchivalCitations,{ref:panel,sessionId:session,messageId:message,complete:true}),
 ])
}}).use(TDesign).use(i18n).mount('#app')
