<script setup lang="ts">
import { computed } from 'vue'
import { DashboardOutlined, CloudServerOutlined, DatabaseOutlined } from '@ant-design/icons-vue'

const props = defineProps<{ metrics: ServerMetrics | null }>()

const items = computed(() => {
  const m = props.metrics
  const memUsed = m?.memUsedMB != null ? (m.memUsedMB / 1024).toFixed(1) : '--'
  const memTotal = m?.memTotalMB != null ? (m.memTotalMB / 1024).toFixed(1) : '--'
  return [
    { icon: DashboardOutlined, label: 'CPU', color: '#1677ff', value: m ? `${m.cpuPercent}%` : '--', sub: '' },
    { icon: CloudServerOutlined, label: '内存', color: '#52c41a', value: m ? `${m.memPercent}%` : '--', sub: m ? `${memUsed}/${memTotal}G` : '' },
    { icon: DatabaseOutlined, label: '磁盘', color: '#fa8c16', value: m ? `${m.diskUsed}` : '--', sub: m ? `/ ${m.diskTotal}` : '' },
  ]
})
</script>

<template>
  <div class="metrics-cards">
    <div v-for="it in items" :key="it.label" class="card">
      <div class="card-top">
        <component :is="it.icon" :style="{ color: it.color, fontSize: '20px' }" />
        <span class="card-label">{{ it.label }}</span>
      </div>
      <div class="card-value" :style="{ color: it.color }">{{ it.value }}</div>
      <div v-if="it.sub" class="card-sub">{{ it.sub }}</div>
    </div>
  </div>
</template>

<style scoped>
.metrics-cards { display: flex; gap: 8px; margin-bottom: 16px; }
.card {
  flex: 1; min-width: 0;
  padding: 10px 12px; background: #141f2b; border-radius: 8px;
  border: 1px solid #1f2d3d;
}
.card-top { display: flex; align-items: center; gap: 6px; margin-bottom: 4px; }
.card-label { font-size: 11px; color: #667788; text-transform: uppercase; letter-spacing: 0.5px; }
.card-value {
  font-size: 21px; font-weight: 700; line-height: 1.2;
  font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
}
.card-sub {
  font-size: 11px; color: #556677; margin-top: 2px;
  font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
</style>
