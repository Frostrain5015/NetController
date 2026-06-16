<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted, shallowRef } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{
  proxyNodes: ProxyNode[]
  connected: string
  myLocation: { lat: number; lng: number; country: string; region: string; city: string } | null
  sameCity: boolean
}>()

const container = ref<HTMLDivElement>()
const chart = shallowRef<echarts.ECharts | null>(null)
let resizeObs: ResizeObserver | null = null
const showOverseasLines = ref(false)
let animTimer: ReturnType<typeof setTimeout> | null = null
const serverLocation: [number, number] = [120.15, 30.28]
const serverRegionColor = '#1677ff'
const localRegionColor = '#13c2c2'
const activeProxyColor = '#ffb454'
const nodeReachableColor = '#6b8cae'
const nodeUnreachableColor = '#8a4a5a'

type CountryGroup = {
  displayName: string
  location: [number, number]
  country: string
  reachable: boolean
  selected: boolean
  minLatency: number
  count: number
}

const connColor = computed(() => {
  if (props.connected === 'connecting') return '#faad14'
  return props.connected === 'connected' ? '#52c41a' : '#ff4d4f'
})

const nodeGroups = computed(() => {
  const map = new Map<string, CountryGroup>()
  for (const n of props.proxyNodes) {
    if (!n.location) continue
    const key = n.country
    const g = map.get(key)
    if (g) {
      if (n.reachable) { g.reachable = true; g.minLatency = Math.min(g.minLatency, n.latencyMs || Infinity) }
      if (n.selected) {
        g.selected = true
        g.displayName = n.displayName || n.name
        g.location = n.location
      }
      g.count++
    } else {
      map.set(key, {
        displayName: n.displayName || n.name,
        location: n.location, country: n.country,
        reachable: n.reachable,
        selected: n.selected,
        minLatency: n.reachable && n.latencyMs > 0 ? n.latencyMs : Infinity,
        count: 1,
      })
    }
  }
  return [...map.values()]
})

const activeGroup = computed(() => nodeGroups.value.find(g => g.selected) ?? null)

function worldRegionName(country: string) {
  const names: Record<string, string> = {
    HK: 'Hong Kong S.A.R.',
    JP: 'Japan',
    SG: 'Singapore',
    US: 'United States of America',
    TW: 'Taiwan',
    KR: 'South Korea',
    DE: 'Germany',
    GB: 'United Kingdom',
    FR: 'France',
    IN: 'India',
    CA: 'Canada',
    AU: 'Australia',
    NL: 'Netherlands',
    RU: 'Russia',
    BR: 'Brazil',
    VN: 'Vietnam',
    TH: 'Thailand',
    PH: 'Philippines',
    MY: 'Malaysia',
    ID: 'Indonesia',
    TR: 'Turkey',
    AE: 'United Arab Emirates',
    ZA: 'South Africa',
    AR: 'Argentina',
    IT: 'Italy',
    ES: 'Spain',
    PL: 'Poland',
    SE: 'Sweden',
    CH: 'Switzerland',
    FI: 'Finland',
  }
  return names[country] || country
}

