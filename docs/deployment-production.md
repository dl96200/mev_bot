# MEV Bot 生产部署文档（BSC）

> 说明：本部署文档面向当前代码版本，目标链为 BSC。包含两种部署方式：
> 1) systemd 裸机部署；2) Docker Compose 部署（带 Prometheus/Grafana）。

## 1. 上线前检查清单

- 已准备 BSC 高可用 RPC（建议主备两条）
- 节点支持 `eth_signTransaction`（或你已改造为外部签名服务）
- 配置了 `MEV_FROM_ADDRESS`
- 配置了风险参数（`MEV_MIN_PROFIT_USD`、`MEV_MAX_SLIPPAGE_BPS`、`MEV_MAX_GAS_USD`）
- 已规划日志、审计文件路径和备份策略（`MEV_AUDIT_LOG`）
- 已配置告警 webhook（可选）

## 2. 方式 A：systemd 部署

### 2.1 创建运行用户与目录

```bash
sudo useradd -r -s /usr/sbin/nologin mevbot || true
sudo mkdir -p /opt/mevbot /var/lib/mevbot
sudo chown -R mevbot:mevbot /opt/mevbot /var/lib/mevbot
```

### 2.2 构建与安装二进制

```bash
cd /workspace/mev_bot
GO111MODULE=on go build -o mevbot ./cmd/mevbot
sudo cp mevbot /opt/mevbot/mevbot
sudo chown mevbot:mevbot /opt/mevbot/mevbot
sudo chmod 750 /opt/mevbot/mevbot
```

### 2.3 准备环境变量文件

复制模板：

```bash
cp deploy/.env.example /opt/mevbot/.env
```

重点修改：
- `MEV_RPC_URL`
- `MEV_FROM_ADDRESS`
- `MEV_POOL_CALLS` / `MEV_ORACLE_CALLS`
- `MEV_AUDIT_LOG=/var/lib/mevbot/audit.jsonl`
- `MEV_ALERT_WEBHOOK`

### 2.4 安装 systemd 服务

```bash
sudo cp deploy/systemd/mevbot.service /etc/systemd/system/mevbot.service
sudo systemctl daemon-reload
sudo systemctl enable mevbot
sudo systemctl start mevbot
sudo systemctl status mevbot
```

### 2.5 运行验证

```bash
journalctl -u mevbot -f
curl -s http://127.0.0.1:9090/healthz
curl -s http://127.0.0.1:9090/metrics
```

## 3. 方式 B：Docker Compose 部署

### 3.1 准备环境配置

```bash
cd /workspace/mev_bot/deploy
cp .env.example .env
```

编辑 `.env` 后启动：

```bash
docker compose up -d
```

### 3.2 访问地址

- MEV Bot metrics: `http://<host>:9090/metrics`
- Prometheus: `http://<host>:9091`
- Grafana: `http://<host>:3000`（默认账号 `admin/admin`）

### 3.3 验证

```bash
docker compose ps
docker compose logs -f mevbot
curl -s http://127.0.0.1:9090/healthz
```

## 4. 监控建议

建议至少告警以下指标：
- `mev_failures_total` 5 分钟突增
- `mev_success_total` 长时间无增长
- `mev_opportunities_total` 长时间为 0

## 5. 升级与回滚

### 5.1 systemd 升级

```bash
cd /workspace/mev_bot
git pull
GO111MODULE=on go build -o mevbot ./cmd/mevbot
sudo cp mevbot /opt/mevbot/mevbot
sudo systemctl restart mevbot
```

### 5.2 回滚

保留上一版本二进制（如 `/opt/mevbot/mevbot.prev`）：

```bash
sudo cp /opt/mevbot/mevbot.prev /opt/mevbot/mevbot
sudo systemctl restart mevbot
```

## 6. 故障排查

### 6.1 交易签名失败
- 检查节点是否支持 `eth_signTransaction`
- 检查 `MEV_FROM_ADDRESS` 是否存在且可签名

### 6.2 无机会数据
- 检查 `MEV_MARKETDATA_SOURCE=rpc`
- 检查 `MEV_ENABLE_TXPOOL=true`
- 检查 `MEV_POOL_CALLS` 与 `MEV_ORACLE_CALLS` 格式是否为 `address:calldata`

### 6.3 指标不可达
- 检查 `MEV_METRICS_ADDR`
- 检查端口映射或防火墙策略
