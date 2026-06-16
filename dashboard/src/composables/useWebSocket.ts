import { ref, onMounted, onUnmounted } from 'vue'

export type ConnStatus = 'connecting' | 'connected' | 'disconnected'

export function useAgentData() {
  const snapshot = ref<Snapshot | null>(null)
  const connected = ref<ConnStatus>('disconnected')

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let wsUrl = 'wss://frostrain.tech/nc/ws'

  function connect() {
    if (ws) { try { ws.close() } catch { /* */ } }
    connected.value = 'connecting'
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      connected.value = 'connected'
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'proxy-ping-result') {
          if (!snapshot.value) return
          const node = snapshot.value.proxyNodes.find((n: ProxyNode) => n.name === data.name)
          if (node) {
            node.latencyMs = data.latencyMs
            node.reachable = data.reachable
            ;(node as any).tested = true
          }
        } else {
          const prevNodes = snapshot.value?.proxyNodes ?? []
          snapshot.value = data as Snapshot
          if (data.proxyNodes && data.proxyNodes.length > 0) {
            // 快照来自 ping 完成后——所有节点均已探测，标记 tested
            for (const n of snapshot.value.proxyNodes) {
              ;(n as any).tested = true
            }
          } else if (prevNodes.length > 0) {
            // 系统快照不含 proxyNodes ——保留已有节点状态
            for (const n of prevNodes) {
              ;(n as any).tested = true // ping 结果可能比快照先到
            }
            snapshot.value.proxyNodes = prevNodes
          }
        }
      } catch { /* */ }
    }

    ws.onclose = () => {
      connected.value = 'disconnected'
      ws = null
      reconnectTimer = setTimeout(connect, 5000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  onMounted(async () => {
    try {
      const cfg = await window.electronAPI?.getConfig()
      if (cfg?.wsUrl) wsUrl = cfg.wsUrl
    } catch { /* */ }
    connect()
  })

  onUnmounted(() => {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    if (ws) { try { ws.close() } catch { /* */ } }
  })

  return { snapshot, connected }
}
