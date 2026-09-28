# ExchangeApp — 汇率兑换与 AI 财经资讯平台

**ExchangeApp** 是一个前后端分离的货币汇率兑换与财经资讯 Web 应用。后端基于 **Go + Gin + GORM + Redis** 构建，前端基于 **Vue 3 + TypeScript + Vite + Element Plus** 开发，并集成了 **DeepSeek 大模型**，通过 **SSE（Server-Sent Events）** 实现文章的流式智能摘要与中英互译。

---

## ✨ 核心功能

- **🔐 用户认证与鉴权**
  - 支持用户注册与登录，密码采用 `bcrypt` 加密存储。
  - 基于 JWT 进行接口鉴权，中间件同时支持 `Authorization` 请求头与 URL Query 参数 `?token=`（兼容前端 `EventSource` SSE 连接鉴权）。
- **💱 货币汇率管理**
  - 查看实时货币兑换汇率列表，支持不同币种间的金额换算。
  - 登录用户可录入新的货币对汇率数据。
- **📰 财经资讯与高并发缓存**
  - **文章发布与详情阅读**：支持发布带标题、预览与正文的财经资讯文章。
  - **分页查询 + Redis 旁路缓存**：文章列表接口支持分页（`?page=1&pageSize=20`）与缓存开关（`?cache=on|off`），缓存 TTL 为 10 分钟；新增文章时自动失效分页缓存。
  - **文章点赞互动**：利用 Redis `INCR` 实现原子计数并同步更新至 MySQL，缓存未命中时自动回源数据库并回填 Redis。
- **🤖 AI 智能助手（SSE 流式输出）**
  - 集成 **DeepSeek Chat API**，自动检测文章语种（中文 / 英文）。
  - **智能摘要（Summarize）**：自动提炼 150 字以内的同语种核心摘要。
  - **中英互译（Translate）**：中文文章自动译为英文，英文文章自动译为中文。
  - **流式响应 + 结果缓存**：基于 `text/event-stream` 实时流式推送给前端（配合打字机光标动效），生成完成后缓存至 Redis（TTL 1 小时），再次请求时直接从 Redis 模拟分块流式返回。
- **📊 日志与性能压测**
  - 内置访问日志（`logs/access.log`）与 Panic 异常恢复日志（`logs/error.log`）中间件。
  - 提供批量造数脚本与并发压测脚本，直观对比开启/关闭 Redis 缓存的性能差异。

---

## 🛠️ 技术栈

