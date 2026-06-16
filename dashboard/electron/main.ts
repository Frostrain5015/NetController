import { app, BrowserWindow, ipcMain, Tray, Menu, nativeImage } from 'electron'
import type { NativeImage } from 'electron'
import { join } from 'path'
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs'
import pkg from 'electron-updater'
const { autoUpdater } = pkg

const APP_ID = 'com.netcontroller.dashboard'
const CONFIG_FILE = join(app.getPath('userData'), 'config.json')

const DEFAULTS = {
  wsUrl: 'wss://frostrain.tech/nc/ws',
}

function loadConfig(): Record<string, unknown> {
  try {
    if (!existsSync(CONFIG_FILE)) return { ...DEFAULTS }
    return JSON.parse(readFileSync(CONFIG_FILE, 'utf-8'))
  } catch { return { ...DEFAULTS } }
}

function saveConfig(data: Record<string, unknown>) {
  const dir = app.getPath('userData')
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  writeFileSync(CONFIG_FILE, JSON.stringify(data))
}

let mainWindow: BrowserWindow | null = null
let tray: Tray | null = null
let quitting = false

type ConnState = 'connected' | 'connecting' | 'disconnected'

// ── Tray icon: glowing diamond tinted by connection state ──────────────
const STATE_RGB: Record<ConnState, [number, number, number]> = {
  connected: [61, 220, 151],
  connecting: [245, 185, 66],
  disconnected: [255, 93, 108],
}

function resolveIcon(): string | undefined {
  const candidates = app.isPackaged
    ? [join(process.resourcesPath, 'icon.ico')]
    : [join(process.cwd(), 'build', 'icon.ico'), join(app.getAppPath(), 'build', 'icon.ico')]
  return candidates.find((path) => existsSync(path))
}

function fallbackTrayIcon(state: ConnState): NativeImage {
  const S = 32, SS = 3, N = S * SS
  const buf = Buffer.alloc(S * S * 4)
  const [cr, cg, cb] = STATE_RGB[state]
  for (let y = 0; y < S; y++) {
    for (let x = 0; x < S; x++) {
      let r = 0, g = 0, b = 0, a = 0
      for (let sy = 0; sy < SS; sy++) {
        for (let sx = 0; sx < SS; sx++) {
          const u = (x * SS + sx + 0.5) / N - 0.5
          const v = (y * SS + sy + 0.5) / N - 0.5
          const dia = (Math.abs(u) + Math.abs(v)) - 0.34 // diamond, <0 inside
          const fill = dia < 0 ? 1 : 0
          const glow = Math.exp(-Math.max(dia, 0) * 9) * 0.8
          const cov = Math.min(1, fill + glow)
          r += cr * cov; g += cg * cov; b += cb * cov; a += cov * 255
        }
      }
      const n = SS * SS, i = (S * y + x) << 2
      buf[i] = Math.round(r / n); buf[i + 1] = Math.round(g / n)
      buf[i + 2] = Math.round(b / n); buf[i + 3] = Math.round(a / n)
    }
  }
  return nativeImage.createFromBuffer(buf, { width: S, height: S })
}

function trayIcon(state: ConnState): NativeImage {
  const icon = resolveIcon()
  if (icon) {
    const image = nativeImage.createFromPath(icon)
    if (!image.isEmpty()) return image
  }
  return fallbackTrayIcon(state)
}

function showMainWindow() {
  if (!mainWindow || mainWindow.isDestroyed()) createWindow()
  if (!mainWindow) return
  if (mainWindow.isMinimized()) mainWindow.restore()
  mainWindow.show()
  mainWindow.focus()
}

function createTray() {
  tray = new Tray(trayIcon('disconnected'))
  tray.setToolTip('NetController — 离线')
  tray.setContextMenu(Menu.buildFromTemplate([
    { label: '显示窗口', click: showMainWindow },
    { type: 'separator' },
    { label: '退出', click: () => { quitNow() } },
  ]))
  tray.on('click', showMainWindow)
  tray.on('double-click', showMainWindow)
}

function updateTray(state: ConnState) {
  if (!tray) return
  try {
    tray.setImage(trayIcon(state))
    const label = state === 'connected' ? '已连接' : state === 'connecting' ? '连接中…' : '离线'
    tray.setToolTip(`NetController — ${label}`)
  } catch { /* */ }
}

