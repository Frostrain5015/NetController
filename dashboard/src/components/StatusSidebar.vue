<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircleFilled, CloseCircleFilled, GlobalOutlined, ApiOutlined, ClusterOutlined } from '@ant-design/icons-vue'
import FlagIcon from './FlagIcon.vue'

const props = defineProps<{ snapshot: Snapshot | null }>()

const projects = computed(() => props.snapshot?.projects ?? [])
const proxy = computed(() => props.snapshot?.proxy ?? null)
const overseasNodes = computed(() => props.snapshot?.overseasNodes ?? [])

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
    </div>
    <div v-else class="empty-text">未配置</div>
  </div>

  <!-- 海外节点 -->
  <div class="section">
    <div class="section-header"><GlobalOutlined /><span>海外节点</span></div>
    <div v-if="overseasNodes.length === 0" class="empty-text">暂无节点</div>
    <div v-for="n in overseasNodes" :key="n.name" class="status-row">
      <span class="node-dot" :class="n.reachable ? 'reachable' : 'unreachable'"></span>
      <FlagIcon :code="n.country" />
      <span class="status-name">{{ n.name }}</span>
      <a-tag v-if="n.reachable" color="green" size="small">{{ n.latencyMs }}ms</a-tag>
      <a-tag v-else color="red" size="small">不可达</a-tag>
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
.node-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
}
.node-dot.reachable { background: #52c41a; box-shadow: 0 0 6px #52c41a66; }
.node-dot.unreachable { background: #ff4d4f; box-shadow: 0 0 6px #ff4d4f66; }
</style>
