<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Download, RefreshCw, X, TriangleAlert, Sparkles } from 'lucide-vue-next'

type Phase = 'hidden' | 'available' | 'downloading' | 'downloaded' | 'error'

const phase = ref<Phase>('hidden')
const version = ref('')
const percent = ref(0)
const speed = ref('')
const errorMsg = ref('')

function fmtSpeed(bps: number): string {
  if (!bps || bps < 1) return ''
  const mb = bps / 1024 / 1024
  return mb >= 1 ? `${mb.toFixed(1)} MB/s` : `${(bps / 1024).toFixed(0)} KB/s`
}

onMounted(() => {
  const api = window.electronAPI?.update
  if (!api) return
  api.onAvailable((info) => { version.value = info.version; phase.value = 'available' })
  api.onProgress((p) => { phase.value = 'downloading'; percent.value = Math.round(p.percent); speed.value = fmtSpeed(p.bytesPerSecond) })
  api.onDownloaded((info) => { version.value = info.version; phase.value = 'downloaded' })
  api.onError((msg) => { errorMsg.value = msg; phase.value = 'error' })
})

function startDownload() {
  phase.value = 'downloading'
  percent.value = 0
  window.electronAPI?.update.download()
}
function dismiss() { phase.value = 'hidden' }

const title = computed(() => {
  switch (phase.value) {
    case 'available': return '发现新版本'
    case 'downloading': return '正在下载更新'
    case 'downloaded': return '更新就绪'
    case 'error': return '更新失败'
    default: return ''
  }
})
</script>

<template>
  <Transition name="upd">
    <div v-if="phase !== 'hidden'" class="upd-card" :class="phase">
      <div class="upd-head">
        <span class="upd-icon">
          <Sparkles v-if="phase === 'available'" :size="15" />
          <Download v-else-if="phase === 'downloading'" :size="15" />
          <RefreshCw v-else-if="phase === 'downloaded'" :size="15" class="spin" />
          <TriangleAlert v-else :size="15" />
        </span>
        <span class="eyebrow">{{ title }}</span>
        <span v-if="version" class="upd-ver">v{{ version }}</span>
        <span
          v-if="phase === 'available' || phase === 'error'"
          class="upd-x" @click="dismiss"
        ><X :size="14" /></span>
      </div>

      <!-- available -->
      <template v-if="phase === 'available'">
        <p class="upd-text">已检测到新版本，更新将在后台下载，完成后自动重启。</p>
        <button class="upd-btn" @click="startDownload">
          <Download :size="14" /> 下载并重启
        </button>
      </template>

      <!-- downloading -->
      <template v-else-if="phase === 'downloading'">
        <div class="upd-meter">
          <div class="upd-fill" :style="{ width: percent + '%' }"></div>
        </div>
        <div class="upd-stat">
          <span class="mono">{{ percent }}%</span>
          <span class="mono dim">{{ speed }}</span>
        </div>
      </template>

      <!-- downloaded -->
      <template v-else-if="phase === 'downloaded'">
        <p class="upd-text">下载完成，正在重启以完成更新…</p>
      </template>

      <!-- error -->
      <template v-else>
        <p class="upd-text err">{{ errorMsg || '无法获取更新，请稍后重试。' }}</p>
        <button class="upd-btn ghost" @click="startDownload">重试</button>
      </template>
    </div>
  </Transition>
</template>

<style scoped>
.upd-card {
  position: fixed; right: 20px; bottom: 20px; z-index: 900;
  width: 312px; padding: 15px 16px;
  background: linear-gradient(180deg, var(--panel-2), var(--panel));
  border: 1px solid var(--line-strong);
  border-radius: 12px;
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(54, 211, 238, 0.06);
}
.upd-card.available { border-color: rgba(54, 211, 238, 0.35); box-shadow: 0 18px 50px rgba(0,0,0,0.5), 0 0 22px rgba(54,211,238,0.12); }
.upd-card.error { border-color: rgba(255, 93, 108, 0.35); }

.upd-head { display: flex; align-items: center; gap: 8px; margin-bottom: 11px; }
.upd-icon { display: inline-flex; color: var(--accent); }
.upd-card.error .upd-icon { color: var(--bad); }
.upd-head .eyebrow { flex: 1; color: var(--text); letter-spacing: 1.5px; }
.upd-ver {
  font-family: var(--mono); font-size: 11px; color: var(--accent);
  padding: 1px 7px; border-radius: 5px;
  background: var(--accent-dim); border: 1px solid rgba(54, 211, 238, 0.22);
}
.upd-x {
  display: inline-flex; cursor: pointer; color: var(--text-dim);
  width: 22px; height: 22px; align-items: center; justify-content: center; border-radius: 5px;
}
.upd-x:hover { color: var(--text); background: rgba(120, 168, 210, 0.10); }

.upd-text { font-size: 12px; line-height: 1.6; color: var(--text-dim); margin-bottom: 12px; }
.upd-text.err { color: var(--bad); word-break: break-word; }

.upd-btn {
  width: 100%; padding: 9px;
  display: inline-flex; align-items: center; justify-content: center; gap: 7px;
  font-family: var(--ui); font-size: 12.5px; font-weight: 600; letter-spacing: 1px;
  color: var(--bg-deep); background: var(--accent);
  border: none; border-radius: 8px; cursor: pointer;
  box-shadow: 0 0 16px var(--accent-glow); transition: filter 0.2s;
}
.upd-btn:hover { filter: brightness(1.1); }
.upd-btn.ghost {
  color: var(--text); background: transparent; box-shadow: none;
  border: 1px solid var(--line-strong);
}
.upd-btn.ghost:hover { border-color: var(--accent); color: var(--accent); filter: none; }

.upd-meter { height: 6px; border-radius: 100px; background: rgba(120, 168, 210, 0.10); overflow: hidden; }
.upd-fill {
  height: 100%; border-radius: 100px;
  background: linear-gradient(90deg, #2bb6d4, var(--accent));
  box-shadow: 0 0 10px var(--accent-glow);
  transition: width 0.3s ease;
}
.upd-stat { display: flex; justify-content: space-between; margin-top: 8px; font-size: 11px; }
.upd-stat .mono { font-family: var(--mono); color: var(--text); }
.upd-stat .dim { color: var(--text-dim); }

.spin { animation: spin 1.1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.upd-enter-active { transition: transform 0.35s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.35s; }
.upd-leave-active { transition: transform 0.25s, opacity 0.25s; }
.upd-enter-from, .upd-leave-to { opacity: 0; transform: translateY(16px); }
</style>
