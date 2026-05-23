<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted, shallowRef } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{
  overseasNodes: OverseasNode[]
  connected: string
  myLocation: { lat: number; lng: number; country: string; region: string; city: string } | null
  sameCity: boolean
}>()

const container = ref<HTMLDivElement>()
const chart = shallowRef<echarts.ECharts | null>(null)
let resizeObs: ResizeObserver | null = null
const showOverseasLines = ref(false)
let animTimer: ReturnType<typeof setTimeout> | null = null

const connColor = computed(() => {
  if (props.connected === 'connecting') return '#faad14'
  return props.connected === 'connected' ? '#52c41a' : '#ff4d4f'
})

function buildOption(showOverseas: boolean): echarts.EChartsOption {
  const scatterData = props.overseasNodes.map(n => ({
    name: n.name, value: [...n.location, n.reachable, n.latencyMs],
  }))
  const hz = [120.15, 30.28]
  const showPcLine = !props.sameCity && props.myLocation
  const myCountry = props.myLocation?.country || ''

  const regions: any[] = [
    { name: 'China', itemStyle: { areaColor: '#1a3050', borderColor: '#1677ff', borderWidth: 1 } },
  ]
  // 本机所在国家也高亮（如果不同于中国）
  if (myCountry && myCountry !== 'China') {
    regions.push({ name: myCountry, itemStyle: { areaColor: '#162d40', borderColor: '#13c2c2', borderWidth: 1 } })
  }

  const series: any[] = [
    // 杭州
    {
      type: 'effectScatter', coordinateSystem: 'geo',
      data: [{ name: '杭州', value: [...hz, props.connected !== 'disconnected' ? 1 : 0] }],
      symbolSize: 14,
      rippleEffect: { brushType: 'stroke', scale: 2.5, period: 3 },
      label: { show: true, formatter: '杭州', position: 'bottom', color: connColor.value, fontSize: 13, fontWeight: 'bold', offset: [0, 8] },
      itemStyle: { color: connColor.value, shadowBlur: 10, shadowColor: connColor.value },
    },
    // 海外节点
    {
      type: 'effectScatter', coordinateSystem: 'geo',
      data: scatterData,
      symbolSize: (val: any) => Math.max(4, Math.min(10, (val[3] > 0 ? val[3] : 300) / 30)),
      label: { formatter: '{b}', position: 'right', show: true, color: '#8899aa', fontSize: 11 },
      itemStyle: { color: (p: any) => (p.value[2] ? '#52c41a' : '#ff4d4f'), shadowBlur: 6, shadowColor: 'rgba(0,0,0,0.5)' },
    },
  ]

  if (props.myLocation) {
    series.push({
      type: 'effectScatter', coordinateSystem: 'geo',
      data: [{ name: '本机', value: [props.myLocation.lng, props.myLocation.lat] }],
      symbolSize: 10, symbol: 'pin',
      rippleEffect: { brushType: 'stroke', scale: 2, period: 2.5 },
      label: { show: true, formatter: '本机', position: 'top', color: '#13c2c2', fontSize: 11, offset: [0, -6] },
      itemStyle: { color: '#13c2c2', shadowBlur: 8, shadowColor: '#13c2c266' },
      zlevel: 1,
    })
  }

  if (showPcLine) {
    series.push({
      type: 'lines', coordinateSystem: 'geo', polyline: false,
      animationDelay: 0,
      data: [{ coords: [[props.myLocation!.lng, props.myLocation!.lat], hz], lineStyle: { color: '#13c2c2', width: 1.5, type: 'dashed' } }],
      effect: { show: true, period: 4, trailLength: 0.2, symbol: 'circle', symbolSize: 4, color: '#13c2c2' },
    })
  }

  if (showOverseas) {
    series.push({
      type: 'lines', coordinateSystem: 'geo', polyline: false,
      animationDelay: 800,
      data: props.overseasNodes.filter(n => n.reachable).map(n => ({
        coords: [hz, n.location], lineStyle: { color: '#52c41a33', width: 1 },
      })),
      effect: { show: true, period: 6, trailLength: 0.3, symbol: 'arrow', symbolSize: 6, color: '#1677ff' },
    })
  }

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      formatter: (p: any) => {
        if (p.seriesType === 'scatter' || p.seriesType === 'effectScatter') {
          if (p.name === '杭州') return `<b>杭州</b><br/>${props.connected === 'connecting' ? '正在连接...' : props.connected === 'connected' ? '服务器已连接' : '服务器未连接'}`
          if (p.name === '本机') return `<b>本机</b><br/>${props.myLocation?.city || ''} ${props.myLocation?.country || ''}`
          const [lng, lat, reachable, ms] = p.value ?? []
          return `<b>${p.name}</b><br/>${reachable ? `可达 ${ms}ms` : '不可达'}`
        }
        return `${p.name}`
      },
    },
    geo: {
      map: 'world', roam: true, zoom: 1.2, center: [20, 25],
      itemStyle: { areaColor: '#141f2b', borderColor: '#2a3a4a', borderWidth: 0.5 },
      emphasis: { itemStyle: { areaColor: '#1a3a5c' }, label: { color: '#fff' } },
      regions,
    },
    series,
  }
}

function render(showOverseas: boolean) {
  if (!chart.value) return
  chart.value.setOption(buildOption(showOverseas), true)
}

watch(() => props.connected, (c, prev) => {
  if (animTimer) { clearTimeout(animTimer); animTimer = null }
  if (c === 'connected' && prev !== 'connected') {
    showOverseasLines.value = false
    render(false)
    animTimer = setTimeout(() => { showOverseasLines.value = true; render(true) }, 1500)
  } else {
    showOverseasLines.value = true
    render(true)
  }
})

onMounted(async () => {
  if (!container.value) return
  try { const r = await fetch('world.json'); echarts.registerMap('world', await r.json()) } catch { /* */ }
  chart.value = echarts.init(container.value)
  render(true)
  resizeObs = new ResizeObserver(() => chart.value?.resize())
  resizeObs.observe(container.value)
})

onUnmounted(() => { if (animTimer) clearTimeout(animTimer); resizeObs?.disconnect(); chart.value?.dispose() })
</script>

<template>
  <div ref="container" class="chart"></div>
</template>

<style scoped>
.chart { width: 100%; height: 100%; }
</style>
