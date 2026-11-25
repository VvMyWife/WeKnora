import {get} from '@/utils/request'
import type {ArchivalCitation,CitationManifest} from '@/utils/archivalCitations'
export type CitationDetail = ArchivalCitation & {snapshot:string;current_content?:string;current_revision:number;current_changed:boolean;snapshot_at:string}
const base=(session:string,message:string)=>`/api/v1/messages/${encodeURIComponent(session)}/${encodeURIComponent(message)}/citations`
export function getCitationManifest(session:string,message:string,signal?:AbortSignal){
 return get<{enabled?:boolean;data:CitationManifest|null;unavailable_sources?:number}>(base(session,message),{signal})
}
export function getCitationEvidence(session:string,message:string,id:string,signal?:AbortSignal){
 return get<{data:CitationDetail}>(`${base(session,message)}/${encodeURIComponent(id)}`,{signal})
}
export function getCitationOrigin(session:string,message:string,id:string,knowledge:string,chunk:string,signal?:AbortSignal){
 return get<{data:{content:string;content_revision:number;state:'current_source_not_snapshot'}}>(`${base(session,message)}/${encodeURIComponent(id)}/origins/${encodeURIComponent(knowledge)}/chunks/${encodeURIComponent(chunk)}`,{signal})
}
