import {cpSync,mkdirSync,readFileSync} from 'node:fs'
import {createRequire} from 'node:module'
import {dirname,resolve} from 'node:path'
const require=createRequire(import.meta.url)
const source=dirname(require.resolve('pdfjs-dist/package.json'))
const {version}=JSON.parse(readFileSync(resolve(source,'package.json'),'utf8'))
const target=resolve(import.meta.dirname,'../public/vendor/pdfjs',version)
mkdirSync(target,{recursive:true})
for(const folder of ['cmaps','standard_fonts','wasm','iccs'])cpSync(resolve(source,folder),resolve(target,folder),{recursive:true})
console.log(`Local PDF.js assets ready: ${version}`)
