<script setup lang="ts">
defineProps<{ code: string }>()
</script>

<template>
  <span v-if="code" class="fi" :class="`fi-${code.toLowerCase()}`" :title="code"></span>
  <span v-else class="fi fi-xx" style="opacity:0.3"></span>
</template>

<style>
@import 'flag-icons/css/flag-icons.min.css';
</style>

<style scoped>
.fi {
  width: 18px; height: 12px;
  border-radius: 1px; flex-shrink: 0;
}
</style>
