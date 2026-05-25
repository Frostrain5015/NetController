<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircleFilled, CloseCircleFilled, GlobalOutlined, ApiOutlined, ClusterOutlined } from '@ant-design/icons-vue'
import FlagIcon from './FlagIcon.vue'

const props = defineProps<{ snapshot: Snapshot | null }>()

const projects = computed(() => props.snapshot?.projects ?? [])
const proxy = computed(() => props.snapshot?.proxy ?? null)
const proxyNodes = computed(() => props.snapshot?.proxyNodes ?? [])
const locatedNodes = computed(() => proxyNodes.value.filter(n => n.location !== null))
const unlocatedNodes = computed(() => proxyNodes.value.filter(n => n.location === null))

function isChild(p: ProjectStatus) { return !!(p as any).parent }
function childOf(parent: string) { return projects.value.filter(p => (p as any).parent === parent) }
</script>

<template>
  <!-- 项目状态 -->
  <div class="section">
    <div class="section-header"><ClusterOutlined /><span>项目状态</span></div>
    <div v-if="projects.length === 0" class="empty-text">暂无项目</div>
    <template v-for="p in projects" :key="p.name">
      <template v-if="!isChild(p)">
        <div class="status-row">
          <CheckCircleFilled v-if="p.alive" style="color: #52c41a; flex-shrink: 0" />
          <CloseCircleFilled v-else style="color: #ff4d4f; flex-shrink: 0" />
          <span class="status-name">{{ p.name }}</span>
          <a-tag v-if="p.alive && p.port" color="blue" size="small">:{{ p.port }}</a-tag>
          <a-tag v-if="p.alive" color="green" size="small">PID {{ p.pid }}</a-tag>
          <a-tag v-else color="red" size="small">未运行</a-tag>
        </div>
        <!-- 子进程 -->
        <div v-for="child in childOf(p.name)" :key="child.name" class="status-row child-row">
          <CheckCircleFilled v-if="child.alive" style="color: #52c41a; flex-shrink: 0; font-size: 10px" />
          <CloseCircleFilled v-else style="color: #ff4d4f; flex-shrink: 0; font-size: 10px" />
          <span class="status-name">{{ child.name }}</span>
          <a-tag v-if="child.alive && child.port" color="blue" size="small">:{{ child.port }}</a-tag>
          <a-tag v-if="child.alive" color="green" size="small">PID {{ child.pid }}</a-tag>
          <a-tag v-else color="red" size="small">未运行</a-tag>
        </div>
      </template>
    </template>
  </div>

  <!-- 代理 (Clash) -->
  <div class="section">
    <div class="section-header"><ApiOutlined /><span>Clash 代理</span></div>
    <div v-if="proxy" class="status-row">
      <CheckCircleFilled v-if="proxy.alive" style="color: #52c41a; flex-shrink: 0" />
      <CloseCircleFilled v-else style="color: #ff4d4f; flex-shrink: 0" />
      <span class="status-name">{{ proxy.name || 'Clash' }}</span>
      <a-tag v-if="proxy.port" color="blue" size="small">:{{ proxy.port }}</a-tag>
      <a-tag v-if="proxy.alive" color="green" size="small">运行中</a-tag>
      <a-tag v-else color="red" size="small">已停止</a-tag>
      <a-tag v-if="proxy.alive && proxy.apiAccessible" color="blue" size="small">API 可达</a-tag>
      <a-tag v-else-if="proxy.alive" color="orange" size="small">API 不可达</a-tag>
    </div>
    <div v-else class="empty-text">未配置</div>
    <!-- 流量进度条 -->
    <div v-if="proxy?.trafficRemainingGB != null" class="traffic-bar">
      <div class="traffic-label">
        <span>剩余流量</span>
        <span>{{ proxy.trafficRemainingGB.toFixed(1) }} / 200 GB</span>
      </div>
      <a-progress
        :percent="Math.min((proxy.trafficRemainingGB / 200) * 100, 100)"
        :stroke-color="proxy.trafficRemainingGB > 40 ? '#52c41a' : proxy.trafficRemainingGB > 10 ? '#faad14' : '#ff4d4f'"
        :show-info="false"
        size="small"
      />
    </div>
    <!-- 套餐到期 -->
    <div v-if="proxy?.planExpiry" class="expiry-row">
      <span>套餐到期</span>
      <span class="expiry-date">{{ proxy.planExpiry }}</span>
    </div>
  </div>

  <!-- 代理节点 -->
  <div class="section">
    <div class="section-header"><GlobalOutlined /><span>代理节点</span></div>
    <div v-if="proxyNodes.length === 0" class="empty-text">
      {{ proxy && proxy.alive ? '未获取到代理节点（Clash API 未响应）' : '代理未运行' }}
    </div>
    <template v-for="n in locatedNodes" :key="n.name">
      <div class="status-row">
        <span class="node-dot" :class="n.reachable ? 'reachable' : 'unreachable'"></span>
        <FlagIcon :code="n.country" />
        <span class="status-name">{{ n.displayName || n.name }}</span>
        <a-tag v-if="n.reachable" color="green" size="small">{{ n.latencyMs }}ms</a-tag>
        <a-tag v-else-if="(n as any).tested" color="red" size="small">超时</a-tag>
        <a-tag v-else color="default" size="small">检测中</a-tag>
      </div>
    </template>
    <div v-if="unlocatedNodes.length > 0" class="sub-header">未定位节点</div>
    <div v-for="n in unlocatedNodes" :key="n.name" class="status-row">
      <span class="node-dot" :class="n.reachable ? 'reachable' : 'unreachable'"></span>
      <a-tag color="purple" size="small">{{ n.type }}</a-tag>
      <span class="status-name">{{ n.displayName || n.name }}</span>
      <a-tag v-if="n.reachable" color="green" size="small">{{ n.latencyMs }}ms</a-tag>
      <a-tag v-else-if="(n as any).tested" color="red" size="small">超时</a-tag>
      <a-tag v-else color="default" size="small">检测中</a-tag>
    </div>
  </div>
</template>

<style scoped>
.section { margin-bottom: 20px; }
.section-header {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px; font-weight: 600; color: #8899aa;
  margin-bottom: 10px; letter-spacing: 0.5px;
}
.status-row {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 10px; margin-bottom: 4px;
  background: #141f2b; border-radius: 6px;
  border: 1px solid #1f2d3d; font-size: 13px;
}
.child-row {
  margin-left: 18px; background: #111c26;
}
.status-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.empty-text { color: #556677; font-size: 12px; padding: 4px 0; }
.sub-header { font-size: 11px; color: #556677; padding: 8px 0 4px 0; border-top: 1px solid #1f2d3d; margin-top: 4px; }
.traffic-bar { padding: 6px 10px; background: #141f2b; border-radius: 6px; border: 1px solid #1f2d3d; margin-bottom: 4px; }
.traffic-label { display: flex; justify-content: space-between; font-size: 11px; color: #8899aa; margin-bottom: 4px; }
.expiry-row { display: flex; justify-content: space-between; align-items: center; padding: 6px 10px; background: #141f2b; border-radius: 6px; border: 1px solid #1f2d3d; margin-bottom: 4px; font-size: 12px; color: #8899aa; }
.expiry-date { color: #e0e0e0; }
.node-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.node-dot.reachable { background: #52c41a; box-shadow: 0 0 6px #52c41a66; }
.node-dot.unreachable { background: #ff4d4f; box-shadow: 0 0 6px #ff4d4f66; }
</style>
