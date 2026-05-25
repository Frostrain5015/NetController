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
  <!-- 台湾 -->
  <svg v-else-if="code === 'TW'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#fe0000" rx="1"/>
    <rect x="3" y="1.5" width="4" height="9" fill="#000095"/>
    <circle cx="5" cy="6" r="1.2" fill="#fff"/>
  </svg>
  <!-- 韩国 -->
  <svg v-else-if="code === 'KR'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#fff" rx="1"/>
    <circle cx="9" cy="6" r="3.6" fill="#c60c30"/>
    <rect y="2" width="18" height="4" fill="#003478"/>
    <circle cx="9" cy="6" r="1" fill="#c60c30"/>
  </svg>
  <!-- 德国 -->
  <svg v-else-if="code === 'DE'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="4" fill="#000" rx="1"/>
    <rect y="4" width="18" height="4" fill="#dd0000"/>
    <rect y="8" width="18" height="4" fill="#ffce00"/>
  </svg>
  <!-- 英国 -->
  <svg v-else-if="code === 'GB'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#012169" rx="1"/>
    <path d="M0 0 l18 12 M18 0 l-18 12" stroke="#fff" stroke-width="2"/>
    <path d="M0 0 l18 12 M18 0 l-18 12" stroke="#c8102e" stroke-width="1"/>
    <path d="M9 0 v12 M0 6 h18" stroke="#fff" stroke-width="3"/>
    <path d="M9 0 v12 M0 6 h18" stroke="#c8102e" stroke-width="1.5"/>
  </svg>
  <!-- 法国 -->
  <svg v-else-if="code === 'FR'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="5" height="12" fill="#001e96" rx="1"/>
    <rect x="5" width="8" height="12" fill="#fff"/>
    <rect x="13" width="5" height="12" fill="#ee2435"/>
  </svg>
  <!-- 加拿大 -->
  <svg v-else-if="code === 'CA'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="4" height="12" fill="#f00" rx="1"/>
    <rect x="4" width="10" height="12" fill="#fff"/>
    <rect x="14" width="4" height="12" fill="#f00"/>
    <path d="M9 3 l-0.8 2 l-1.5-0.5 l0.5 1.5 l-1.5 0.5 l1.5 0.5 l-0.5 1.5 l1.5-0.5 l0.8 2 l0.8-2 l1.5 0.5 l-0.5-1.5 l1.5-0.5 l-1.5-0.5 l0.5-1.5 l-1.5 0.5 Z" fill="#f00"/>
  </svg>
  <!-- 澳大利亚 -->
  <svg v-else-if="code === 'AU'" width="18" height="12" viewBox="0 0 18 12">
    <rect width="18" height="12" fill="#00008b" rx="1"/>
    <rect x="0" y="0" width="6" height="4" fill="#00008b"/>
    <path d="M0 0 l3 2 M3 0 l-3 2 M0 0 v4 M3 0 v4 M0 0 h3 M0 2 h3" stroke="#fff" stroke-width="0.6"/>
    <circle cx="11" cy="4" r="0.5" fill="#fff"/>
    <circle cx="14" cy="6" r="0.4" fill="#fff"/>
    <circle cx="13" cy="9" r="0.5" fill="#fff"/>
    <circle cx="10" cy="7" r="0.3" fill="#fff"/>
    <circle cx="16" cy="3" r="0.4" fill="#fff"/>
  </svg>
  <!-- fallback -->
  <span v-else class="flag-fallback">{{ code }}</span>
</template>

<style scoped>
svg { flex-shrink: 0; vertical-align: middle; border-radius: 1px; }
.flag-fallback { font-size: 10px; color: #8899aa; }
</style>
