<script setup lang="ts">
defineProps<{ code: string }>()
</script>

<template>
  <!-- 日本 -->
  <svg v-if="code === 'JP'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#fff" rx="1"/>
    <circle cx="9" cy="6" r="3.2" fill="#bc002d"/>
  </svg>
  <!-- 香港 -->
  <svg v-else-if="code === 'HK'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#de2910" rx="1"/>
    <circle cx="9" cy="6" r="2.8" fill="none" stroke="#fff" stroke-width="0.8"/>
    <circle cx="9" cy="6" r="0.6" fill="#fff"/>
  </svg>
  <!-- 新加坡 -->
  <svg v-else-if="code === 'SG'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="6" fill="#ed2939" rx="1"/>
    <rect y="6" width="18" height="6" fill="#fff" rx="1"/>
    <path d="M3 3.5 l0.8 0.5 l-0.3-0.9 l0.7-0.1 l-0.4-0.6 l0.4-0.6 l-0.7-0.1 l0.3-0.9 l-0.8 0.5 l-0.8-0.5 l0.3 0.9 l-0.7 0.1 l0.4 0.6 l-0.4 0.6 l0.7 0.1 Z" fill="#fff" transform="translate(0,0.5)"/>
  </svg>
  <!-- 美国 -->
  <svg v-else-if="code === 'US'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#fff" rx="1"/>
    <rect v-for="i in 7" :key="i" :y="(i-1)*1.7" width="18" height="0.85" fill="#b22234"/>
    <rect width="7.2" height="5.1" fill="#3c3b6e"/>
  </svg>
  <!-- fallback -->
  <span v-else class="flag-fallback">{{ code }}</span>
</template>

<style scoped>
svg { flex-shrink: 0; vertical-align: middle; border-radius: 1px; }
.flag-fallback { font-size: 10px; color: #8899aa; }
</style>
