#!/usr/bin/env python3
"""
NetController SSH Bridge
通过 paramiko SSH 连接到阿里云服务器，远程采集进程/端口/系统状态，
本地探测海外节点可达性，通过 WebSocket 推送给 Electron 仪表盘。

用法:
  python bridge.py --host 116.62.179.231 --user root --password xxx [--port 22] [--config config.yaml]
"""

import argparse
import asyncio
import json
import os
import socket
import sys
import threading
import time
from datetime import datetime

import paramiko
import websockets
import yaml

# ---------------------------------------------------------------------------
# 数据结构
# ---------------------------------------------------------------------------

SNAPSHOT_LOCK = threading.Lock()
SNAPSHOT: dict = {"timestamp": 0, "serverMetrics": {}, "projects": [], "proxy": {}, "overseasNodes": []}

# ---------------------------------------------------------------------------
# SSH 采集
# ---------------------------------------------------------------------------

def ssh_connect(host: str, port: int, user: str, password: str) -> paramiko.SSHClient:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=host, port=port, username=user, password=password, timeout=10)
    return client


def ssh_exec(ssh: paramiko.SSHClient, cmd: str, timeout: int = 5) -> str:
    """执行远程命令，返回 stdout 字符串。"""
    try:
        stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
        return stdout.read().decode("utf-8", errors="replace").strip()
    except Exception:
        return ""


def check_process(ssh: paramiko.SSHClient, process_name: str) -> tuple[bool, int, float, int]:
    """检查进程存活，返回 (alive, pid, cpu%, mem_mb)。"""
    out = ssh_exec(ssh,
        f"ps aux 2>/dev/null | grep -v grep | grep '{process_name}' | awk '{{print $2,$3,$4}}'")
    if not out:
        return False, 0, 0.0, 0

    parts = out.split()
    try:
        pid = int(parts[0])
        cpu = float(parts[1])
        mem_pct = float(parts[2])
    except (IndexError, ValueError):
        return False, 0, 0.0, 0

    # 获取内存绝对值 (MB)
    mem_kb = 0
    try:
        mem_out = ssh_exec(ssh, f"ps -p {pid} -o rss= 2>/dev/null")
        if mem_out:
            mem_kb = int(mem_out.strip())
    except Exception:
        pass

    return True, pid, cpu, mem_kb // 1024


def check_port(ssh: paramiko.SSHClient, port: int) -> bool:
    """检查端口是否在监听。"""
    out = ssh_exec(ssh, f"ss -tlnp 2>/dev/null | grep ':{port} ' || netstat -tlnp 2>/dev/null | grep ':{port} '")
    return bool(out)


def check_proxy_connections(ssh: paramiko.SSHClient, port: int) -> int:
    """获取代理端口的活跃连接数。"""
    out = ssh_exec(ssh,
        f"ss -tn state established '( sport = :{port} )' 2>/dev/null | tail -n +2 | wc -l")
    try:
        return int(out.strip())
    except ValueError:
        return 0