function buildOption(showOverseas: boolean): echarts.EChartsOption {
  const hz = serverLocation
  const groups = nodeGroups.value
  const currentGroup = activeGroup.value
  const showPcLine = !!props.myLocation
  const myCountry = props.myLocation?.country || ''
  const activeRegionName = currentGroup ? worldRegionName(currentGroup.country) : ''

  const scatterData = groups.map(g => ({
    name: g.displayName,
    value: [...g.location, g.reachable ? 1 : 0, g.minLatency < Infinity ? g.minLatency : 9999, g.count, g.selected ? 1 : 0],
  }))

  const regions: any[] = [
    { name: 'China', itemStyle: { areaColor: '#1a3050', borderColor: serverRegionColor, borderWidth: 1 } },
  ]
  if (myCountry && myCountry !== 'China') {
    regions.push({ name: myCountry, itemStyle: { areaColor: '#162d40', borderColor: localRegionColor, borderWidth: 1 } })
  }
  if (activeRegionName) {
    regions.push({ name: activeRegionName, itemStyle: { areaColor: '#3a2b19', borderColor: activeProxyColor, borderWidth: 1.4 } })
  }

  const series: any[] = [
    {
      id: 'hz', type: 'effectScatter', coordinateSystem: 'geo',
      data: [{ name: '杭州', value: [...hz, props.connected !== 'disconnected' ? 1 : 0] }],
      symbolSize: 14,
      rippleEffect: { brushType: 'stroke', scale: 2.5, period: 3 },
      label: { show: true, formatter: '杭州', position: 'top', color: connColor.value, fontSize: 13, fontWeight: 'bold', offset: [0, -10] },
      itemStyle: { color: connColor.value, shadowBlur: 10, shadowColor: connColor.value },
    },
    {
      id: 'groups', type: 'effectScatter', coordinateSystem: 'geo',
      data: scatterData,
      symbolSize: (value: any[]) => (value?.[5] ? 13 : 8),
      label: { formatter: '{b}', position: 'right', show: true, color: (p: any) => (p.value?.[5] ? activeProxyColor : '#8899aa'), fontSize: 12 },
      itemStyle: {
        color: (p: any) => (p.value?.[5] ? activeProxyColor : p.value?.[2] ? nodeReachableColor : nodeUnreachableColor),
        shadowBlur: (p: any) => (p.value?.[5] ? 14 : 6),
        shadowColor: (p: any) => (p.value?.[5] ? activeProxyColor : 'rgba(0,0,0,0.5)'),
      },
    },
  ]

  if (props.myLocation) {
    const offset = props.sameCity ? 0.3 : 0
    series.push({
      id: 'me', type: 'effectScatter', coordinateSystem: 'geo',
      data: [{ name: '本机', value: [props.myLocation.lng + offset, props.myLocation.lat - offset] }],
      symbolSize: 10, symbol: 'pin',
      rippleEffect: { brushType: 'stroke', scale: 2, period: 2.5 },
      label: { show: true, formatter: '本机', position: 'top', color: '#13c2c2', fontSize: 11, offset: [0, -6] },
      itemStyle: { color: '#13c2c2', shadowBlur: 8, shadowColor: '#13c2c266' },
      zlevel: 1,
    })
  }

  if (showPcLine) {
    series.push({
      id: 'pc-line', type: 'lines', coordinateSystem: 'geo', polyline: false,
      data: [{ coords: [[props.myLocation!.lng, props.myLocation!.lat], hz], lineStyle: { color: '#13c2c2', width: 1.5, type: 'dashed' } }],
      effect: { show: true, period: 4, trailLength: 0.2, symbol: 'circle', symbolSize: 4, color: '#13c2c2' },
    })
  }

  if (showOverseas) {
    series.push({
      id: 'overseas-lines', type: 'lines', coordinateSystem: 'geo', polyline: false,
      animationDelay: 800,
      data: groups.filter(g => g.reachable && !g.selected).map(g => ({
        name: g.country,
        coords: [hz, g.location], lineStyle: { color: '#6b8cae33', width: 1 },
      })),
      effect: { show: true, period: 6, trailLength: 0.3, symbol: 'arrow', symbolSize: 6, color: '#1677ff' },
    })
    if (currentGroup) {
      series.push({
        id: 'current-proxy-line', type: 'lines', coordinateSystem: 'geo', polyline: false,
        zlevel: 2,
        data: [{
          name: currentGroup.displayName,
          coords: [hz, currentGroup.location],
          lineStyle: { color: activeProxyColor, width: 2.2, opacity: currentGroup.reachable ? 0.9 : 0.55 },
        }],
        effect: { show: true, period: 3.6, trailLength: 0.28, symbol: 'arrow', symbolSize: 8, color: activeProxyColor },
      })
    }
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
          const g = groups.find(g => g.displayName === p.name)
          const ns = g ? ` (${g.count} 个节点)` : ''
          return `<b>${p.name}</b>${ns}<br/>${reachable ? `可达 ${ms < 9999 ? ms + 'ms' : ''}` : '不可达'}`
        }
        return `${p.name}`
      },
    },
    geo: {
      map: 'world', roam: true, zoom: 1.2, center: [0, 25],
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

// 首次连接：延迟飞线动画
watch(() => props.connected, (c, prev) => {
  if (animTimer) { clearTimeout(animTimer); animTimer = null }
  if (c === 'connected' && prev !== 'connected') {
    showOverseasLines.value = false
    render(false)
    animTimer = setTimeout(() => { showOverseasLines.value = true; render(true) }, 1500)
  } else if (c !== 'connected') {
    showOverseasLines.value = false
    render(false)
  }
})

// 代理节点变化防抖渲染
let renderTimer: ReturnType<typeof setTimeout> | null = null
watch(() => props.proxyNodes, () => {
  if (!renderTimer) {
    renderTimer = setTimeout(() => {
      renderTimer = null
      if (showOverseasLines.value) render(true)
    }, 300)
  }
}, { deep: true })

onMounted(async () => {
  if (!container.value) return
  try { const r = await fetch('world.json'); echarts.registerMap('world', await r.json()) } catch { /* */ }
  chart.value = echarts.init(container.value)
  render(true)
  resizeObs = new ResizeObserver(() => chart.value?.resize())
  resizeObs.observe(container.value)
})

onUnmounted(() => { if (animTimer) clearTimeout(animTimer); if (renderTimer) clearTimeout(renderTimer); resizeObs?.disconnect(); chart.value?.dispose() })
</script>

<template>
  <div ref="container" class="chart"></div>
</template>

<style scoped>
.chart { width: 100%; height: 100%; }
</style>
