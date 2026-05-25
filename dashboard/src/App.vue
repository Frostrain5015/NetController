<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAgentData } from './composables/useWebSocket'
import { useGeoLocation } from './composables/useGeoLocation'
import ServerMetrics from './components/ServerMetrics.vue'
import StatusSidebar from './components/StatusSidebar.vue'
import ChinaMap from './components/ChinaMap.vue'
import WorldMap from './components/WorldMap.vue'
import SettingsModal from './components/SettingsModal.vue'
import { MinusOutlined, BorderOutlined, CloseOutlined } from '@ant-design/icons-vue'

const { snapshot, connected } = useAgentData()
const { location: myLocation } = useGeoLocation()
const mapTab = ref<'china' | 'world'>('china')

// 判断本机是否也在杭州（同城则隐藏连线）
const sameCity = computed(() => {
  if (!myLocation.value) return false
  return myLocation.value.city === 'Hangzhou' || myLocation.value.city === '杭州'
})

function winMinimize() { window.electronAPI?.minimize() }
function winMaximize() { window.electronAPI?.maximize() }
function winClose() { window.electronAPI?.close() }
</script>

<template>
  <a-config-provider
    :theme="{
      algorithm: undefined,
      token: {
        colorPrimary: '#1677ff',
        colorBgBase: '#0f1923',
        colorTextBase: '#e0e0e0',
        colorBorder: '#1f2d3d',
        colorBgContainer: '#141f2b',
      },
    }"
  >
    <div class="app-shell">
      <!-- 自定义暗色标题栏 -->
      <div class="title-bar">
        <div class="title-left">
          <span class="title-icon">&#x25C8;</span>
          <span class="title-text">NetController</span>
        </div>
        <div class="title-center"></div>
        <div class="title-right">
          <SettingsModal />
          <span class="win-ctrl" @click="winMinimize"><MinusOutlined /></span>
          <span class="win-ctrl" @click="winMaximize"><BorderOutlined /></span>
          <span class="win-ctrl win-close" @click="winClose"><CloseOutlined /></span>
        </div>
      </div>

      <a-layout class="main-layout">
        <a-layout-sider width="300" class="app-sider">
          <div class="sider-scroll">
            <ServerMetrics :metrics="snapshot?.serverMetrics ?? null" />
            <StatusSidebar :snapshot="snapshot" />
          </div>
        </a-layout-sider>

        <a-layout-content class="map-content">
          <div class="map-toolbar">
            <a-segmented
              v-model:value="mapTab"
              :options="[
                { label: '中国', value: 'china' },
                { label: '世界', value: 'world' },
              ]"
            />
          </div>
          <div class="map-container">
            <ChinaMap
              v-if="mapTab === 'china'"
              :proxy-nodes="snapshot?.proxyNodes ?? []"
              :connected="connected"
              :my-location="myLocation"
              :same-city="sameCity"
            />
            <WorldMap
              v-else
              :proxy-nodes="snapshot?.proxyNodes ?? []"
              :connected="connected"
              :my-location="myLocation"
              :same-city="sameCity"
            />
          </div>
        </a-layout-content>
      </a-layout>
    </div>
  </a-config-provider>
</template>

<style>
.app-shell { width: 100vw; height: 100vh; display: flex; flex-direction: column; background: #0f1923; }
.title-bar {
  height: 36px; display: flex; align-items: center; justify-content: space-between;
  padding: 0 8px 0 14px;
  background: #0a1119; border-bottom: 1px solid #1f2d3d;
  -webkit-app-region: drag; flex-shrink: 0; user-select: none;
}
.title-left { display: flex; align-items: center; gap: 8px; }
.title-icon { color: #1677ff; font-size: 14px; }
.title-text { font-size: 13px; font-weight: 600; letter-spacing: 1px; }
.title-center { flex: 1; }
.title-right { display: flex; align-items: center; gap: 4px; -webkit-app-region: no-drag; }
.conn-tag { font-size: 11px; }
.conn-dot { display: inline-block; width: 6px; height: 6px; border-radius: 50%; margin-right: 4px; }
.win-ctrl {
  display: inline-flex; align-items: center; justify-content: center;
  width: 32px; height: 26px; color: #8899aa; font-size: 12px;
  border-radius: 4px; cursor: pointer;
}
.win-ctrl:hover { color: #e0e0e0; background: #1f2d3d; }
.win-ctrl.win-close:hover { color: #fff; background: #c0392b; }
.main-layout { flex: 1; overflow: hidden; background: transparent !important; }
.app-sider { background: #0f1923 !important; border-right: 1px solid #1f2d3d; }
.sider-scroll { height: 100%; overflow-y: auto; padding: 12px; }
.sider-scroll::-webkit-scrollbar { width: 4px; }
.sider-scroll::-webkit-scrollbar-thumb { background: #1f2d3d; border-radius: 2px; }
.map-content { position: relative; overflow: hidden; }
.map-toolbar { position: absolute; top: 12px; left: 50%; transform: translateX(-50%); z-index: 10; }
.map-container { width: 100%; height: 100%; }
</style>