def collect_system(ssh: paramiko.SSHClient) -> dict:
    """采集系统指标。"""
    cpu = 0.0
    cpu_out = ssh_exec(ssh, "top -bn1 2>/dev/null | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
    if cpu_out:
        try:
            cpu = float(cpu_out)
        except ValueError:
            pass

    mem = 0.0
    mem_out = ssh_exec(ssh, "free -m 2>/dev/null | grep Mem | awk '{print $3/$2*100}'")
    if mem_out:
        try:
            mem = float(mem_out)
        except ValueError:
            pass

    disk = 0.0
    disk_out = ssh_exec(ssh, "df -h / 2>/dev/null | tail -1 | awk '{print $5}' | sed 's/%//'")
    if disk_out:
        try:
            disk = float(disk_out)
        except ValueError:
            pass

    return {
        "cpuPercent": round(cpu, 1),
        "memPercent": round(mem, 1),
        "diskPercent": round(disk, 1),
    }


# ---------------------------------------------------------------------------
# 海外节点探测（本地执行）
# ---------------------------------------------------------------------------

def probe_tcp(host: str, port: int, timeout: float = 5.0) -> tuple[bool, int]:
    """TCP 探测，返回 (reachable, latency_ms)。"""
    try:
        start = time.perf_counter()
        with socket.create_connection((host, port), timeout=timeout):
            latency = int((time.perf_counter() - start) * 1000)
        return True, latency
    except Exception:
        return False, -1


# ---------------------------------------------------------------------------
# 采集主循环
# ---------------------------------------------------------------------------

def collect_snapshot(config: dict, ssh_client: paramiko.SSHClient) -> dict:
    """采集一次完整快照。"""
    snap = {
        "timestamp": int(time.time()),
        "serverMetrics": collect_system(ssh_client),
        "projects": [],
        "proxy": {},
        "overseasNodes": [],
    }

    for proj in config.get("projects", []):
        alive, pid, cpu, mem = check_process(ssh_client, proj["processName"])
        snap["projects"].append({
            "name": proj["name"],
            "alive": alive,
            "pid": pid,
            "port": proj.get("port", 0),
            "portOpen": check_port(ssh_client, proj["port"]) if proj.get("port") else False,
            "cpuPercent": round(cpu, 1),
            "memMB": mem,
        })

    proxy_cfg = config.get("proxy", {})
    if proxy_cfg:
        px_alive, _, _, _ = check_process(ssh_client, proxy_cfg["processName"])
        snap["proxy"] = {
            "alive": px_alive,
            "port": proxy_cfg.get("port", 0),
            "portOpen": check_port(ssh_client, proxy_cfg["port"]) if proxy_cfg.get("port") else False,
            "activeConnections": check_proxy_connections(ssh_client, proxy_cfg["port"]) if proxy_cfg.get("port") else 0,
        }

    for node in config.get("overseasNodes", []):
        reachable, latency = probe_tcp(node["host"], node.get("port", 443))
        snap["overseasNodes"].append({
            "name": node["name"],
            "location": [node["lng"], node["lat"]],
            "reachable": reachable,
            "latencyMs": latency,
        })

    return snap


def collection_loop(host: str, port: int, user: str, password: str, config: dict):
    """后台采集线程 — 每 5 秒采集一次。"""
    global SNAPSHOT
    ssh: paramiko.SSHClient | None = None

    while True:
        try:
            if ssh is None or not ssh.get_transport() or not ssh.get_transport().is_active():
                if ssh:
                    try:
                        ssh.close()
                    except Exception:
                        pass
                print(f"[{_ts()}] SSH connecting to {host}:{port}...")
                ssh = ssh_connect(host, port, user, password)
                print(f"[{_ts()}] SSH connected")

            snap = collect_snapshot(config, ssh)
            with SNAPSHOT_LOCK:
                SNAPSHOT = snap

        except Exception as e:
            print(f"[{_ts()}] Collection error: {e}")
            ssh = None
            with SNAPSHOT_LOCK:
                SNAPSHOT["_error"] = str(e)
            time.sleep(5)
            continue

        time.sleep(5)


# ---------------------------------------------------------------------------
# WebSocket 服务
# ---------------------------------------------------------------------------

async def ws_handler(websocket):
    """每个 WebSocket 客户端连接后持续推送快照。"""
    print(f"[{_ts()}] WS client connected")
    try:
        while True:
            with SNAPSHOT_LOCK:
                data = json.dumps(SNAPSHOT, ensure_ascii=False)
            await websocket.send(data)
            await asyncio.sleep(3)
    except websockets.exceptions.ConnectionClosed:
        pass
    finally:
        print(f"[{_ts()}] WS client disconnected")


async def start_ws_server(listen_addr: str):
    """启动 WebSocket 服务。"""
    host, port_str = listen_addr.rsplit(":", 1)
    port = int(port_str)
    print(f"[{_ts()}] WebSocket server on {host}:{port}")
    async with websockets.serve(ws_handler, host.strip() or "0.0.0.0", port):
        await asyncio.Future()  # run forever


# ---------------------------------------------------------------------------
# 入口
# ---------------------------------------------------------------------------

def _ts() -> str:
    return datetime.now().strftime("%H:%M:%S")


def main():
    parser = argparse.ArgumentParser(description="NetController SSH Bridge")
    parser.add_argument("--host", required=True, help="阿里云服务器 IP")
    parser.add_argument("--port", type=int, default=22, help="SSH 端口")
    parser.add_argument("--user", required=True, help="SSH 用户名")
    parser.add_argument("--password", required=True, help="SSH 密码")
    parser.add_argument("--config", default="config.yaml", help="配置文件路径")
    args = parser.parse_args()

    # 环境变量也可覆盖（避免密码出现在命令行历史中）
    host = os.environ.get("NC_SSH_HOST", args.host)
    ssh_port = int(os.environ.get("NC_SSH_PORT", args.port))
    user = os.environ.get("NC_SSH_USER", args.user)
    password = os.environ.get("NC_SSH_PASSWORD", args.password)

    with open(args.config, "r", encoding="utf-8") as f:
        config = yaml.safe_load(f)

    listen_addr = config.get("listen", ":9527")

    # 启动后台采集线程
    t = threading.Thread(
        target=collection_loop,
        args=(host, ssh_port, user, password, config),
        daemon=True,
    )
    t.start()

    # 启动 WebSocket 服务（阻塞）
    asyncio.run(start_ws_server(listen_addr))


if __name__ == "__main__":
    main()
