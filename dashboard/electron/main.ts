import { app, BrowserWindow, ipcMain, Tray, Menu, nativeImage, Notification } from 'electron'
import { join } from 'path'
import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs'
import { createCipheriv, createDecipheriv, randomBytes, scryptSync } from 'crypto'
import { Client } from 'ssh2'

const CONFIG_KEY = 'netcontroller'
const CONFIG_FILE = join(app.getPath('userData'), 'config.json')

const DEFAULTS = {
  ssh: { host: '116.62.179.231', port: 22, username: 'root', password: '' },
  config: {
    socksProxy: '127.0.0.1:7897',
    projects: [
      { name: 'PP Typeset', processName: 'pp-typeset', port: 4443, pm2: true },
      { name: 'PP Typeset Worker', processName: 'pp-typeset-worker', port: 0, pm2: true, parent: 'PP Typeset' },
      { name: 'MongoDB', processName: 'mongod', port: 27017, pm2: false, parent: 'PP Typeset' },
      { name: 'Investory', processName: 'java', port: 8443, pm2: false },
      { name: 'MySQL', processName: 'mysqld', port: 3306, pm2: false, parent: 'Investory' },
      { name: 'Blades of Hex', processName: 'blades-of-hex', port: 3000, pm2: true },
      { name: 'Webhook', processName: 'deploy-webhook', port: 9000, pm2: true },
      { name: 'Nginx', processName: 'nginx', port: 80, pm2: false },
    ],
    proxy: { processName: 'mihomo', port: 7897 },
    overseasNodes: [
      { name: '东京', host: 'www.yahoo.co.jp', port: 443, lat: 35.69, lng: 139.69, country: 'JP' },
      { name: '香港', host: 'www.hk01.com', port: 443, lat: 22.30, lng: 114.17, country: 'HK' },
      { name: '新加坡', host: 'www.channelnewsasia.com', port: 443, lat: 1.35, lng: 103.82, country: 'SG' },
      { name: '旧金山', host: 'www.github.com', port: 443, lat: 37.77, lng: -122.42, country: 'US' },
    ],
  },
}

function deriveKey(): Buffer { return scryptSync(CONFIG_KEY, 'netcontroller-salt', 32) }

let _configCache: any = null

function loadConfig(): any {
  if (_configCache) return _configCache
  try {
    if (!existsSync(CONFIG_FILE)) return (_configCache = JSON.parse(JSON.stringify(DEFAULTS)))
    const raw = readFileSync(CONFIG_FILE)
    const iv = raw.subarray(0, 16), tag = raw.subarray(raw.length - 16), enc = raw.subarray(16, raw.length - 16)
    const dec = createDecipheriv('aes-256-gcm', deriveKey(), iv)
    dec.setAuthTag(tag)
    _configCache = JSON.parse(Buffer.concat([dec.update(enc), dec.final()]).toString('utf-8'))
  } catch { _configCache = JSON.parse(JSON.stringify(DEFAULTS)) }
  return _configCache
}

function saveConfig(data: any) {
  _configCache = data
  const dir = app.getPath('userData')
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  const json = JSON.stringify(data)
  const iv = randomBytes(16)
  const cipher = createCipheriv('aes-256-gcm', deriveKey(), iv)
  const enc = Buffer.concat([cipher.update(json, 'utf-8'), cipher.final()])
  writeFileSync(CONFIG_FILE, Buffer.concat([iv, enc, cipher.getAuthTag()]))
}

let mainWindow: BrowserWindow | null = null
let tray: Tray | null = null
let sshClient: Client | null = null
let collectTimer: ReturnType<typeof setInterval> | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let wasConnected = false
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

function updateTray(c: boolean) {
  if (!tray) return
  tray.setImage(trayIcon(c)); tray.setToolTip(c ? 'NetController - 已连接' : 'NetController - 未连接')
}

function createTray() {
  tray = new Tray(trayIcon(false)); tray.setToolTip('NetController')
  tray.setContextMenu(Menu.buildFromTemplate([
    { label: '显示窗口', click: () => { mainWindow?.show(); mainWindow?.focus() } },
    { type: 'separator' },
    { label: '退出', click: () => { quitting = true; app.quit() } },
  ]))
  tray.on('double-click', () => { mainWindow?.show(); mainWindow?.focus() })
}

