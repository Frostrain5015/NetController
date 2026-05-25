# NetController

阿里云服务器监控仪表盘 — 通过 SSH 远程采集服务器系统指标、进程状态、端口监听和海外节点可达性，在 Electron 桌面端以地图和仪表盘形式实时展示。

## 架构

```
┌──────────────────────────────────────────────────────────┐
│  Electron Dashboard (Vue 3 + Ant Design + ECharts)       │
│  - 世界地图 (海外节点延迟)                                 │
│  - 中国地图 (国内节点延迟，预留)                             │
│  - 系统指标仪表盘 (CPU / 内存 / 磁盘)                       │
│  - 项目进程状态面板                                        │
│  - 系统托盘 + 断连通知                                      │
│  ┌──────────────┐  ┌─────────────────────────────────┐   │
│  │ SSH Bridge    │  │ Go Agent (部署在目标服务器)       │   │
│  │ (paramiko)   │  │ - WebSocket 服务端 (:9527)       │   │
│  │ Python 3     │  │ - 每 5s 采集系统/进程/端口状态     │   │
│  └──────────────┘  └─────────────────────────────────┘   │
└──────────────────────────────────────────────────────────┘
```

**两种采集模式：**

| 模式 | 说明 |
|------|------|
| **Go Agent** | 编译后部署到阿里云服务器，直接采集系统指标，Dashboard 通过 WebSocket 连接 |
| **SSH Bridge** | Python 脚本通过 SSH 远程执行命令采集（fallback 方案，无需在服务器部署 agent） |
| **内置 SSH** | Electron 主进程直接使用 `ssh2` 库连接服务器采集（启动即用，默认模式） |

## 项目结构

```
NetController/
├── agent/                    # Go Agent — 部署在目标服务器
│   ├── main.go               # 入口：WebSocket 服务 + 采集循环
│   ├── agent.yaml             # 配置文件（进程/端口/海外节点）
│   ├── collector/             # 系统指标/进程/端口/TCP 探测采集器
│   │   ├── system.go
│   │   ├── process.go
│   │   └── network.go
│   ├── config/config.go       # YAML 配置加载
│   └── ws/hub.go              # WebSocket Hub（广播快照）
│
├── bridge/                    # Python SSH Bridge — fallback 方案
│   ├── bridge.py              # paramiko SSH 采集 + WebSocket 推送
│   ├── requirements.txt
│   └── run_bridge.bat
│
└── dashboard/                 # Electron 桌面仪表盘
    ├── electron/
    │   ├── main.ts            # 主进程：SSH2 直连、采集调度、托盘、加密存储
    │   └── preload.ts         # 预加载脚本：contextBridge API
    ├── src/
    │   ├── App.vue            # 主布局（自定义标题栏 + 内容区）
    │   ├── main.ts            # Vue 入口
    │   ├── components/
    │   │   ├── WorldMap.vue    # 世界地图（海外节点延迟热力图）
    │   │   ├── ChinaMap.vue    # 中国地图
    │   │   ├── ServerMetrics.vue     # 系统指标环形图
    │   │   ├── StatusSidebar.vue     # 进程状态侧边栏
    │   │   ├── SettingsModal.vue     # SSH 连接设置弹窗
    │   │   └── FlagIcon.vue          # 国旗图标
    │   └── composables/
    │       ├── useWebSocket.ts       # WS 连接管理
    │       └── useGeoLocation.ts     # 地理位置工具
    ├── electron-builder.yml   # 打包配置
    └── package.json
```

## 功能

- **系统监控**：CPU 使用率、内存使用量/百分比、磁盘使用量
- **进程监控**：按进程名匹配，支持 PM2 管理进程和系统进程，显示 PID / CPU / 内存
- **端口检测**：检查进程端口是否在监听
- **代理状态**：显示代理进程存活、端口监听、活跃连接数
- **海外节点探测**：通过代理测试海外节点 TCP 可达性和延迟，在世界地图上可视化
- **系统托盘**：连接状态指示（绿/红点），断线桌面通知
- **配置加密**：SSH 密码等敏感信息使用 AES-256-GCM 加密存储

## 构建 & 运行

### 前置条件

- [Node.js](https://nodejs.org/) >= 18
- [Go](https://go.dev/) >= 1.26（仅编译 Agent 时需要）
- [Python](https://www.python.org/) >= 3.9（仅使用 SSH Bridge 时需要）

### Dashboard（推荐使用）

```bash
cd dashboard

# 安装依赖
npm install

# 开发模式（需要 VITE_DEV_SERVER_URL）
npm run dev

# 构建 Windows 便携版 exe
npm run build
```

构建产物在 `dashboard/release/NetController.exe`。

### Go Agent

```bash
cd agent

# Windows 交叉编译 (部署到 Linux 服务器)
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o netcontroller-agent .
# 或在 Linux 上直接编译为可执行文件
go build -o netcontroller-agent .

# 上传到服务器后运行
./netcontroller-agent -c agent.yaml
```

### SSH Bridge

```bash
cd bridge
pip install -r requirements.txt
python bridge.py --host <服务器IP> --user root --password <密码> --config config.yaml
```

或使用环境变量避免密码出现在命令行历史中：

```bash
$env:NC_SSH_HOST = "116.62.179.231"
$env:NC_SSH_USER = "root"
$env:NC_SSH_PASSWORD = "your-password"
python bridge.py --host $env:NC_SSH_HOST --user $env:NC_SSH_USER --password $env:NC_SSH_PASSWORD
```

### 配置说明

配置文件为 YAML 格式（Go Agent）或在 Dashboard 界面中编辑（内置 SSH 模式）。主要配置项：

```yaml
projects:
  - name: "项目名称"
    processName: "进程名（ps aux 匹配）"
    port: 端口号（0 表示不检测）
    pm2: true          # 是否 PM2 管理（Dashboard 模式）
    parent: "父项目"    # 分组（Dashboard 模式）

proxy:
  processName: "代理进程名"
  port: 代理端口

overseasNodes:
  - name: "东京"
    host: "www.yahoo.co.jp"
    port: 443
    lat: 35.69
    lng: 139.69
```

## 技术栈

| 层 | 技术 |
|----|------|
| 桌面框架 | Electron 30 + Vue 3 + TypeScript |
| UI 组件 | Ant Design Vue 4 |
| 图表 | ECharts 5 + 世界/中国地图 GeoJSON |
| 网络 | ssh2 (Node.js) / paramiko (Python) / gorilla/websocket (Go) |
| Go 采集 | gopsutil v4 (系统指标) |
| 构建 | Vite 5 + electron-builder |
