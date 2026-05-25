import { app, BrowserWindow, ipcMain, Tray, Menu, nativeImage } from 'electron'
import { join } from 'path'
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs'

const CONFIG_FILE = join(app.getPath('userData'), 'config.json')

const DEFAULTS = {
  wsUrl: 'ws://116.62.179.231:9527/ws',
}

function loadConfig(): any {
  try {
    if (!existsSync(CONFIG_FILE)) return { ...DEFAULTS }
    return JSON.parse(readFileSync(CONFIG_FILE, 'utf-8'))
  } catch { return { ...DEFAULTS } }
}

function saveConfig(data: any) {
  const dir = app.getPath('userData')
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  writeFileSync(CONFIG_FILE, JSON.stringify(data))
}

let mainWindow: BrowserWindow | null = null
let tray: Tray | null = null
let quitting = false

function trayIcon(connected: boolean): nativeImage {
  const size = 16; const canvas = Buffer.alloc(size * size * 4)
  const color = connected ? [82, 196, 26, 255] : [255, 77, 79, 255]
  for (let i = 0; i < size * size; i++) {
    const cx = i % size, cy = Math.floor(i / size)
    if ((cx - size / 2) ** 2 + (cy - size / 2) ** 2 <= (size / 2 - 1) ** 2) {
      canvas[i * 4] = color[0]; canvas[i * 4 + 1] = color[1]
      canvas[i * 4 + 2] = color[2]; canvas[i * 4 + 3] = color[3]
    }
  }
  return nativeImage.createFromBuffer(canvas, { width: size, height: size })
}

function createTray() {
  tray = new Tray(trayIcon(false)); tray.setToolTip('NetController')
  tray.setContextMenu(Menu.buildFromTemplate([
    { label: '显示窗口', click: () => { mainWindow?.show(); mainWindow?.focus() } },
    { type: 'separator' },
    { label: '退出', click: () => {
      quitting = true
      if (tray) { try { tray.destroy() } catch { /* */ }; tray = null }
      if (mainWindow) {
        try { mainWindow.removeAllListeners('close') } catch { /* */ }
        try { mainWindow.close() } catch { /* */ }
      }
    } },
  ]))
  tray.on('double-click', () => { mainWindow?.show(); mainWindow?.focus() })
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400, height: 900, minWidth: 1100, minHeight: 700,
    title: 'NetController', backgroundColor: '#0f1923', frame: false,
    webPreferences: { preload: join(__dirname, 'preload.js'), contextIsolation: true, nodeIntegration: false },
  })
  if (process.env.VITE_DEV_SERVER_URL) mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL)
  else mainWindow.loadFile(join(__dirname, '../dist/index.html'))
  mainWindow.on('close', (e) => { if (!quitting) { e.preventDefault(); mainWindow?.hide() } })
}

// IPC
ipcMain.handle('get-config', () => loadConfig())
ipcMain.handle('save-config', (_e, cfg) => { saveConfig(cfg); return true })
ipcMain.on('win-minimize', () => mainWindow?.minimize())
ipcMain.on('win-maximize', () => mainWindow?.isMaximized() ? mainWindow?.unmaximize() : mainWindow?.maximize())
ipcMain.on('win-close', () => mainWindow?.close())

app.whenReady().then(() => { createTray(); createWindow() })
app.on('window-all-closed', () => { if (quitting) app.quit() })
app.on('activate', () => { BrowserWindow.getAllWindows().length === 0 ? createWindow() : mainWindow?.show() })
app.on('before-quit', () => {
  quitting = true
  try { mainWindow?.removeAllListeners('close') } catch { /* */ }
})