function notify(title: string, body: string) {
  if (Notification.isSupported()) new Notification({ title, body }).show()
}

function sshExec(cmd: string, timeout = 10): Promise<string> {
  return new Promise((resolve) => {
    if (!sshClient) { resolve(''); return }
    sshClient.exec(cmd, (err, stream) => {
      if (err) { resolve(''); return }
      let out = ''
      stream.on('data', (d: Buffer) => { out += d.toString() })
      stream.stderr.on('data', () => {})
      stream.on('close', () => resolve(out.trim()))
      setTimeout(() => resolve(out.trim()), timeout * 1000)
    })
  })
}

async function collectPM2(): Promise<Map<string, { alive: boolean; pid: number; cpu: number; memMB: number }>> {
  const map = new Map()
  const raw = await sshExec('pm2 jlist 2>/dev/null')
  if (!raw) return map
  try {
    for (const p of JSON.parse(raw)) {
      map.set(p.name, {
        alive: p.pm2_env?.status === 'online',
        pid: p.pid || 0,
        cpu: p.monit?.cpu || 0,
        memMB: Math.round((p.monit?.memory || 0) / 1024 / 1024),
      })
    }
  } catch { /* pm2 not available */ }
  return map
}

async function collectProcess(name: string): Promise<{ alive: boolean; pid: number; cpu: number; memMB: number }> {
  const out = await sshExec(`ps aux 2>/dev/null | grep -v grep | grep '${name}' | awk '{print $2,$3,$4}'`)
  if (!out) return { alive: false, pid: 0, cpu: 0, memMB: 0 }
  const parts = out.split(/\s+/)
  const pid = parseInt(parts[0]) || 0
  const cpu = parseFloat(parts[1]) || 0
  let memMB = 0
  if (pid > 0) {
    const rss = await sshExec(`ps -p ${pid} -o rss= 2>/dev/null`)
    memMB = Math.floor((parseInt(rss) || 0) / 1024)
  }
  return { alive: !!pid, pid, cpu, memMB }
}

async function collectPort(port: number) {
  if (!port) return false
  return !!(await sshExec(`ss -tlnp 2>/dev/null | grep ':${port} ' || netstat -tlnp 2>/dev/null | grep ':${port} '`))
}

async function collectSystem() {
  const [cpuOut, memLine, diskLine] = await Promise.all([
    sshExec("top -bn1 2>/dev/null | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1"),
    sshExec("free -m 2>/dev/null | grep Mem | awk '{print $2,$3,$4}'"),
    sshExec("df -h / 2>/dev/null | tail -1 | awk '{print $2,$3,$4}'"),
  ])
  const cpuPercent = parseFloat(cpuOut) || 0
  const memParts = memLine.split(/\s+/)
  const memTotal = parseInt(memParts[0]) || 0
  const memUsed = parseInt(memParts[1]) || 0
  const diskParts = diskLine.split(/\s+/)
  const diskTotal = diskParts[0] || '?'
  const diskUsed = diskParts[1] || '?'
  return {
    cpuPercent,
    memPercent: memTotal ? Math.round(memUsed / memTotal * 1000) / 10 : 0,
    memUsedMB: memUsed,
    memTotalMB: memTotal,
    diskPercent: memTotal ? 0 : 0, // computed below
    diskUsed,
    diskTotal,
  }
}

async function probeViaProxy(socksProxy: string, host: string, port: number, timeout = 5): Promise<[boolean, number]> {
  const scheme = port === 443 ? 'https' : 'http'
  // curl 超时返回 000，此时视为不可达；成功则返回 total time
  const out = await sshExec(`curl --socks5-hostname ${socksProxy} --connect-timeout ${timeout} --max-time ${timeout} -s -o /dev/null -w '%{http_code} %{time_total}' -k -L ${scheme}://${host}:${port} 2>/dev/null`)
  if (!out) return [false, -1]
  const parts = out.split(/\s+/)
  const httpCode = parseInt(parts[0]) || 0
  const totalMs = Math.round(parseFloat(parts[1] || '0') * 1000)
  // HTTP 000 = curl 连接失败/超时
  if (httpCode === 0) return [false, -1]
  return [totalMs > 0 && totalMs < timeout * 1000, totalMs]
}

