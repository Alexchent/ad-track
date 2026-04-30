# ad-track

`ad-track` 是一个基于 Go + Gin 的广告归因回传服务，当前主要对接 vivo 营销开放平台。服务接收广告平台点击监测数据，将设备标识与点击参数写入 Redis；当业务侧上报激活/归因请求时，服务根据 OAID 或 IMEI 查找点击数据，并调用 vivo 行为数据上传接口完成转化回传。

## 功能特性

- vivo OAuth 授权码换取广告主 `access_token`，并按广告主维度存储到 Redis。
- vivo 点击监测数据接收，支持批量 JSON 数据写入。
- 基于 OAID / IMEI 的点击数据匹配与归因回传。
- 自动识别明文与 MD5 设备标识类型：`OAID`、`OAID_MD5`、`IMEI`、`IMEI_MD5`。
- Redis 保存点击数据，默认有效期 30 天。
- Gin HTTP 服务，支持 request id、访问日志、健康检查、Prometheus 指标接口。
- 使用 `endless` 支持平滑重启与优雅关闭。
- 使用 `slog` + `lumberjack` 输出结构化日志并支持日志轮转。

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

### 1. 准备依赖

- Go 环境
- Redis 服务
- vivo 营销开放平台应用信息：
  - `client_id`
  - `client_secret`
  - 广告主授权码
  - 应用包名与 vivo `srcId` 映射

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 创建配置文件

服务默认读取当前目录下的 `conf.yaml`，也可以通过 `-f` 指定配置文件路径。

示例：

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

### 4. 启动服务

```bash
go run . -f conf.yaml
```

构建后运行：

```bash
go build -o ad-track .
./ad-track -f conf.yaml
```

启动成功后控制台会输出：

```text
服务启动, pid=<pid>, adder=:8080
```

## 配置说明

| 配置项 | 类型 | 说明 |
| --- | --- | --- |
| `Port` | string | HTTP 监听地址，例如 `:8080` |
| `Env` | string | 运行环境，可选 |
| `CachePrefix` | string | 缓存前缀，可选 |
| `Redis.Addr` | string | Redis 地址 |
| `Redis.Password` | string | Redis 密码 |
| `Redis.Db` | int | Redis DB |
| `MySQL.DSN` | string | MySQL DSN，当前代码中暂未使用 |
| `Log.Filename` | string | 日志文件路径 |
| `Log.Encoding` | string | 日志格式，支持 `json`、`console` |
| `Log.Level` | string | 日志级别，支持 `debug`、`info`、`warn`、`error`、`fatal` |
| `Log.MaxSize` | int | 单个日志文件最大大小，单位 MB |
| `Log.MaxAge` | int | 日志保留天数 |
| `Log.Compress` | bool | 是否压缩历史日志 |
| `VIVO.Host` | string | vivo 营销 API 地址，空值时使用官方默认地址 |
| `VIVO.ClientId` | string | vivo 开放平台 Client ID |
| `VIVO.ClientSecret` | string | vivo 开放平台 Client Secret |
| `VIVO.APP` | map | 应用包名到 vivo `srcId` 的映射 |

## HTTP 接口

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

## 注意事项

- `conf.yaml` 未提交到仓库，需要部署时自行创建。
- `/report` 的 `package_name` 必须能在 `VIVO.APP` 中找到对应的 `srcId`。
- vivo 回传依赖点击数据中的 `advertiserId`，点击监测数据缺少该字段会导致回传失败。
- vivo access token 需要先通过 `/vivo/auth` 写入 Redis，否则归因回传无法获取 token。
- 当前 `GetToken` 中自动刷新 token 的代码处于注释状态，生产环境需关注 token 过期后的续期策略。
- `middleware.PrometheusMetrics()` 已实现但当前入口未注册，如需接口维度指标请在路由初始化时添加。