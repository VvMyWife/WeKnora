import type { ArchiveItem } from '@/api/archive-mapping';
export function relativeOriginalPath(file: Pick<File,'name'|'webkitRelativePath'>): string {
  const parts=(file.webkitRelativePath || file.name).replaceAll('\\','/').split('/');
  return parts.length>1 ? parts.slice(1).join('/') : parts[0];
}
export function commonParsedRoot(paths: string[]): string {
  if (!paths.length) return '';
  const prefix=paths[0].split('/').slice(0,-1);
  for (const p of paths.slice(1)) {const dirs=p.split('/').slice(0,-1);while(prefix.some((x,i)=>dirs[i]!==x)) prefix.pop();}
  return prefix.join('/');
}
export function archiveCounts(items: ArchiveItem[]) {
  const count=(status:string)=>items.filter(i=>i.status===status).length;
  return {total:items.length,bound:count('bound'),failed:count('failed'),skipped:count('skipped'),pending:items.filter(i=>!['bound','failed','skipped'].includes(i.status)).length};
}
export async function runArchiveUploads<T>(items:T[],send:(item:T)=>Promise<void>,stop:()=>boolean) {
  let cursor=0;
  await Promise.all(Array.from({length:Math.min(2,items.length)},async()=>{while(!stop() && cursor<items.length){const item=items[cursor++];await send(item);}}));
}
