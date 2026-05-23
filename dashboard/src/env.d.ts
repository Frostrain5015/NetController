/// <reference types="vite/client" />

interface ElectronAPI {
  onAgentUpdate: (callback: (data: Snapshot) => void) => void
  onAgentConnection: (callback: (connected: boolean) => void) => void
  getSettings: () => Promise<{ host: string; port: number; username: string; password: string }>
  saveSettings: (data: { host: string; port: number; username: string; password: string }) => Promise<boolean>
  getConfig: () => Promise<any>
  saveConfig: (config: any) => Promise<boolean>
  minimize: () => void
  maximize: () => void
  close: () => void
}

interface Window { electronAPI?: ElectronAPI }

interface ProjectStatus {
  name: string; alive: boolean; pid: number; port: number; portOpen: boolean
  cpuPercent: number; memMB: number; parent?: string
}
interface ProxyStatus {
  name?: string; alive: boolean; port: number; portOpen: boolean; activeConnections: number
}
interface OverseasNode {
  name: string; location: [number, number]; reachable: boolean; latencyMs: number; country: string
}
interface ServerMetrics {
  cpuPercent: number; memPercent: number; diskPercent: number
  memUsedMB: number; memTotalMB: number; diskUsed: string; diskTotal: string
}
interface Snapshot {
  timestamp: number; serverMetrics: ServerMetrics
  projects: ProjectStatus[]; proxy: ProxyStatus; overseasNodes: OverseasNode[]
}
