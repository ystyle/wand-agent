# wand-agent

WebSocket PTY agent for [wand](https://github.com/ystyle/wand) — a HarmonyOS terminal emulator.

Creates a PTY session for each WebSocket connection, forwarding terminal I/O between the HarmonyOS app and a shell running in an openEuler container.

## Quick Start

```bash
# ARM64（LOH / 轻量级鸿蒙）
wget https://github.com/ystyle/wand-agent/releases/latest/download/wand-agent-linux-arm64 -O wand-agent
chmod +x wand-agent

# x86_64
wget https://github.com/ystyle/wand-agent/releases/latest/download/wand-agent-linux-amd64 -O wand-agent
chmod +x wand-agent

# 启动
./wand-agent --token harmonyterm
```

## 后台管理（推荐：zinit + zshrc）

### 安装 zinit

```bash
mkdir -p ~/bin
wget https://github.com/threefoldtech/zinit/releases/latest/download/zinit-linux-arm64 -O ~/bin/zinit
chmod +x ~/bin/zinit
```

### 注册系统服务

```bash
sudo ~/bin/zinit init 2>/dev/null || true
sudo tee /etc/zinit/wand-agent.yaml << 'EOF'
exec: /home/user/wand-agent --token harmonyterm
after:
  - net-eth0
EOF
sudo zinit start wand-agent
```

### zshrc 快捷管理

在 `~/.zshrc` 中添加：

```zsh
alias wa-up='sudo zinit start wand-agent'
alias wa-down='sudo zinit stop wand-agent'
alias wa-restart='sudo zinit restart wand-agent'
alias wa-status='sudo zinit status wand-agent'
alias wa-log='sudo zinit log wand-agent'
```

重载配置：

```bash
source ~/.zshrc
```

## 连接 wand

1. 在 openEuler 容器中查看 IP：

   ```bash
   ip addr show eth0
   ```

   默认 IP: `172.16.100.2` · 端口: `8765`

2. 在 wand 中点击连接状态圆点 → **连接管理...**
3. 填入上述 IP、端口和 Token（默认 `harmonyterm`）
4. 点击「保存 & 重连」

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--host` | `0.0.0.0` | 监听地址 |
| `--port` | `8765` | 监听端口 |
| `--token` | `harmonyterm` | WebSocket 连接认证 Token |
| `--max-sessions` | `10` | 最大会话数 |

## 协议

WebSocket 端点: `/ws?token=<token>&cols=80&rows=24&cwd=/path`

- **Binary frames** (client→server): 终端输入（键盘数据）
- **Binary frames** (server→client): PTY 输出（ANSI/VT 序列）
- **Text frames** (client→server): JSON 控制消息

### 控制消息

| Type | Direction | Description |
|------|-----------|-------------|
| `{"type":"resize","cols":80,"rows":24}` | client→server | Resize PTY |
| `{"type":"cwd"}` | client→server | Query working directory |
| `{"type":"cwd","dir":"/path"}` | server→client | Working directory response |
| `{"type":"fork","cwd":"/path"}` | client→server | Fork new session at path |
| `{"type":"forked","id":"..."}` | server→client | New session ID |
| `{"type":"ping","ts":123}` | bidirectional | Heartbeat |
| `{"type":"error","error":"..."}` | server→client | Error notification |

## 构建

```bash
go build -buildvcs=false -o wand-agent .
```

## License

MIT