async function collectAndBroadcast() {
  const cfg = loadConfig().config
  const snap: any = { timestamp: Math.floor(Date.now() / 1000), serverMetrics: {}, projects: [], proxy: {}, overseasNodes: [] }
  try {
    const [sys, pm2Map] = await Promise.all([collectSystem(), collectPM2()])
    snap.serverMetrics = sys

    for (const proj of cfg.projects) {
      let proc: { alive: boolean; pid: number; cpu: number; memMB: number }
      if (proj.pm2) {
        proc = pm2Map.get(proj.processName) || { alive: false, pid: 0, cpu: 0, memMB: 0 }
      } else {
        proc = await collectProcess(proj.processName)
      }
      snap.projects.push({
        name: proj.name, alive: proc.alive, pid: proc.pid,
        port: proj.port || 0,
        portOpen: proj.port ? await collectPort(proj.port) : false,
        cpuPercent: Math.round(proc.cpu * 10) / 10, memMB: proc.memMB,
        parent: proj.parent || '',
      })
    }

    const px = await collectProcess(cfg.proxy.processName)
    snap.proxy = {
      name: 'Clash',
      alive: px.alive, port: cfg.proxy.port,
      portOpen: cfg.proxy.port ? await collectPort(cfg.proxy.port) : false,
      activeConnections: 0,
    }

    const socksProxy = cfg.socksProxy || '127.0.0.1:7897'
    snap.overseasNodes = await Promise.all(
      cfg.overseasNodes.map(async (node: any) => {
        const [reachable, ms] = await probeViaProxy(socksProxy, node.host, node.port)
        return { name: node.name, location: [node.lng, node.lat], reachable, latencyMs: ms, country: node.country || '' }
      })
    )

    mainWindow?.webContents.send('agent-update', snap)
    if (!wasConnected) {
      updateTray(true); mainWindow?.webContents.send('agent-connection', { status: 'connected' })
      wasConnected = true
    }
  } catch (e: any) { console.error('[collect]', e.message) }
}

function sshConnect() {
  const { host, port, username, password } = loadConfig().ssh
  if (!host || !password) {
    mainWindow?.webContents.send('agent-connection', { status: 'disconnected' }); updateTray(false)
    reconnectTimer = setTimeout(sshConnect, 10000); return
  }
  if (sshClient) { try { sshClient.end() } catch { /* ok */ } }
  mainWindow?.webContents.send('agent-connection', { status: 'connecting' })
  sshClient = new Client()
  sshClient.on('ready', () => {
    wasConnected = false
    if (collectTimer) clearInterval(collectTimer)
    collectAndBroadcast() // 立即采集
    collectTimer = setInterval(collectAndBroadcast, 15000)
  })
  sshClient.on('error', (err) => { console.error('[ssh] error:', err.message) })
  sshClient.on('close', () => {
    if (collectTimer) { clearInterval(collectTimer); collectTimer = null }
    updateTray(false); mainWindow?.webContents.send('agent-connection', { status: 'disconnected' })
    if (wasConnected) notify('NetController', '与服务器断开连接，正在重连...')
    wasConnected = false; sshClient = null
    reconnectTimer = setTimeout(sshConnect, 5000)
  })
  sshClient.connect({ host, port, username, password, readyTimeout: 10000 })
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

ipcMain.handle('get-settings', () => { const s = loadConfig().ssh; return { host: s.host, port: s.port, username: s.username, password: s.password } })
ipcMain.handle('save-settings', (_e, data) => { const c = loadConfig(); c.ssh = data; saveConfig(c); sshConnect(); return true })
ipcMain.handle('get-config', () => loadConfig().config)
ipcMain.handle('save-config', (_e, config) => { const c = loadConfig(); c.config = config; saveConfig(c); return true })
ipcMain.on('win-minimize', () => mainWindow?.minimize())
ipcMain.on('win-maximize', () => mainWindow?.isMaximized() ? mainWindow?.unmaximize() : mainWindow?.maximize())
ipcMain.on('win-close', () => mainWindow?.close())

app.whenReady().then(() => { createTray(); createWindow(); sshConnect() })
app.on('window-all-closed', () => {})
app.on('activate', () => { BrowserWindow.getAllWindows().length === 0 ? createWindow() : mainWindow?.show() })
app.on('before-quit', () => {
  quitting = true
  if (reconnectTimer) clearTimeout(reconnectTimer)
  if (collectTimer) clearInterval(collectTimer)
  if (sshClient) { sshClient.end(); sshClient = null }
})
