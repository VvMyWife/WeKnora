<template>
  <div class="archive-brand" :class="{ 'archive-brand--compact': compact }" aria-label="档案智查">
    <span class="archive-brand__symbol" aria-hidden="true">
      <img src="@/assets/img/yifang-brand-symbol.png" alt="" />
    </span>
    <span v-if="!compact" class="archive-brand__wordmark">
      <strong>档案智查</strong>
      <small>ARCHIVE INTELLIGENCE</small>
    </span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })
</script>

<style scoped>
.archive-brand {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
  color: var(--archive-ink, #221f1c);
  user-select: none;
}

.archive-brand__symbol {
  display: block;
  width: 39px;
  height: 39px;
  flex: 0 0 39px;
}

.archive-brand__symbol img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.archive-brand__wordmark {
  display: grid;
  min-width: 0;
  gap: 1px;
  line-height: 1;
  white-space: nowrap;
}

.archive-brand__wordmark strong {
  font-family: "Noto Serif SC", "Source Han Serif SC", "Songti SC", serif;
  font-size: 21px;
  font-weight: 750;
  letter-spacing: 0.14em;
}

.archive-brand__wordmark small {
  color: var(--archive-muted, #81766c);
  font-size: 8px;
  font-weight: 700;
  letter-spacing: 0.16em;
}

.archive-brand--compact .archive-brand__symbol {
  width: 34px;
  height: 34px;
  flex-basis: 34px;
}

:global(html[theme-mode="dark"]) .archive-brand {
  color: #f3ece4;
}

:global(html[theme-mode="dark"]) .archive-brand__wordmark small {
  color: #a89e94;
}
</style>