### 后端 (`backend/`)
| 类别 | 技术 / 库 | 说明 |
| :--- | :--- | :--- |
| 开发语言 | Go `1.24` | 后端核心开发语言 |
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) `v1.10.1` | 高性能 HTTP Web 框架 |
| ORM 框架 | [GORM](https://gorm.io/) `v1.30.0` | MySQL 数据库对象关系映射 |
| 数据库 | MySQL | 持久化存储用户、文章、汇率数据 |
| 缓存数据库 | Redis (`go-redis`) | 分页缓存、点赞计数、AI 结果缓存 |
| 配置管理 | [Viper](https://github.com/spf13/viper) `v1.20.1` | 解析 `config/config.yml` 配置 |
| 身份认证 | `golang-jwt/jwt/v5` + `x/crypto/bcrypt` | JWT 令牌签发校验与密码哈希 |
| 大模型集成 | DeepSeek Chat API | 基于 HTTP Stream + SSE 的文章摘要与翻译 |

### 前端 (`frontend/`)
| 类别 | 技术 / 库 | 说明 |
| :--- | :--- | :--- |
| 前端框架 | Vue `3.4` (Composition API) | 响应式视图层框架 |
| 开发语言 | TypeScript `5.2` | 静态类型检查 |
| 构建工具 | Vite `5.4` | 前端极速开发服务器与打包工具 |
| UI 组件库 | Element Plus `2.7` / Vant `4.9` | 桌面端与移动端 UI 组件支持 |
| 路由与状态 | Vue Router `4.3` + Pinia `2.1` | 前端路由管理与用户认证状态管理 |
| 网络请求 | Axios `1.7` + `EventSource` | RESTful API 调用与 SSE 流式数据接收 |

---

## 📂 项目目录结构

```text
ExchangeApp/
├── backend/                        # Go 后端工程
│   ├── config/
│   │   ├── config.go               # Viper 配置加载与初始化入口
│   │   ├── config.yml              # 端口、MySQL、DeepSeek API 配置文件
│   │   ├── db.go                   # GORM MySQL 连接与连接池配置
│   │   └── reidis.go               # Redis 客户端初始化
│   ├── controllers/
│   │   ├── ai_controller.go        # DeepSeek AI 摘要/翻译 SSE 流式接口
│   │   ├── artical_controller.go   # 文章创建、详情查询、分页+Redis缓存查询
│   │   ├── auth_controller.go      # 用户登录与注册接口
│   │   ├── exchange_rate_controller.go # 汇率查询与创建接口
│   │   └── like_controller.go      # 文章点赞与点赞数查询接口
│   ├── global/
│   │   └── global.go               # 全局变量（MySQL Db、RedisDB 实例）
│   ├── middlewares/
│   │   ├── auth_middlewares.go     # JWT 认证中间件
│   │   ├── logging.go              # 访问日志与 Panic 恢复日志中间件
│   │   └── logging_test.go         # 日志中间件单元测试
│   ├── models/
│   │   ├── article.go              # 文章数据模型
│   │   ├── exchange_rate.go        # 汇率数据模型
│   │   └── user.go                 # 用户数据模型
│   ├── router/
│   │   └── router.go               # Gin 路由注册与 CORS 跨域配置
│   ├── utils/
│   │   └── utils.go                # 密码加密校验、JWT 生成与解析工具
│   ├── artical_insert.go           # 批量插入 1000 条测试文章脚本
│   ├── benchmark.go                # Redis 缓存开启/关闭并发性能压测脚本
│   ├── main.go                     # 后端程序主入口
│   └── go.mod                      # Go 模块依赖定义
└── frontend/                       # Vue 3 前端工程
    ├── src/
    │   ├── components/             # 登录 (Login.vue)、注册 (Register.vue) 组件
    │   ├── router/index.ts         # 前端路由配置
    │   ├── store/auth.ts           # Pinia 认证状态管理
    │   ├── types/Article.d.ts      # TypeScript 接口类型定义
    │   ├── views/
    │   │   ├── HomeView.vue            # 首页视图
    │   │   ├── CurrencyExchangeView.vue# 货币汇率换算视图
    │   │   ├── NewsView.vue            # 财经文章列表与发布视图
    │   │   └── NewsDetailView.vue      # 文章详情、点赞与 AI 助手视图
    │   ├── App.vue                 # 根组件与顶部导航栏
    │   ├── axios.ts                # Axios 实例与请求拦截器配置
    │   └── main.ts                 # 前端应用入口
    ├── index.html
    ├── package.json
    └── vite.config.ts              # Vite 配置与开发代理
```

---

## 🚀 快速开始

### 1. 环境要求
- **Go**: `>= 1.24`
- **Node.js**: `>= 18.x`
- **MySQL**: `>= 8.0`（或 `5.7+`）
- **Redis**: `>= 6.0`（默认连接 `localhost:6379`）

### 2. 配置后端 (`backend/config/config.yml`)

在启动后端前，请确保本地 MySQL 和 Redis 服务已运行，并检查 `backend/config/config.yml` 配置：

```yaml
App:
  name: CurrencyExchangeApp
  port: :3000

Database:
  dsn: user:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local
  MaxIdleConns: "11"
  MaxOpenCons: "114"

DeepSeek:
  api_key: "your-deepseek-api-key"
  base_url: "https://api.deepseek.com/v1/chat/completions"
  model: "deepseek-chat"
```

> **提示**：数据库表结构（`users`、`articles`、`exchange_rates`）会在调用相应创建接口时通过 GORM `AutoMigrate` 自动迁移建表。

### 3. 启动后端服务

```bash
cd backend
go mod tidy
go run main.go
```
后端服务默认监听在 `http://localhost:3000`。

### 4. 启动前端服务

```bash
cd frontend
npm install
npm run dev
```
前端开发服务器默认运行在 `http://localhost:5173`。

---

## 📡 API 接口一览

| 模块 | 方法 | 路径 | 鉴权 | 说明 |
| :--- | :--- | :--- | :---: | :--- |
| **认证** | `POST` | `/api/auth/register` | 否 | 用户注册并返回 JWT Token |
| **认证** | `POST` | `/api/auth/login` | 否 | 用户登录并返回 JWT Token |
| **汇率** | `GET` | `/api/exchangeRates` | 否 | 获取所有货币汇率列表 |
| **汇率** | `POST` | `/api/exchangeRates` | 是 | 新增一条货币汇率记录 |
| **文章** | `GET` | `/api/articles` | 是 | 分页获取文章列表（支持 `?page=1&pageSize=20&cache=on\|off`） |
| **文章** | `POST` | `/api/articles` | 是 | 发布新文章（自动清理分页缓存） |
| **文章** | `GET` | `/api/articles/:id` | 是 | 根据 ID 获取文章详情 |
| **点赞** | `POST` | `/api/articles/:id/like` | 是 | 为指定文章点赞 |
| **点赞** | `GET` | `/api/articles/:id/like` | 是 | 获取指定文章的点赞总数 |
| **AI 助手** | `GET` | `/api/articles/:id/ai` | 是 | SSE 流式生成摘要或翻译（`?action=summarize\|translate`） |

---

## 🧪 性能压测与辅助脚本

项目在 `backend/` 目录下提供了测试数据填充与缓存性能对比压测脚本：

1. **批量插入 1000 条测试文章**：
   ```bash
   cd backend
   go run artical_insert.go
   ```
2. **对比开启/关闭 Redis 缓存的并发性能**（需先启动服务或去除对应路由鉴权后测试）：
   ```bash
   cd backend
   go run benchmark.go
   ```
