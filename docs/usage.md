# MEV 机器人使用文档

本文档说明本项目的环境准备、构建、部署与启动方式，面向本地开发与服务器运行场景。

## 1. 环境要求

### 1.1 基础依赖
- Go 1.21+（与 `go.mod` 保持一致）
- Linux/macOS（推荐 Linux 服务器环境）
- 可访问的 EVM 节点（本地节点或第三方 RPC）

### 1.2 推荐节点与通道
- **执行节点**：Geth / Erigon（建议至少 2 个 RPC 以做容灾）
- **私有通道**：Flashbots / MEV-Share（若可用）

## 2. 配置说明

项目通过环境变量配置，核心变量如下：

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `MEV_CHAIN` | 链名称 | `ethereum` |
| `MEV_RPC_URL` | RPC 地址 | `http://localhost:8545` |
| `MEV_PRIVATE_RELAY` | 私有交易通道 URL | 空 |
| `MEV_MIN_PROFIT_USD` | 最小利润阈值 | `5.0` |
| `MEV_MAX_GAS_GWEI` | 最大 Gas 上限 | `120` |
| `MEV_RISK_LIMIT_USD` | 风险限额 | `5000` |
| `MEV_OPPORTUNITY_FILE` | JSONL 机会数据文件（用于回放/测试） | 空 |

## 3. 构建与运行

### 3.1 本地构建
```bash
# 进入项目根目录
cd /workspace/mev_bot

# 构建
GO111MODULE=on go build -o mevbot ./cmd/mevbot
```

### 3.2 本地运行
```bash
MEV_CHAIN=ethereum \
MEV_RPC_URL=http://localhost:8545 \
MEV_MIN_PROFIT_USD=5.0 \
MEV_MAX_GAS_GWEI=120 \
./mevbot
```

### 3.3 机会回放（JSONL）

准备 JSONL 文件（每行一个机会对象）：
```json
{"type":"dex-arb","payload":{"pair":"WETH/USDC","spread":"0.4%"}}
{"type":"flashloan-arb","payload":{"asset":"USDC","amount":"500000"}}
```

运行：
```bash
MEV_OPPORTUNITY_FILE=./opportunities.jsonl ./mevbot
```

## 4. 部署建议

### 4.1 推荐目录结构
```
/opt/mevbot
├── mevbot          # 二进制文件
└── logs/           # 日志目录
```

### 4.2 systemd 服务示例
```ini
[Unit]
Description=MEV Bot Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/mevbot
ExecStart=/opt/mevbot/mevbot
Restart=always
RestartSec=3
Environment=MEV_CHAIN=ethereum
Environment=MEV_RPC_URL=http://localhost:8545
Environment=MEV_MIN_PROFIT_USD=5.0
Environment=MEV_MAX_GAS_GWEI=120

[Install]
WantedBy=multi-user.target
```

启用与启动：
```bash
sudo systemctl daemon-reload
sudo systemctl enable mevbot
sudo systemctl start mevbot
```

## 5. 日志与监控
- 默认输出到标准输出，可由 systemd 捕获。
- 可通过 `journalctl -u mevbot -f` 查看实时日志。

## 6. 常见问题

### 6.1 无法连接 RPC
- 检查 `MEV_RPC_URL` 是否正确。
- 确保节点允许外部访问并已开放端口。

### 6.2 私有通道无效
- 检查 `MEV_PRIVATE_RELAY` URL 是否可用。
- 确认通道支持目标链与交易格式。
