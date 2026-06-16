import { contextBridge, ipcRenderer } from 'electron'
import type { IpcRendererEvent } from 'electron'

interface WsConfig { wsUrl: string }
interface UpdateInfo { version: string }
interface UpdateProgress {
  percent: number; transferred: number; total: number; bytesPerSecond: number
}

contextBridge.exposeInMainWorld('electronAPI', {
  getConfig: () => ipcRenderer.invoke('get-config'),
  saveConfig: (config: WsConfig) => ipcRenderer.invoke('save-config', config),
  minimize: () => ipcRenderer.send('win-minimize'),
  maximize: () => ipcRenderer.send('win-maximize'),
  close: () => ipcRenderer.send('win-close'),
  reportConnection: (state: string) => ipcRenderer.send('conn-state', state),
  update: {
    onAvailable: (cb: (info: UpdateInfo) => void) =>
      ipcRenderer.on('update:available', (_e: IpcRendererEvent, info: UpdateInfo) => cb(info)),
    onProgress: (cb: (p: UpdateProgress) => void) =>
      ipcRenderer.on('update:progress', (_e: IpcRendererEvent, p: UpdateProgress) => cb(p)),
    onDownloaded: (cb: (info: UpdateInfo) => void) =>
      ipcRenderer.on('update:downloaded', (_e: IpcRendererEvent, info: UpdateInfo) => cb(info)),
    onNone: (cb: () => void) =>
      ipcRenderer.on('update:none', () => cb()),
    onError: (cb: (msg: string) => void) =>
      ipcRenderer.on('update:error', (_e: IpcRendererEvent, msg: string) => cb(msg)),
    check: () => ipcRenderer.invoke('update:check'),
    download: () => ipcRenderer.send('update:download'),
  },
})
