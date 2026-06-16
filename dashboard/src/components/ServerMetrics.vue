<script setup lang="ts">
import { computed } from 'vue'
import { Cpu, MemoryStick, HardDrive } from 'lucide-vue-next'

const props = defineProps<{ metrics: ServerMetrics | null }>()

interface Gauge {
  key: string
  label: string
  icon: typeof Cpu
  pct: number | null
  value: string
  sub: string
}

const RADIUS = 24
const CIRC = 2 * Math.PI * RADIUS

const gauges = computed<Gauge[]>(() => {
  const m = props.metrics
  const memUsed = m?.memUsedMB != null ? (m.memUsedMB / 1024).toFixed(1) : '--'
  const memTotal = m?.memTotalMB != null ? (m.memTotalMB / 1024).toFixed(1) : '--'
  return [
    { key: 'cpu', label: 'CPU', icon: Cpu, pct: m?.cpuPercent ?? null, value: m ? `${m.cpuPercent}` : '--', sub: 'LOAD' },
    { key: 'mem', label: 'MEM', icon: MemoryStick, pct: m?.memPercent ?? null, value: m ? `${m.memPercent}` : '--', sub: m ? `${memUsed}/${memTotal}G` : '' },
    { key: 'disk', label: 'DISK', icon: HardDrive, pct: m?.diskPercent ?? null, value: m ? `${m.diskPercent}` : '--', sub: m ? `${m.diskUsed}/${m.diskTotal}` : '' },
  ]
})

function tone(pct: number | null): string {
  if (pct == null) return 'idle'
  if (pct >= 90) return 'bad'
  if (pct >= 75) return 'warn'
  return 'ok'
}
function offset(pct: number | null): number {
  return CIRC * (1 - (pct ?? 0) / 100)
}
</script>

<template>
  <section class="metrics nc-rise">
    <div
      v-for="(g, i) in gauges"
      :key="g.key"
      class="gauge-card"
      :class="tone(g.pct)"
      :style="{ animationDelay: `${i * 60}ms` }"
    >
      <svg class="gauge" viewBox="0 0 60 60">
        <circle class="track" cx="30" cy="30" :r="RADIUS" />
        <circle
          class="arc" cx="30" cy="30" :r="RADIUS"
          :stroke-dasharray="CIRC"
          :stroke-dashoffset="offset(g.pct)"
          transform="rotate(-90 30 30)"
        />
      </svg>
      <div class="gauge-center">
        <component :is="g.icon" :size="14" class="gauge-icon" />
        <div class="gauge-val">{{ g.value }}<small v-if="g.pct != null">%</small></div>
      </div>
      <div class="gauge-meta">
        <span class="gauge-label">{{ g.label }}</span>
        <span v-if="g.sub" class="gauge-sub">{{ g.sub }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.metrics { display: flex; gap: 10px; margin-bottom: 22px; }
.gauge-card {
  position: relative;
  flex: 1; min-width: 0;
  display: flex; flex-direction: column; align-items: center;
  padding: 14px 6px 10px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  transition: border-color 0.2s, transform 0.2s;
}
.gauge-card:hover { transform: translateY(-2px); border-color: var(--line-strong); }

.gauge { width: 60px; height: 60px; }
.gauge .track { fill: none; stroke: rgba(120, 168, 210, 0.10); stroke-width: 4; }
.gauge .arc {
  fill: none; stroke-width: 4; stroke-linecap: round;
  transition: stroke-dashoffset 0.7s cubic-bezier(0.22, 1, 0.36, 1), stroke 0.3s;
}

.gauge-center {
  position: absolute; top: 14px; left: 0; right: 0; height: 60px;
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 1px;
  pointer-events: none;
}
.gauge-icon { color: var(--text-faint); }
.gauge-val {
  font-family: var(--mono); font-weight: 700; font-size: 17px; line-height: 1;
  color: var(--text);
}
.gauge-val small { font-size: 9px; color: var(--text-dim); margin-left: 1px; font-weight: 500; }

.gauge-meta { display: flex; flex-direction: column; align-items: center; gap: 1px; margin-top: 7px; }
.gauge-label { font-size: 10px; font-weight: 600; letter-spacing: 2px; color: var(--text-dim); }
.gauge-sub {
  font-family: var(--mono); font-size: 9.5px; color: var(--text-faint);
  max-width: 92px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

/* Tone-driven arc + glow */
.gauge-card.ok   .arc { stroke: var(--accent); }
.gauge-card.ok   .gauge-icon { color: var(--accent); }
.gauge-card.warn .arc { stroke: var(--warn); }
.gauge-card.warn .gauge-icon { color: var(--warn); }
.gauge-card.bad  .arc { stroke: var(--bad); }
.gauge-card.bad  .gauge-icon { color: var(--bad); }
.gauge-card.bad  { border-color: rgba(255, 93, 108, 0.30); }
.gauge-card.idle .arc { stroke: rgba(120, 168, 210, 0.15); }
</style>
