# MEV 机器人使用文档（目标链：BSC）

本文档按 1/2/3/4 顺序描述当前实现：
1) 链上数据采集；2) 交易构建与发送；3) 模拟与风控；4) 监控与告警。

## 1. 环境要求
- Go 1.21+
- BSC RPC 节点（建议自建 + 第三方备份）
- 可选：私有 relay（Flashbots/MEV-Share 兼容接口）

## 2. 关键配置
| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `MEV_CHAIN` | 链名称 | `bsc` |
| `MEV_RPC_URL` | BSC RPC 地址 | `https://bsc-dataseed.binance.org` |
| `MEV_WS_URL` | WebSocket 地址（后续可接入订阅） | 空 |
| `MEV_MARKETDATA_SOURCE` | `rpc/file/ticker` | `rpc` |
| `MEV_BLOCK_POLL_INTERVAL` | 区块轮询间隔 | `2s` |
| `MEV_ENABLE_TXPOOL` | 是否解析 txpool mempool | `true` |
| `MEV_POOL_CALLS` | DEX 池 eth_call 列表 | 空 |
| `MEV_ORACLE_CALLS` | 预言机 eth_call 列表 | 空 |
| `MEV_FROM_ADDRESS` | 发送地址（节点需支持签名） | 空 |
| `MEV_PRIVATE_RELAY` | 私有 bundle relay | 空 |
| `MEV_BUNDLE_SIGNATURE` | relay 认证头 | 空 |
| `MEV_MAX_SLIPPAGE_BPS` | 最大滑点 | `50` |
| `MEV_MAX_GAS_USD` | 最大 Gas 成本 | `50` |
| `MEV_AUDIT_LOG` | 审计日志 JSONL | 空 |
| `MEV_ALERT_WEBHOOK` | 告警 webhook | 空 |
| `MEV_METRICS_ADDR` | 指标端口 | `:9090` |

## 3. 构建与启动
```bash
cd /workspace/mev_bot
GO111MODULE=on go build -o mevbot ./cmd/mevbot
```

### BSC 启动示例
```bash
MEV_CHAIN=bsc \
MEV_RPC_URL=https://bsc-dataseed.binance.org \
MEV_MARKETDATA_SOURCE=rpc \
MEV_BLOCK_POLL_INTERVAL=2s \
MEV_ENABLE_TXPOOL=true \
MEV_FROM_ADDRESS=0xYourAddress \
MEV_POOL_CALLS=0xPoolAddress:0xCallData \
MEV_ORACLE_CALLS=0xOracleAddress:0xCallData \
MEV_AUDIT_LOG=./audit.jsonl \
MEV_METRICS_ADDR=:9090 \
./mevbot
```

## 4. 四阶段能力说明

### 4.1 链上数据采集
- 连接 RPC，轮询新区块 `eth_blockNumber` + `eth_getBlockByNumber`
- 解析 txpool（`txpool_content`）作为 mempool 来源
- 调用 DEX 池与预言机 `eth_call`

### 4.2 真实交易构建与发送
- 构建 EIP-1559（type 0x2）交易参数
- 公共通道：`eth_signTransaction` + `eth_sendRawTransaction`（含 nonce 管理）
- 私有通道：`eth_sendBundle` 提交

### 4.3 模拟与风控
- 模拟：`eth_call` + `eth_estimateGas`
- 风控：最小利润、最大滑点、Gas 上限、revert 拒绝
- 失败回退：发送重试 + 指数退避 + priority fee 递增

### 4.4 监控与告警
- 内置 `/metrics`（Prometheus 文本格式）
- webhook 告警
- JSONL 审计日志与回放读取


## 5. 生产部署
- 详细部署步骤请参考 `docs/deployment-production.md`。
