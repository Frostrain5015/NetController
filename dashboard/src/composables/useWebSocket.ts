import { ref, onMounted, onUnmounted } from 'vue'

export type ConnStatus = 'connecting' | 'connected' | 'disconnected'

export interface SelectResult { ok: boolean; message?: string }

export function useAgentData() {
  const snapshot = ref<Snapshot | null>(null)
  const connected = ref<ConnStatus>('disconnected')

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let wsUrl = 'wss://frostrain.tech/nc/ws'

  // 等待服务器 proxy-select-result 的 Promise 回调
  let selectResolver: ((result: SelectResult) => void) | null = null

  function selectProxyNode(node: ProxyNode): Promise<SelectResult> {
    return new Promise<SelectResult>((resolve) => {
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        resolve({ ok: false, message: '未连接' })
        return
      }
      // 超时 10s
      const timer = setTimeout(() => {
        selectResolver = null
        resolve({ ok: false, message: '超时' })
      }, 10000)
      selectResolver = (result) => {
        clearTimeout(timer)
        selectResolver = null
        resolve(result)
      }
      ws.send(JSON.stringify({
        type: 'proxy-select',
        name: node.name,
        group: node.group,
      }))
    })
  }

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
        } else if (data.type === 'proxy-select-result') {
          // 先通知等待中的 selectProxyNode Promise
          if (selectResolver) {
            selectResolver({ ok: !!data.ok, message: data.message })
          }
          if (!snapshot.value || !data.ok) return
          for (const node of snapshot.value.proxyNodes) {
            node.selected = node.name === data.name
          }
        } else {
          const prevNodes = snapshot.value?.proxyNodes ?? []
          snapshot.value = data as Snapshot
          if (data.proxyNodes && data.proxyNodes.length > 0) {
            for (const n of snapshot.value.proxyNodes) {
              ;(n as any).tested = true
            }
          } else if (prevNodes.length > 0) {
            for (const n of prevNodes) {
              ;(n as any).tested = true
            }
            snapshot.value.proxyNodes = prevNodes
          }
        }
      } catch { /* */ }
    }

    ws.onclose = () => {
      connected.value = 'disconnected'
      ws = null
      // 拒绝正在进行的选择请求
      if (selectResolver) { selectResolver({ ok: false, message: '连接断开' }); selectResolver = null }
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

  return { snapshot, connected, selectProxyNode }
}
