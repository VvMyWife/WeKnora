<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getCitationManifest } from '@/api/archival-citations'
import { get } from '@/utils/request'
import { citationSourceGroups } from '@/utils/archivalCitations'
import { loadPdfJs } from './source-preview/pdfjsLoader'
import type { PDFDocumentLoadingTask } from 'pdfjs-dist'

const props = defineProps<{ sessionId: string; messageId: string }>()
const emit = defineEmits<{ renderedDocuments: [ids: string[]] }>()
type Preview = { url: string; page: number; title: string }
const root = ref<HTMLElement>()
const dialog = ref<HTMLDialogElement>()
const selected = ref<Preview>()
const previews = ref<Preview[]>([])
const visible = ref(false)
let observer: IntersectionObserver | undefined

onMounted(() => {
  if (!root.value || typeof IntersectionObserver === 'undefined') {
    visible.value = true
    return
  }
  observer = new IntersectionObserver(entries => {
    if (!entries.some(entry => entry.isIntersecting)) return
    visible.value = true
    observer?.disconnect()
  }, { rootMargin: '150px' })
  observer.observe(root.value)
})
onBeforeUnmount(() => observer?.disconnect())

watch(() => [visible.value, props.sessionId, props.messageId] as const, ([active, session, message], _, cleanup) => {
  if (!active || !session || !message) return
  const controller = new AbortController()
  const signal = controller.signal
  const urls: string[] = []
  let task: PDFDocumentLoadingTask | undefined
  previews.value = []
  emit("renderedDocuments", [])
  cleanup(() => {
    controller.abort()
    if (task) void task.destroy().catch(() => {})
    dialog.value?.close()
    selected.value = undefined
    urls.forEach(url => URL.revokeObjectURL(url))
  })

  async function load() {
    const manifest = await getCitationManifest(session, message, signal)
    const groups = citationSourceGroups(manifest.data).filter(group =>
      !group.derived && group.version &&
      group.citations.some(citation => citation.physical_pages && citation.location.status === 'known'),
    )
    if (!groups.length || signal.aborted) return
    const lib = await loadPdfJs()
    const assets = `${import.meta.env.BASE_URL}vendor/pdfjs/${lib.version}/`
    const renderedDocuments: string[] = []
    const renderedVersions = new Set<string>()
    for (const group of groups) {
      if (signal.aborted) return
      if (renderedVersions.has(group.version)) continue
      try {
        const path = `/api/v1/knowledge/${encodeURIComponent(group.knowledgeId)}/preview?source=archive&archive_version_id=${encodeURIComponent(group.version)}`
        const blob = await get<Blob>(path, { responseType: 'blob', signal })
        const data = new Uint8Array(await blob.arrayBuffer())
        if (signal.aborted) return
        task = lib.getDocument({
          data, cMapUrl: assets + 'cmaps/', cMapPacked: true,
          standardFontDataUrl: assets + 'standard_fonts/', wasmUrl: assets + 'wasm/', iccUrl: assets + 'iccs/',
        })
        task.onPassword = () => { void task?.destroy().catch(() => {}) }
        const pdf = await task.promise
        const pages = new Set<number>()
        for (const citation of group.citations) {
          if (!citation.physical_pages || citation.location.status !== 'known') continue
          for (const number of citation.location.pages) {
            if (!Number.isSafeInteger(number) || number < 1 || number > pdf.numPages) continue
            pages.add(number)
          }
        }
        const initialCount = previews.value.length
        for (const number of [...pages].sort((a, b) => a - b)) {
          if (signal.aborted) return
          const page = await pdf.getPage(number)
          const view = page.getViewport({ scale: 1 })
          const viewport = page.getViewport({ scale: Math.min(1400 / view.width, 2000 / view.height) })
          const canvas = document.createElement('canvas')
          canvas.width = Math.ceil(viewport.width)
          canvas.height = Math.ceil(viewport.height)
          await page.render({ canvas, viewport }).promise
          const image = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve, 'image/png'))
          if (image && !signal.aborted) {
            const url = URL.createObjectURL(image)
            urls.push(url)
            previews.value.push({ url, page: number, title: group.title })
          }
        }
        if (!signal.aborted && previews.value.length > initialCount) {
          renderedVersions.add(group.version)
          renderedDocuments.push(group.knowledgeId.toLowerCase())
          emit('renderedDocuments', [...renderedDocuments])
        }
      } catch {
        if (signal.aborted) return
      } finally {
        await task?.destroy().catch(() => {})
        task = undefined
      }
    }
  }
  void load().catch(() => {})
}, { immediate: true })

async function enlarge(preview: Preview) {
  selected.value = preview
  await nextTick()
  dialog.value?.showModal()
}
</script>

<template>
  <section ref="root" class="inline-citation-pages" aria-label="引用原件页面">
    <figure v-for="preview in previews" :key="preview.url">
      <button type="button" :aria-label="`放大 PDF 第 ${preview.page} 页`" @click="enlarge(preview)">
        <img :src="preview.url" :alt="`原件 PDF 第 ${preview.page} 页`" />
      </button>
      <figcaption>{{ preview.title }} · PDF 第 {{ preview.page }} 页</figcaption>
    </figure>
    <Teleport to="body">
      <dialog ref="dialog" class="inline-citation-dialog" aria-label="引用页放大预览" @close="selected = undefined">
        <button type="button" autofocus @click="dialog?.close()">关闭</button>
        <img v-if="selected" :src="selected.url" :alt="`原件 PDF 第 ${selected.page} 页`" />
      </dialog>
    </Teleport>
  </section>
</template>

<style scoped>
.inline-citation-pages:has(figure) { margin: 12px 0; }
.inline-citation-pages p, figcaption { color: var(--td-text-color-secondary); font-size: 13px; }
figure { margin: 0 0 12px; }
figure button { display: block; width: 340px; max-width: 100%; padding: 0; overflow: hidden; cursor: zoom-in; border: 1px solid var(--td-component-stroke); border-radius: var(--app-radius-sm); }
figure img { display: block; width: 100%; }
.inline-citation-dialog { max-width: min(94vw, 1000px); max-height: 94vh; border: 0; border-radius: var(--app-radius-md); }
.inline-citation-dialog::backdrop { background: #0009; }
.inline-citation-dialog button { position: sticky; top: 0; display: block; margin-left: auto; cursor: pointer; }
.inline-citation-dialog img { display: block; width: 100%; }
</style>