function quitNow() {
  quitting = true
  if (tray) { try { tray.destroy() } catch { /* */ }; tray = null }
  if (mainWindow) {
    try { mainWindow.removeAllListeners('close') } catch { /* */ }
    try { mainWindow.close() } catch { /* */ }
  }
  app.quit()
}

function createWindow() {
  const icon = resolveIcon()
  mainWindow = new BrowserWindow({
    width: 1400, height: 900, minWidth: 1100, minHeight: 700,
    title: 'NetController', backgroundColor: '#06090e', frame: false,
    ...(icon ? { icon } : {}),
    webPreferences: { preload: join(__dirname, 'preload.js'), contextIsolation: true, nodeIntegration: false },
  })
  if (process.env.VITE_DEV_SERVER_URL) mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL)
  else mainWindow.loadFile(join(__dirname, '../dist/index.html'))
  mainWindow.on('close', (e) => { if (!quitting) { e.preventDefault(); mainWindow?.hide() } })
}

// ── Auto-update (electron-updater, GitHub provider) ────────────────────
function sendToRenderer(channel: string, payload?: unknown) {
  if (mainWindow && !mainWindow.isDestroyed()) mainWindow.webContents.send(channel, payload)
}

function setupAutoUpdate() {
  autoUpdater.autoDownload = false
  autoUpdater.autoInstallOnAppQuit = false

  autoUpdater.on('update-available', (info) => {
    sendToRenderer('update:available', { version: info.version })
  })
  autoUpdater.on('update-not-available', () => sendToRenderer('update:none'))
  autoUpdater.on('download-progress', (p) => {
    sendToRenderer('update:progress', {
      percent: p.percent, transferred: p.transferred, total: p.total, bytesPerSecond: p.bytesPerSecond,
    })
  })
  autoUpdater.on('update-downloaded', (info) => {
    sendToRenderer('update:downloaded', { version: info.version })
    // Give the UI a moment to show "restarting", then relaunch into the installer.
    setTimeout(() => {
      quitting = true
      if (tray) { try { tray.destroy() } catch { /* */ }; tray = null }
      try { mainWindow?.removeAllListeners('close') } catch { /* */ }
      autoUpdater.quitAndInstall(false, true)
    }, 1400)
  })
  autoUpdater.on('error', (err) => {
    sendToRenderer('update:error', String(err?.message || err))
  })

  // Renderer asks to start the download.
  ipcMain.on('update:download', () => {
    autoUpdater.downloadUpdate().catch((e) => sendToRenderer('update:error', String(e)))
  })
  ipcMain.handle('update:check', async () => {
    if (!app.isPackaged) {
      sendToRenderer('update:none')
      return false
    }
    try {
      await autoUpdater.checkForUpdates()
      return true
    } catch (e) {
      sendToRenderer('update:error', String(e))
      return false
    }
  })

  if (app.isPackaged) {
    autoUpdater.checkForUpdates().catch(() => { /* offline / no release yet */ })
    setInterval(() => { autoUpdater.checkForUpdates().catch(() => { /* */ }) }, 6 * 60 * 60 * 1000)
  }
}

// IPC ───────────────────────────────────────────────────────────────────
ipcMain.handle('get-config', () => loadConfig())
ipcMain.handle('save-config', (_e, cfg: Record<string, unknown>) => { saveConfig(cfg); return true })
ipcMain.on('win-minimize', () => mainWindow?.minimize())
ipcMain.on('win-maximize', () => mainWindow?.isMaximized() ? mainWindow?.unmaximize() : mainWindow?.maximize())
ipcMain.on('win-close', () => mainWindow?.close())
ipcMain.on('conn-state', (_e, state: ConnState) => updateTray(state))

// ── Single-instance lock: only one Dashboard per device ────────────────
const gotLock = app.requestSingleInstanceLock()
if (!gotLock) {
  app.quit()
} else {
  app.on('second-instance', () => {
    showMainWindow()
  })

  app.whenReady().then(() => {
    app.setAppUserModelId(APP_ID)
    createTray()
    createWindow()
    setupAutoUpdate()
  })

  app.on('window-all-closed', () => { if (quitting) app.quit() })
  app.on('activate', showMainWindow)
  app.on('before-quit', () => {
    quitting = true
    try { mainWindow?.removeAllListeners('close') } catch { /* */ }
  })
}
