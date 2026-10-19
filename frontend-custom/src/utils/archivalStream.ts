// Both quick QA and agent streams must replace provisional text with the same
// canonical answer; never append the final answer a second time.
export function applyValidatedCitationAnswer(message:Record<string,unknown>,payload:Record<string,unknown>|undefined):boolean {
 const manifest=payload?.archival_citations
 const answer=payload?.final_answer
 if(!manifest||typeof manifest!=='object'||typeof answer!=='string')return false
 message.archival_citations=manifest
 message.citation_status=(manifest as Record<string,unknown>).state
 message.content=answer
 if(Array.isArray(message.agentEventStream)){
  // Restored quick-QA timelines may be markRaw. Replace the array reference,
  // otherwise the visible answer can remain the provisional unnumbered text.
  message.agentEventStream=[...message.agentEventStream.map(entry=>entry.type==='answer'?{...entry,superseded:true}:entry),{type:'answer',event_id:'archival-validated',content:answer,done:true}]
 }
 return true
}
