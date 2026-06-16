<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Minus, Square, X, Globe } from 'lucide-vue-next'
import { useAgentData } from './composables/useWebSocket'
import { useGeoLocation } from './composables/useGeoLocation'
import ServerMetrics from './components/ServerMetrics.vue'
import StatusSidebar from './components/StatusSidebar.vue'
import ChinaMap from './components/ChinaMap.vue'
import WorldMap from './components/WorldMap.vue'
import SettingsModal from './components/SettingsModal.vue'
import UpdateBanner from './components/UpdateBanner.vue'

const { snapshot, connected, selectProxyNode } = useAgentData()

// 把连接状态同步给主进程，驱动托盘图标/提示
watch(connected, (s) => window.electronAPI?.reportConnection(s), { immediate: true })
const { location: myLocation } = useGeoLocation()
const mapTab = ref<'china' | 'world'>('china')

// 判断本机是否也在杭州（同城则隐藏连线）
const sameCity = computed(() => {
  if (!myLocation.value) return false
  return myLocation.value.city === 'Hangzhou' || myLocation.value.city === '杭州'
})

const connMeta = computed(() => {
  switch (connected.value) {
    case 'connected': return { label: 'LIVE', cls: 'ok' }
    case 'connecting': return { label: 'CONNECTING', cls: 'warn' }
    default: return { label: 'OFFLINE', cls: 'bad' }
  }
})

function winMinimize() { window.electronAPI?.minimize() }
function winMaximize() { window.electronAPI?.maximize() }
function winClose() { window.electronAPI?.close() }
</script>

<template>
  <div class="app-bg"></div>
  <div class="app-shell">
    <!-- ── 标题栏 ───────────────────────────── -->
    <header class="title-bar">
      <div class="title-left">
        <span class="brand-mark"></span>
        <span class="brand-text">NET<b>CONTROLLER</b></span>
        <span class="brand-sep"></span>
        <span class="conn-badge" :class="connMeta.cls">
          <span class="conn-dot"></span>{{ connMeta.label }}
        </span>
      </div>
      <div class="title-right">
        <SettingsModal />
        <span class="win-ctrl" @click="winMinimize"><Minus :size="15" /></span>
        <span class="win-ctrl" @click="winMaximize"><Square :size="12" /></span>
        <span class="win-ctrl win-close" @click="winClose"><X :size="15" /></span>
      </div>
    </header>

    <!-- ── 主体 ─────────────────────────────── -->
    <div class="main-layout">
      <aside class="app-sider">
        <div class="sider-scroll">
          <ServerMetrics :metrics="snapshot?.serverMetrics ?? null" />
          <StatusSidebar :snapshot="snapshot" @select-proxy-node="selectProxyNode" />
        </div>
      </aside>

      <main class="map-content">
        <div class="map-toolbar">
          <div class="seg">
            <button
              class="seg-btn" :class="{ active: mapTab === 'china' }"
              @click="mapTab = 'china'"
            >中国</button>
            <button
              class="seg-btn" :class="{ active: mapTab === 'world' }"
              @click="mapTab = 'world'"
            ><Globe :size="13" /> 世界</button>
          </div>
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
      </main>
    </div>

    <UpdateBanner />
  </div>
</template>

<style>
.app-shell {
  position: relative;
  z-index: 1;
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

/* ── Title bar ── */
.title-bar {
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 8px 0 16px;
  background: linear-gradient(180deg, rgba(10, 16, 24, 0.92), rgba(7, 11, 17, 0.92));
  border-bottom: 1px solid var(--line);
  backdrop-filter: blur(8px);
  -webkit-app-region: drag;
  flex-shrink: 0;
  user-select: none;
}
.title-left { display: flex; align-items: center; gap: 12px; }
.brand-mark {
  width: 11px; height: 11px;
  background: var(--accent);
  box-shadow: 0 0 10px var(--accent-glow);
  transform: rotate(45deg);
  border-radius: 2px;
}
.brand-text {
  font-size: 13px; font-weight: 500; letter-spacing: 2.5px; color: var(--text-dim);
}
.brand-text b { color: var(--text); font-weight: 700; }
.brand-sep { width: 1px; height: 16px; background: var(--line-strong); }

.conn-badge {
  display: inline-flex; align-items: center; gap: 7px;
  font-size: 10.5px; font-weight: 600; letter-spacing: 1.5px;
  padding: 3px 10px 3px 9px;
  border-radius: 100px;
  border: 1px solid transparent;
}
.conn-dot { width: 7px; height: 7px; border-radius: 50%; }
.conn-badge.ok { color: var(--ok); background: rgba(61, 220, 151, 0.10); border-color: rgba(61, 220, 151, 0.25); }
.conn-badge.ok .conn-dot { background: var(--ok); box-shadow: 0 0 8px var(--ok); animation: nc-pulse 1.8s ease-in-out infinite; }
.conn-badge.warn { color: var(--warn); background: rgba(245, 185, 66, 0.10); border-color: rgba(245, 185, 66, 0.25); }
.conn-badge.warn .conn-dot { background: var(--warn); box-shadow: 0 0 8px var(--warn); animation: nc-pulse 0.9s ease-in-out infinite; }
.conn-badge.bad { color: var(--bad); background: rgba(255, 93, 108, 0.10); border-color: rgba(255, 93, 108, 0.25); }
.conn-badge.bad .conn-dot { background: var(--bad); box-shadow: 0 0 8px var(--bad); }

.title-right { display: flex; align-items: center; gap: 2px; -webkit-app-region: no-drag; }
.win-ctrl {
  display: inline-flex; align-items: center; justify-content: center;
  width: 34px; height: 28px; color: var(--text-dim);
  border-radius: 6px; cursor: pointer; transition: all 0.15s;
}
.win-ctrl:hover { color: var(--text); background: rgba(120, 168, 210, 0.10); }
.win-ctrl.win-close:hover { color: #fff; background: var(--bad); }

/* ── Layout ── */
.main-layout { flex: 1; display: flex; overflow: hidden; }
.app-sider {
  width: 312px; flex-shrink: 0;
  background: linear-gradient(180deg, rgba(11, 17, 26, 0.55), rgba(7, 11, 17, 0.35));
  border-right: 1px solid var(--line);
}
.sider-scroll { height: 100%; overflow-y: auto; padding: 18px 16px 24px; }

/* ── Map area ── */
.map-content { position: relative; flex: 1; overflow: hidden; }
.map-toolbar { position: absolute; top: 16px; left: 50%; transform: translateX(-50%); z-index: 10; }
.map-container { width: 100%; height: 100%; }

.seg {
  display: flex; gap: 3px; padding: 3px;
  background: rgba(8, 13, 20, 0.72);
  border: 1px solid var(--line-strong);
  border-radius: 100px;
  backdrop-filter: blur(10px);
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.4);
}
.seg-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 16px;
  font-family: var(--ui); font-size: 12.5px; font-weight: 600; letter-spacing: 1px;
  color: var(--text-dim); background: transparent;
  border: none; border-radius: 100px; cursor: pointer;
  transition: all 0.2s;
}
.seg-btn:hover { color: var(--text); }
.seg-btn.active {
  color: var(--bg-deep);
  background: var(--accent);
  box-shadow: 0 0 16px var(--accent-glow);
}
</style>
