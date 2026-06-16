<script setup lang="ts">
import { computed } from 'vue'
import { Boxes, Waypoints, Globe } from 'lucide-vue-next'
import FlagIcon from './FlagIcon.vue'

const props = defineProps<{ snapshot: Snapshot | null }>()

const projects = computed(() => props.snapshot?.projects ?? [])
const proxy = computed(() => props.snapshot?.proxy ?? null)
const proxyNodes = computed(() => props.snapshot?.proxyNodes ?? [])
const locatedNodes = computed(() => proxyNodes.value.filter(n => n.location !== null))
const unlocatedNodes = computed(() => proxyNodes.value.filter(n => n.location === null))
const rootProjects = computed(() => projects.value.filter(p => !p.parent))

function childOf(parent: string) { return projects.value.filter(p => p.parent === parent) }

const trafficPct = computed(() => {
  const g = proxy.value?.trafficRemainingGB
  if (g == null) return 0
  return Math.min((g / 200) * 100, 100)
})
const trafficTone = computed(() => {
  const g = proxy.value?.trafficRemainingGB ?? 0
  return g > 40 ? 'ok' : g > 10 ? 'warn' : 'bad'
})
</script>

<template>
  <!-- ── 项目状态 ── -->
  <div class="section nc-rise" style="animation-delay: 80ms">
    <div class="section-head">
      <Boxes :size="14" />
      <span class="eyebrow">项目状态</span>
      <span class="count">{{ rootProjects.length }}</span>
    </div>
    <div v-if="projects.length === 0" class="empty">暂无项目</div>
    <template v-for="p in rootProjects" :key="p.name">
      <div class="row">
        <span class="stat-dot" :class="p.alive ? 'ok' : 'bad'"></span>
        <span class="row-name">{{ p.name }}</span>
        <span class="spacer"></span>
        <span v-if="p.alive && p.port" class="chip accent">:{{ p.port }}</span>
        <span v-if="p.alive" class="chip muted">PID {{ p.pid }}</span>
        <span v-else class="chip bad">DOWN</span>
      </div>
      <div v-for="child in childOf(p.name)" :key="child.name" class="row child">
        <span class="stat-dot sm" :class="child.alive ? 'ok' : 'bad'"></span>
        <span class="row-name">{{ child.name }}</span>
        <span class="spacer"></span>
        <span v-if="child.alive && child.port" class="chip accent">:{{ child.port }}</span>
        <span v-if="child.alive" class="chip muted">PID {{ child.pid }}</span>
        <span v-else class="chip bad">DOWN</span>
      </div>
    </template>
  </div>

  <!-- ── Clash 代理 ── -->
  <div class="section nc-rise" style="animation-delay: 140ms">
    <div class="section-head">
      <Waypoints :size="14" />
      <span class="eyebrow">Clash 代理</span>
    </div>
    <div v-if="proxy" class="row">
      <span class="stat-dot" :class="proxy.alive ? 'ok' : 'bad'"></span>
      <span class="row-name">{{ proxy.name || 'Clash' }}</span>
      <span class="spacer"></span>
      <span v-if="proxy.port" class="chip accent">:{{ proxy.port }}</span>
      <span v-if="proxy.alive && proxy.apiAccessible" class="chip ok">API</span>
      <span v-else-if="proxy.alive" class="chip warn">NO API</span>
      <span v-else class="chip bad">DOWN</span>
    </div>
    <div v-else class="empty">未配置</div>

    <!-- 流量 -->
    <div v-if="proxy?.trafficRemainingGB != null" class="meter">
      <div class="meter-head">
        <span>剩余流量</span>
        <span class="mono">{{ proxy.trafficRemainingGB.toFixed(1) }} / 200 GB</span>
      </div>
      <div class="meter-track">
        <div class="meter-fill" :class="trafficTone" :style="{ width: trafficPct + '%' }"></div>
      </div>
    </div>
    <!-- 套餐到期 -->
    <div v-if="proxy?.planExpiry" class="kv">
      <span>套餐到期</span>
      <span class="mono kv-val">{{ proxy.planExpiry }}</span>
    </div>
  </div>

  <!-- ── 代理节点 ── -->
  <div class="section nc-rise" style="animation-delay: 200ms">
    <div class="section-head">
      <Globe :size="14" />
      <span class="eyebrow">代理节点</span>
      <span v-if="proxyNodes.length" class="count">{{ proxyNodes.length }}</span>
    </div>
    <div v-if="proxyNodes.length === 0" class="empty">
      {{ proxy && proxy.alive ? '未获取到代理节点（Clash API 未响应）' : '代理未运行' }}
    </div>

    <div v-for="n in locatedNodes" :key="n.name" class="row">
      <span class="node-dot" :class="n.reachable ? 'on' : 'off'"></span>
      <FlagIcon :code="n.country" />
      <span class="row-name">{{ n.displayName || n.name }}</span>
      <span class="spacer"></span>
      <span v-if="n.reachable" class="chip ok mono">{{ n.latencyMs }}ms</span>
      <span v-else-if="(n as any).tested" class="chip bad">超时</span>
      <span v-else class="chip muted">检测中</span>
    </div>

    <div v-if="unlocatedNodes.length > 0" class="sub-head">未定位节点</div>
    <div v-for="n in unlocatedNodes" :key="n.name" class="row">
      <span class="node-dot" :class="n.reachable ? 'on' : 'off'"></span>
      <span class="chip ghost">{{ n.type }}</span>
      <span class="row-name">{{ n.displayName || n.name }}</span>
      <span class="spacer"></span>
      <span v-if="n.reachable" class="chip ok mono">{{ n.latencyMs }}ms</span>
      <span v-else-if="(n as any).tested" class="chip bad">超时</span>
      <span v-else class="chip muted">检测中</span>
    </div>
  </div>
