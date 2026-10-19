import { get, post, put, postUpload } from '@/utils/request';
export interface ArchiveVersion {
  id: string; knowledge_id: string; file_name: string; relative_path: string;
  file_type: string; file_size: number; sha256: string; created_by: string; created_at: string;
}
export interface ArchiveTarget { id: string; title: string; path: string; mapped?: boolean }
export interface ArchiveItem {
  id: string; relative_path: string; file_size: number; candidates: ArchiveTarget[] | null;
  knowledge_id: string; status: string; error: string; version_id: string;
}
export interface ArchiveBatch { id: string; parsed_root: string; created_at: string }
const base = (kb: string) => `/api/v1/knowledge-bases/${encodeURIComponent(kb)}/knowledge/archive-mapping/batches`;
export const listArchiveBatches = (kb: string) => get<{data:{batches: ArchiveBatch[]; targets: ArchiveTarget[]}}>(base(kb));
export const getArchiveBatch = (kb: string, batch: string) => get<{data:{batch: ArchiveBatch; items: ArchiveItem[]}}>(`${base(kb)}/${encodeURIComponent(batch)}`);
export const createArchiveBatch = (kb: string, parsed_root: string, files: {relative_path:string; file_size:number}[]) => post<{data:{batch:ArchiveBatch; items:ArchiveItem[]}}>(base(kb), {parsed_root,files});
export const uploadArchiveItem = (kb: string, batch: string, item: ArchiveItem, file: File, replace: boolean) => {
  const data = new FormData(); data.append('file',file); data.append('knowledge_id',item.knowledge_id); data.append('replace',String(replace));
  return postUpload(`${base(kb)}/${batch}/items/${item.id}/file`,data,undefined,{timeout:0});
};
export const skipArchiveItem = (kb:string,batch:string,item:string) => put(`${base(kb)}/${batch}/items/${item}/skip`);
export const getArchiveSource = (id:string) => get<{data:{active:ArchiveVersion|null; versions:ArchiveVersion[]; storage_bytes:number}}>(`/api/v1/knowledge/${encodeURIComponent(id)}/archive-source`);
export const setArchiveSource = (id:string,version_id:string) => put(`/api/v1/knowledge/${encodeURIComponent(id)}/archive-source`,{version_id});
