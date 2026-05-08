# ad-track

[![Go](https://img.shields.io/badge/Go-1.25.8-00ADD8?logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-HTTP%20Framework-00ADD8)](https://gin-gonic.com/)
[![Redis](https://img.shields.io/badge/Redis-required-DC382D?logo=redis)](https://redis.io/)
[![Prometheus](https://img.shields.io/badge/Prometheus-metrics-E6522C?logo=prometheus)](https://prometheus.io/)

`ad-track` 是一个基于 Go + Gin 的广告归因回传服务。当前实现主要对接 vivo 营销开放平台：服务接收广告平台点击监测数据，将设备标识与点击参数写入 Redis；当业务侧上报激活/归因请求时，服务根据 OAID 或 IMEI 匹配点击数据，并调用 vivo 行为数据上传接口完成转化回传。

> 当前项目仍以 vivo 归因为核心场景。多渠道归因、自动 token 刷新、完整部署清单等能力尚未完整实现，详见 [Roadmap](#roadmap) 与 [注意事项](#注意事项)。

## 目录

- [ad-track](#ad-track)
  - [目录](#目录)
  - [功能特性](#功能特性)
  - [架构概览](#架构概览)
  - [技术栈](#技术栈)
  - [项目结构](#项目结构)
  - [快速开始](#快速开始)
    - [环境要求](#环境要求)
    - [安装依赖](#安装依赖)
    - [创建配置文件](#创建配置文件)
    - [本地运行](#本地运行)
    - [构建运行](#构建运行)
    - [验证服务](#验证服务)
  - [配置说明](#配置说明)
  - [HTTP API](#http-api)
    - [健康检查](#健康检查)
    - [Prometheus 指标](#prometheus-指标)
    - [vivo 授权回调](#vivo-授权回调)
    - [vivo 点击监测接收](#vivo-点击监测接收)
    - [归因回传](#归因回传)
  - [vivo 回传字段说明](#vivo-回传字段说明)
  - [数据存储](#数据存储)
    - [点击数据](#点击数据)
    - [vivo Token](#vivo-token)
  - [日志与链路追踪](#日志与链路追踪)
  - [监控指标](#监控指标)
  - [平滑重启](#平滑重启)
  - [测试](#测试)
  - [贡献指南](#贡献指南)
    - [Commit 建议](#commit-建议)
  - [Roadmap](#roadmap)
  - [注意事项](#注意事项)
  - [License](#license)

## 功能特性

- vivo OAuth 授权码换取广告主 `access_token`，并按广告主维度存储到 Redis。
- vivo 点击监测数据接收，支持批量 JSON 数据写入。
- 基于 OAID / IMEI 的点击数据匹配与归因回传。
- 自动识别明文与 MD5 设备标识类型：`OAID`、`OAID_MD5`、`IMEI`、`IMEI_MD5`。
- Redis 保存点击数据，默认有效期 30 天。
- Gin HTTP 服务，支持 request id、访问日志、健康检查、Prometheus 指标接口。
- 使用 `endless` 支持平滑重启与优雅关闭。
- 使用 `slog` + `lumberjack` 输出结构化日志并支持日志轮转。

## 架构概览

```text
ad platform / vivo click callback
        |
        v
POST /vivo/click
        |
        v
Redis click:<oaid|imei>  <--------------------+
        |                                      |
        |                                      |
business activation/report                     |
        |                                      |
        v                                      |
GET /report?oaid=...&imei=...&package_name=...+
        |
        v
match click data -> resolve vivo token -> upload ACTIVATION behavior to vivo
```

核心流程：

1. `/vivo/click` 接收点击监测数据，将点击参数按设备 ID 写入 Redis。
2. `/vivo/auth` 使用 vivo 授权码换取广告主 token，并将 token 写入 Redis。
3. `/report` 根据 OAID / IMEI 查询点击数据，匹配 vivo 流量后调用 vivo 行为上传接口。

## 技术栈

- Go 1.25.8
- Gin
- go-zero `conf` 配置加载
- Redis
- Prometheus client
- slog / zap / lumberjack
- fvbock/endless

## 项目结构

```text
.
├── main.go                         # 服务入口：加载配置、初始化日志、注册路由、启动 HTTP 服务
├── router.go                       # 路由注册
├── config/                         # 配置结构定义
├── handler/                        # HTTP Handler
│   ├── attribute.go                # 归因上报入口
│   ├── click.go                    # 点击数据保存辅助逻辑
│   └── vivo.go                     # vivo 授权与点击监测接口
├── logic/                          # 业务逻辑
│   ├── click.go                    # 点击数据 Redis 读写
│   ├── interface.go                # 归因接口定义
│   └── vivo.go                     # vivo 归因回传实现
├── middleware/                     # Gin 中间件
│   ├── cors.go
│   ├── logger.go                   # 请求日志与 trace id 注入
│   ├── metrics.go                  # Prometheus 指标中间件
│   └── ratelimit.go
├── pkg/
│   ├── cache/                      # 缓存封装
│   ├── logger/                     # 日志初始化
│   └── vivo/                       # vivo 营销 API 客户端
├── svc/
│   └── servicecontext.go           # 依赖初始化与服务上下文
└── doc/                            # 对接文档资料
```

## 快速开始

### 环境要求

- Go 1.25.8 或兼容版本
- Redis 6.x+
- vivo 营销开放平台应用信息：
  - `client_id`
  - `client_secret`
  - 广告主授权码
  - 应用包名与 vivo `srcId` 映射

### 安装依赖

```bash
go mod tidy
```

### 创建配置文件

服务默认读取当前目录下的 `conf.yaml`，也可以通过 `-f` 指定配置文件路径。

```yaml
Port: ":8080"
Env: "dev"
CachePrefix: "ad-track"

Redis:
  Addr: "127.0.0.1:6379"
  Password: ""
  Db: 0

MySQL:
  DSN: ""

Log:
  Filename: "./logs/ad-track.log"
  Encoding: "json"
  Level: "info"
  MaxSize: 100
  MaxAge: 7
  Compress: true

VIVO:
  Host: "https://marketing-api.vivo.com.cn"
  ClientId: "your-client-id"
  ClientSecret: "your-client-secret"
  APP:
    "com.example.app": "your-vivo-src-id"
```

### 本地运行

```bash
go run . -f conf.yaml
```

启动成功后控制台会输出类似内容：

```text
服务启动, pid=<pid>, adder=:8080
```

### 构建运行

```bash
go build -o ad-track .
./ad-track -f conf.yaml
```

### 验证服务

```bash
curl http://127.0.0.1:8080/health
```

预期响应：

```text
ok
```

## 配置说明

| 配置项 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `Port` | string | 是 | HTTP 监听地址，例如 `:8080` |
| `Env` | string | 否 | 运行环境 |
| `CachePrefix` | string | 否 | 缓存前缀 |
| `Redis.Addr` | string | 是 | Redis 地址 |
| `Redis.Password` | string | 否 | Redis 密码 |
| `Redis.Db` | int | 否 | Redis DB |
| `MySQL.DSN` | string | 否 | MySQL DSN，当前代码中暂未使用 |
| `Log.Filename` | string | 是 | 日志文件路径 |
| `Log.Encoding` | string | 否 | 日志格式，支持 `json`、`console` |
| `Log.Level` | string | 否 | 日志级别，支持 `debug`、`info`、`warn`、`error`、`fatal` |
| `Log.MaxSize` | int | 否 | 单个日志文件最大大小，单位 MB |
| `Log.MaxAge` | int | 否 | 日志保留天数 |
| `Log.Compress` | bool | 否 | 是否压缩历史日志 |
| `VIVO.Host` | string | 否 | vivo 营销 API 地址，空值时使用官方默认地址 |
| `VIVO.ClientId` | string | 是 | vivo 开放平台 Client ID |
| `VIVO.ClientSecret` | string | 是 | vivo 开放平台 Client Secret |
| `VIVO.APP` | map | 是 | 应用包名到 vivo `srcId` 的映射 |

## HTTP API

### 健康检查

```http
GET /health
```

响应：

```text
ok
```

### Prometheus 指标

```http
GET /metrics
```

返回 Prometheus 格式的指标数据。

### vivo 授权回调

```http
GET /vivo/auth?code=<authorization_code>
```

用途：

1. 使用 vivo 授权码换取 `access_token` 和 `refresh_token`。
2. 查询当前 token 对应的广告主 UUID。
3. 将 token 信息写入 Redis。

成功响应：

```text
ok
```

Redis token key 格式：

```text
vivo_token_<clientId>_<advertiserId>
```

### vivo 点击监测接收

```http
POST /vivo/click?channel=vivo
Content-Type: application/json
```

请求体为数组，每个元素是一条点击数据：

```json
[
  {
    "oaid": "device-oaid",
    "imei": "device-imei",
    "clickId": "vivo-click-id",
    "advertiserId": "vivo-advertiser-id",
    "pkgName": "com.example.app"
  }
]
```

处理逻辑：

- 将 URL 查询参数 `channel` 写入每条点击数据。
- 如果存在 `oaid`，以 `click:<oaid>` 写入 Redis Hash。
- 如果存在 `imei`，以 `click:<imei>` 写入 Redis Hash。
- 点击数据默认保存 30 天。

成功响应：

```json
{
  "code": 0,
  "msg": "操作成功"
}
```

### 归因回传

```http
GET /report?oaid=<oaid>&imei=<imei>&user_id=<user_id>&package_name=<package_name>
```

参数说明：

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `oaid` | 否 | 设备 OAID，`oaid` 和 `imei` 至少传一个 |
| `imei` | 否 | 设备 IMEI，`oaid` 和 `imei` 至少传一个 |
| `user_id` | 否 | 应用侧用户 ID，会作为 `app_uid` 写入归因数据 |
| `package_name` | 是 | 应用包名，用于查找 `VIVO.APP` 中配置的 `srcId` |

处理逻辑：

1. 根据 `oaid`、`imei` 依次查询 Redis 点击数据。
2. 未匹配到点击数据时返回 `click data not found`。
3. 点击数据中的 `channel` 不包含 `vivo` 时直接返回 `not vivo channel`。
4. 将 `user_id` 写入 `app_uid`，将 `package_name` 写入 `pkgName`。
5. 根据点击数据中的 `advertiserId` 获取 Redis 中保存的 vivo access token。
6. 调用 vivo 行为上传接口，上报 `ACTIVATION` 转化事件。

成功响应：

```json
{
  "code": 0,
  "msg": "操作成功"
}
```

## vivo 回传字段说明

归因回传时会构造 vivo `BehaviorRequest`：

```json
{
  "srcType": "APP",
  "pkgName": "com.example.app",
  "srcId": "vivo-src-id",
  "dataList": [
    {
      "userIdType": "OAID",
      "userId": "device-id",
      "clickId": "vivo-click-id",
      "cvType": "ACTIVATION",
      "cvTime": 1710000000000
    }
  ]
}
```

设备标识类型判断规则：

| 输入字段 | 长度 | vivo `userIdType` |
| --- | --- | --- |
| `oaid` | 32 | `OAID_MD5` |
| `oaid` | 非 32 | `OAID` |
| `imei` | 32 | `IMEI_MD5` |
| `imei` | 非 32 | `IMEI` |

点击 ID 兼容字段：

- 优先读取 `clickId`
- 如果为空，读取 `ClickId`

## 数据存储

### 点击数据

Redis Hash key：

```text
click:<device_id>
```

示例：

```text
click:device-oaid
click:device-imei
```

TTL：

```text
30 天
```

### vivo Token

Redis String key：

```text
vivo_token_<clientId>_<advertiserId>
```

value 为 JSON：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "token_date": 1710000000000,
  "refresh_token_date": 1710000000000
}
```

如果 `refresh_token_date` 存在，Redis key 会按该时间设置过期时间。

## 日志与链路追踪

服务启动时通过 `pkg/logger` 初始化全局 `slog` logger：

- 日志同时输出到控制台和文件。
- 支持 JSON 或文本格式。
- 支持日志轮转与压缩。
- `middleware.RequestLogger` 会为每个请求注入 request id 作为 `traceId`。
- 访问日志包含 method、path、query、client_ip、status、latency。

## 监控指标

项目定义了以下 Prometheus 指标：

| 指标 | 类型 | 说明 |
| --- | --- | --- |
| `http_requests_total` | CounterVec | HTTP 请求总数 |
| `http_request_duration_seconds` | HistogramVec | HTTP 请求耗时 |
| `http_requests_in_flight` | GaugeVec | 当前处理中的请求数 |

当前路由已暴露 `/metrics` 指标接口。如需采集业务接口请求指标，需要在 Gin 路由上启用 `middleware.PrometheusMetrics()`。

## 平滑重启

服务使用 `github.com/fvbock/endless` 启动 HTTP Server：

- `SIGHUP`：平滑重启
- `SIGTERM` / `SIGINT`：优雅关闭

示例：

```bash
kill -HUP <pid>
kill -TERM <pid>
```

## 测试

运行全部测试：

```bash
go test ./...
```

当前项目包含 vivo summary 相关测试文件：

```text
pkg/vivo/vivo_summary_test.go
```

## 贡献指南

欢迎提交 Issue 和 Pull Request。建议在贡献前遵循以下流程：

1. Fork 本仓库并创建特性分支。
2. 保持代码风格与现有项目一致。
3. 为新增逻辑补充必要的单元测试或说明验证方式。
4. 运行测试并确保通过：

```bash
go test ./...
```

5. 提交 PR 时说明变更动机、实现方式、影响范围和测试结果。

### Commit 建议

推荐使用语义化提交信息：

```text
feat: add new attribution channel
fix: handle missing vivo token
docs: update deployment guide
test: add click storage tests
```

## Roadmap

- [ ] 提供 `conf.example.yaml` 示例配置文件。
- [ ] 注册并验证 `middleware.PrometheusMetrics()`，完善接口维度指标采集。
- [ ] 明确或修复 vivo token 缺失时的 fallback 行为。
- [ ] 启用并测试 vivo token 自动刷新逻辑。
- [ ] 抽象多渠道归因接口，支持更多广告平台。
- [ ] 增加 Redis 与 vivo 回传流程的集成测试。
- [ ] 补充 Dockerfile、Compose 或 Kubernetes 部署示例。

## 注意事项

- `conf.yaml` 未提交到仓库，需要部署时自行创建。
- `/report` 的 `package_name` 必须能在 `VIVO.APP` 中找到对应的 `srcId`。
- vivo 回传依赖点击数据中的 `advertiserId`，点击监测数据缺少该字段会导致回传失败。
- vivo access token 需要先通过 `/vivo/auth` 写入 Redis，否则归因回传无法获取 token。
- 当前 `GetToken` 中自动刷新 token 的代码处于注释状态，生产环境需关注 token 过期后的续期策略。
- `middleware.PrometheusMetrics()` 已实现但当前入口未注册，如需接口维度指标请在路由初始化时添加。
- 请勿将真实的 `client_secret`、授权码、token 或生产 Redis 密码提交到仓库。

## License

当前仓库尚未包含明确的开源许可证文件。若计划作为正式开源项目发布，请在仓库根目录添加 `LICENSE` 文件，并在本节声明对应许可证，例如 MIT、Apache-2.0 或 GPL-3.0。