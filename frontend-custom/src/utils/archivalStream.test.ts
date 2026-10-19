import assert from 'node:assert/strict'
import test from 'node:test'
import {applyValidatedCitationAnswer} from './archivalStream.ts'
test('normal QA replaces provisional answer and carries manifest',()=>{
 const message:Record<string,unknown>={content:'draft',is_completed:false}
 const manifest={state:'verified',citations:[]}
 assert.equal(applyValidatedCitationAnswer(message,{archival_citations:manifest,final_answer:'canonical [1]'}),true)
 assert.equal(message.content,'canonical [1]');assert.equal(message.archival_citations,manifest)
 assert.equal(message.is_completed,false) // caller completes only after canonical replacement
})
test('agent supersedes earlier answer events, not its tool history',()=>{
 const message={content:'draft',agentEventStream:[{type:'tool_call'},{type:'answer',content:'draft'}]} as Record<string,unknown>
 const provisional=message.agentEventStream as Record<string,unknown>[]
 applyValidatedCitationAnswer(message,{archival_citations:{state:'needs_review'},final_answer:'final'})
 const entries=message.agentEventStream as Record<string,unknown>[]
 assert.notEqual(entries,provisional);assert.equal(provisional[1].superseded,undefined)
 assert.equal(entries[0].superseded,undefined);assert.equal(entries[1].superseded,true)
 assert.equal(entries[2].content,'final')
 assert.equal(applyValidatedCitationAnswer(message,{}),false)
 assert.equal(message.content,'final')
})
