import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electronAPI', {
  onAgentUpdate: (callback: (data: any) => void) => {
    ipcRenderer.on('agent-update', (_event, data) => callback(data))
  },
  onAgentConnection: (callback: (connected: boolean) => void) => {
    ipcRenderer.on('agent-connection', (_event, connected) => callback(connected))
  },
  getSettings: () => ipcRenderer.invoke('get-settings'),
  saveSettings: (data: { host: string; port: number; username: string; password: string }) =>
    ipcRenderer.invoke('save-settings', data),
  getConfig: () => ipcRenderer.invoke('get-config'),
  saveConfig: (config: any) => ipcRenderer.invoke('save-config', config),
  // 窗口控制
  minimize: () => ipcRenderer.send('win-minimize'),
  maximize: () => ipcRenderer.send('win-maximize'),
  close: () => ipcRenderer.send('win-close'),
})
