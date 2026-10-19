import { test } from 'node:test';
import assert from 'node:assert/strict';
import { relativeOriginalPath, commonParsedRoot, runArchiveUploads } from './archiveMapping';
test('strip only the explicit upload root, preserve nested duplicate names',()=>{
 assert.equal(relativeOriginalPath({name:'办法.pdf',webkitRelativePath:'原件/甲/办法.pdf'}),'甲/办法.pdf');
 assert.equal(relativeOriginalPath({name:'办法.pdf',webkitRelativePath:'原件/乙/办法.pdf'}),'乙/办法.pdf');
 assert.equal(commonParsedRoot(['政策_MD/甲/办法.md','政策_MD/乙/办法.md']),'政策_MD');
 assert.equal(commonParsedRoot(['甲.md','乙/办法.md']),'');
});
test('at most two uploads concurrently; stop preserves unstarted work',async()=>{
 let active=0,peak=0,done=0;
 await runArchiveUploads([1,2,3,4],async()=>{active++;peak=Math.max(peak,active);await new Promise(r=>setTimeout(r,3));active--;done++;},()=>done>=1);
 assert.equal(peak,2);assert.equal(done,2);
});
