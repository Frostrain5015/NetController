<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { SettingOutlined } from '@ant-design/icons-vue'

const visible = ref(false)
const loading = ref(false)
const form = ref({ host: '', port: 22, username: 'root', password: '' })
const saved = ref(false)

onMounted(async () => {
  if (window.electronAPI) {
    const s = await window.electronAPI.getSettings()
    form.value = { ...s }
  }
})

async function handleOk() {
  loading.value = true
  if (window.electronAPI) {
    await window.electronAPI.saveSettings({ ...form.value })
  }
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
    title="SSH 连接设置"
    :footer="null"
    :maskClosable="true"
    width="400px"
  >
    <a-form layout="vertical" class="settings-form">
      <a-form-item label="服务器地址">
        <a-input v-model:value="form.host" placeholder="116.62.179.231" />
      </a-form-item>
      <a-form-item label="SSH 端口">
        <a-input-number v-model:value="form.port" :min="1" :max="65535" style="width:100%" />
      </a-form-item>
      <a-form-item label="用户名">
        <a-input v-model:value="form.username" placeholder="root" />
      </a-form-item>
      <a-form-item label="密码">
        <a-input-password v-model:value="form.password" placeholder="输入 SSH 密码" />
      </a-form-item>
      <a-form-item>
        <a-button
          type="primary"
          :loading="loading"
          :class="{ 'save-ok': saved }"
          @click="handleOk"
          block
        >
          {{ saved ? '已保存 — 正在重连...' : '保存并连接' }}
        </a-button>
      </a-form-item>
    </a-form>

    <a-divider />

    <div class="security-hint">
      密码通过 AES 加密存储在本机，不会上传到任何服务器。
    </div>
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
.security-hint {
  font-size: 11px; color: #556677; text-align: center;
}
</style>
