<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCw, Settings, X } from 'lucide-vue-next'

const visible = ref(false)
const loading = ref(false)
const form = ref({ wsUrl: 'wss://frostrain.tech/nc/ws' })
const saved = ref(false)
const checkingUpdate = ref(false)
const updateState = ref('')

onMounted(async () => {
  try {
    const cfg = await window.electronAPI?.getConfig()
    if (cfg?.wsUrl) form.value.wsUrl = cfg.wsUrl
  } catch { /* */ }
  const update = window.electronAPI?.update
  update?.onAvailable((info) => {
    checkingUpdate.value = false
    updateState.value = `发现新版本 v${info.version}`
  })
  update?.onNone(() => {
    checkingUpdate.value = false
    updateState.value = '已是最新版本'
  })
  update?.onError((msg) => {
    checkingUpdate.value = false
    updateState.value = msg || '检查更新失败'
  })
})

function open() { visible.value = true }
function close() { if (!loading.value) visible.value = false }

async function handleSave() {
  loading.value = true
  try {
    await window.electronAPI?.saveConfig({ wsUrl: form.value.wsUrl })
  } catch { /* */ }
  loading.value = false
  saved.value = true
  setTimeout(() => { saved.value = false; visible.value = false }, 900)
}

async function checkUpdate() {
  checkingUpdate.value = true
  updateState.value = '正在检查更新'
  try {
    await window.electronAPI?.update.check()
  } catch {
    checkingUpdate.value = false
    updateState.value = '检查更新失败'
  }
}
</script>

<template>
  <span class="settings-btn" @click="open"><Settings :size="16" /></span>

  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="overlay" @click.self="close">
        <div class="modal">
          <div class="modal-head">
            <span class="eyebrow">连接设置</span>
            <span class="modal-x" @click="close"><X :size="15" /></span>
          </div>

          <label class="field-label">Go Agent WebSocket 地址</label>
          <input
            v-model="form.wsUrl"
            class="field-input"
            placeholder="wss://frostrain.tech/nc/ws"
            spellcheck="false"
            @keydown.enter="handleSave"
          />

          <button
            class="save-btn"
            :class="{ ok: saved }"
            :disabled="loading"
            @click="handleSave"
          >
            <span v-if="loading" class="spinner"></span>
            {{ saved ? '已保存 — 重新连接中…' : loading ? '保存中…' : '保存' }}
          </button>

          <button
            class="check-btn"
            :disabled="checkingUpdate"
            @click="checkUpdate"
          >
            <RefreshCw :size="14" :class="{ spin: checkingUpdate }" />
            {{ checkingUpdate ? '正在检查' : '检查更新' }}
          </button>
          <div v-if="updateState" class="update-state">{{ updateState }}</div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.settings-btn {
  display: inline-flex; align-items: center; justify-content: center;
  width: 34px; height: 28px; color: var(--text-dim);
  border-radius: 6px; cursor: pointer; transition: all 0.15s;
  -webkit-app-region: no-drag;
}
.settings-btn:hover { color: var(--accent); background: rgba(120, 168, 210, 0.10); }

.overlay {
  position: fixed; inset: 0; z-index: 1000;
  display: flex; align-items: center; justify-content: center;
  background: rgba(4, 7, 11, 0.6);
  backdrop-filter: blur(6px);
}
.modal {
  width: 440px; max-width: calc(100vw - 48px);
  padding: 22px;
  background: linear-gradient(180deg, var(--panel-2), var(--panel));
  border: 1px solid var(--line-strong);
  border-radius: 14px;
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(54, 211, 238, 0.06);
}
.modal-head {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 20px;
}
.modal-head .eyebrow { font-size: 12px; letter-spacing: 2.5px; color: var(--accent); }
.modal-x {
  display: inline-flex; cursor: pointer; color: var(--text-dim);
  width: 26px; height: 26px; align-items: center; justify-content: center;
  border-radius: 6px; transition: all 0.15s;
}
.modal-x:hover { color: var(--text); background: rgba(120, 168, 210, 0.10); }

.field-label {
  display: block; font-size: 11px; font-weight: 500; letter-spacing: 0.5px;
  color: var(--text-dim); margin-bottom: 8px;
}
.field-input {
  width: 100%; padding: 11px 13px;
  font-family: var(--mono); font-size: 13px; color: var(--text);
  background: var(--bg-deep);
  border: 1px solid var(--line-strong); border-radius: 8px;
  outline: none; transition: border-color 0.18s, box-shadow 0.18s;
}
.field-input::placeholder { color: var(--text-faint); }
.field-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(54, 211, 238, 0.12);
}

.save-btn {
  width: 100%; margin-top: 20px; padding: 11px;
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  font-family: var(--ui); font-size: 13px; font-weight: 600; letter-spacing: 1px;
  color: var(--bg-deep); background: var(--accent);
  border: none; border-radius: 8px; cursor: pointer;
  transition: all 0.2s; box-shadow: 0 0 18px var(--accent-glow);
}
.save-btn:hover:not(:disabled) { filter: brightness(1.1); }
.save-btn:disabled { cursor: default; opacity: 0.85; }
.save-btn.ok { background: var(--ok); box-shadow: 0 0 18px rgba(61, 220, 151, 0.5); }

.check-btn {
  width: 100%; margin-top: 10px; padding: 10px;
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  font-family: var(--ui); font-size: 12.5px; font-weight: 600; letter-spacing: 1px;
  color: var(--text); background: transparent;
  border: 1px solid var(--line-strong); border-radius: 8px; cursor: pointer;
  transition: all 0.2s;
}
.check-btn:hover:not(:disabled) { color: var(--accent); border-color: var(--accent); background: rgba(54, 211, 238, 0.08); }
.check-btn:disabled { cursor: default; opacity: 0.75; }
.update-state {
  margin-top: 9px;
  font-size: 11.5px;
  color: var(--text-dim);
  word-break: break-word;
}

.spinner {
  width: 13px; height: 13px; border-radius: 50%;
  border: 2px solid rgba(4, 7, 11, 0.3); border-top-color: var(--bg-deep);
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.spin { animation: spin 0.8s linear infinite; }

/* Transition */
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-active .modal { transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1); }
.modal-enter-from .modal { transform: translateY(12px) scale(0.97); }
</style>
