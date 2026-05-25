<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { SettingOutlined } from '@ant-design/icons-vue'

const visible = ref(false)
const loading = ref(false)
const form = ref({ wsUrl: 'ws://116.62.179.231:9527/ws' })
const saved = ref(false)

onMounted(async () => {
  try {
    const cfg = await window.electronAPI?.getConfig()
    if (cfg?.wsUrl) form.value.wsUrl = cfg.wsUrl
  } catch { /* */ }
})

async function handleOk() {
  loading.value = true
  try {
    await window.electronAPI?.saveConfig({ wsUrl: form.value.wsUrl })
  } catch { /* */ }
  loading.value = false
  saved.value = true
  setTimeout(() => { saved.value = false; visible.value = false }, 800)
}
</script>

<template>
  <a-button type="text" @click="visible = true" class="settings-btn">
    <SettingOutlined />
  </a-button>

  <a-modal
    v-model:open="visible"
    title="连接设置"
    :footer="null"
    :maskClosable="true"
    width="420px"
  >
    <a-form layout="vertical" class="settings-form">
      <a-form-item label="Go Agent WebSocket 地址">
        <a-input v-model:value="form.wsUrl" placeholder="ws://116.62.179.231:9527/ws" />
      </a-form-item>

      <a-form-item>
        <a-button
          type="primary"
          :loading="loading"
          :class="{ 'save-ok': saved }"
          @click="handleOk"
          block
        >
          {{ saved ? '已保存 — 重新连接中...' : '保存' }}
        </a-button>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<style scoped>
.settings-btn {
  color: #8899aa; font-size: 16px;
  -webkit-app-region: no-drag;
}
.settings-btn:hover { color: #e0e0e0; }
.settings-form { margin-top: 16px; }
.save-ok { background: #52c41a !important; border-color: #52c41a !important; }
</style>
