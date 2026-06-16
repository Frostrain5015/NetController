/// <reference types="vite/client" />

interface UpdateInfo { version: string }
interface UpdateProgress {
  percent: number; transferred: number; total: number; bytesPerSecond: number
}
interface UpdateAPI {
  onAvailable: (cb: (info: UpdateInfo) => void) => void
  onProgress: (cb: (p: UpdateProgress) => void) => void
  onDownloaded: (cb: (info: UpdateInfo) => void) => void
  onNone: (cb: () => void) => void
  onError: (cb: (msg: string) => void) => void
  check: () => Promise<boolean>
  download: () => void
}

interface ElectronAPI {
  getConfig: () => Promise<{ wsUrl: string }>
  saveConfig: (config: { wsUrl: string }) => Promise<boolean>
  minimize: () => void
  maximize: () => void
  close: () => void
  reportConnection: (state: string) => void
  update: UpdateAPI
}

interface Window { electronAPI?: ElectronAPI }

interface ProjectStatus {
  name: string; alive: boolean; pid: number; port: number; portOpen: boolean
  cpuPercent: number; memMB: number; parent?: string
}
interface ProxyStatus {
  name?: string; alive: boolean; port: number; portOpen: boolean; activeConnections: number
  apiAccessible: boolean
  trafficRemainingGB: number | null
  trafficUsedGB: number | null
  trafficTotalGB: number | null
  planExpiry: string | null
}
interface ProxyNode {
  name: string; displayName: string; group: string; groupType: string; type: string
  latencyMs: number; reachable: boolean; selected: boolean
  location: [number, number] | null; country: string
}
interface ServerMetrics {
  cpuPercent: number; memPercent: number; diskPercent: number
  memUsedMB: number; memTotalMB: number; diskUsed: string; diskTotal: string
}
interface Snapshot {
  timestamp: number; serverMetrics: ServerMetrics
  projects: ProjectStatus[]; proxy: ProxyStatus; proxyNodes: ProxyNode[]
}
