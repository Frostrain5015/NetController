import { ref, onMounted } from 'vue'

export type ConnStatus = 'connecting' | 'connected' | 'disconnected'

export function useAgentData() {
  const snapshot = ref<Snapshot | null>(null)
  const connected = ref<ConnStatus>('disconnected')

  onMounted(() => {
    window.electronAPI?.onAgentUpdate((data: Snapshot) => { snapshot.value = data })
    window.electronAPI?.onAgentConnection((s: any) => {
      connected.value = (s && s.status) ? s.status : (s ? 'connected' : 'disconnected')
    })
  })

  return { snapshot, connected }
}