</template>

<style scoped>
.section { margin-bottom: 24px; }
.section-head {
  display: flex; align-items: center; gap: 8px;
  margin-bottom: 12px;
  color: var(--accent);
}
.section-head .eyebrow { flex: 1; }
.count {
  font-family: var(--mono); font-size: 10px; font-weight: 700;
  color: var(--text-dim);
  padding: 1px 7px; border-radius: 100px;
  background: rgba(120, 168, 210, 0.08); border: 1px solid var(--line);
}

.row {
  display: flex; align-items: center; gap: 9px;
  padding: 9px 11px; margin-bottom: 5px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  font-size: 13px;
  transition: border-color 0.18s, background 0.18s;
}
.row:hover { background: var(--panel-hover); border-color: var(--line-strong); }
.row.child { margin-left: 16px; background: var(--bg-deep); }
.row-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text); }
.spacer { flex: 1; }

/* Status dots */
.stat-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.stat-dot.sm { width: 6px; height: 6px; }
.stat-dot.ok { background: var(--ok); box-shadow: 0 0 8px rgba(61, 220, 151, 0.7); }
.stat-dot.bad { background: var(--bad); box-shadow: 0 0 8px rgba(255, 93, 108, 0.6); }
.node-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.node-dot.on { background: var(--ok); box-shadow: 0 0 7px rgba(61, 220, 151, 0.7); }
.node-dot.off { background: var(--text-faint); }

/* Chips */
.chip {
  flex-shrink: 0;
  font-size: 10.5px; font-weight: 600; letter-spacing: 0.4px;
  padding: 2px 7px; border-radius: 5px;
  border: 1px solid transparent; line-height: 1.5;
}
.chip.mono { font-family: var(--mono); font-weight: 500; letter-spacing: 0; }
.chip.accent { color: var(--accent); background: var(--accent-dim); border-color: rgba(54, 211, 238, 0.22); }
.chip.ok { color: var(--ok); background: rgba(61, 220, 151, 0.12); border-color: rgba(61, 220, 151, 0.25); }
.chip.warn { color: var(--warn); background: rgba(245, 185, 66, 0.12); border-color: rgba(245, 185, 66, 0.25); }
.chip.bad { color: var(--bad); background: rgba(255, 93, 108, 0.12); border-color: rgba(255, 93, 108, 0.25); }
.chip.muted { color: var(--text-dim); background: rgba(120, 168, 210, 0.06); border-color: var(--line); }
.chip.ghost { color: var(--text-dim); background: transparent; border-color: var(--line-strong); }

/* Empty / sub */
.empty { color: var(--text-faint); font-size: 12px; padding: 2px 2px 4px; }
.sub-head {
  font-size: 10px; font-weight: 600; letter-spacing: 1.5px; text-transform: uppercase;
  color: var(--text-faint); padding: 10px 2px 6px;
  border-top: 1px solid var(--line); margin-top: 6px;
}

/* Traffic meter */
.meter {
  padding: 10px 11px; margin-top: 5px;
  background: var(--panel); border: 1px solid var(--line); border-radius: var(--radius-sm);
}
.meter-head {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 11px; color: var(--text-dim); margin-bottom: 8px;
}
.meter-head .mono { font-family: var(--mono); color: var(--text); }
.meter-track { height: 5px; border-radius: 100px; background: rgba(120, 168, 210, 0.10); overflow: hidden; }
.meter-fill {
  height: 100%; border-radius: 100px;
  transition: width 0.7s cubic-bezier(0.22, 1, 0.36, 1);
}
.meter-fill.ok { background: linear-gradient(90deg, #2bbf87, var(--ok)); box-shadow: 0 0 10px rgba(61, 220, 151, 0.5); }
.meter-fill.warn { background: linear-gradient(90deg, #d99a2a, var(--warn)); box-shadow: 0 0 10px rgba(245, 185, 66, 0.5); }
.meter-fill.bad { background: linear-gradient(90deg, #e0455a, var(--bad)); box-shadow: 0 0 10px rgba(255, 93, 108, 0.5); }

/* Key-value */
.kv {
  display: flex; justify-content: space-between; align-items: center;
  padding: 9px 11px; margin-top: 5px;
  background: var(--panel); border: 1px solid var(--line); border-radius: var(--radius-sm);
  font-size: 12px; color: var(--text-dim);
}
.kv-val { font-family: var(--mono); color: var(--text); }
</style>
