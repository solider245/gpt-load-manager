# GPT-Load Manager — 产品路线图

> Ubuntu 上管理一台机器，和 SSH 上去操作没什么区别。
> 管理一百台，就需要面板了。
> **GPT-Load Manager** 就是这个面板。

---

## 一、产品定位

### 一句话

> 一个 Web 面板，让你通过浏览器就能在任意服务器上部署和管理 GPT-Load 集群——SSH 搞定部署，HTTP 搞定监控，不碰 gpt-load 一行代码。

### 核心理念

- **预制菜模式**：面板内置 gpt-load 二进制包 + 配置模板，SSH 过去直接解压、配置、启动
- **零侵入**：gpt-load 原版不动，部署完后就是标准 gpt-load，拆了面板照样跑
- **渐进式**：先用 SSH 解决部署，再用 HTTP 解决监控，最后用 Agent 解决进阶管控

### 解决的问题

| 场景 | 原来 | 现在 |
|------|------|------|
| 部署一个新节点 | SSH 上去手动下载、配 .env、配 systemd | 面板填 IP + 密码/密钥，点一下 |
| 查看集群状态 | 每台机器 curl /health，自己拼 | 面板一个页面全展示 |
| 升级版本 | 每台机器手动替换二进制、重启 | 面板选择版本，批量推送重启 |
| NAT 节点部署 | 无解，得有人去现场 | curl \| bash 一条命令 |

---

## 二、总体架构

```
┌──────────────────────────────────────────────────────────────────┐
│                        GPT-Load Manager                          │
│                                                                  │
│  ┌──────────────┐   ┌────────────────────────────────────────┐  │
│  │  Web UI      │   │  API                                   │  │
│  │  ┌────────┐  │   │  ┌─────────┐  ┌─────────┐  ┌───────┐ │  │
│  │  │节点管理 │  │   │  │ 部署 API │  │ 监控 API │  │系统API│ │  │
│  │  │一键部署 │  │   │  │ /deploy │  │ /monitor│  │/system│ │  │
│  │  │监控看板 │  │   │  └────┬────┘  └────┬────┘  └───┬───┘ │  │
│  │  │版本管理 │  │   │       │           │            │     │  │
│  │  │日志查看 │  │   └───────┼───────────┼────────────┘     │  │
│  │  └────────┘  │           │           │                   │  │
│  └──────┬───────┘           │           │                   │  │
│         │                   │           │                   │  │
│         ▼                   ▼           ▼                   │  │
│  ┌──────────────────────────────────────────────────────┐   │  │
│  │              核心引擎 (Engine Layer)                   │   │  │
│  │                                                       │   │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │   │  │
│  │  │ SSH 连接器    │  │ HTTP 轮询器  │  │ 部署引擎   │ │   │  │
│  │  │ (ssh +       │  │ (gpt-load    │  │ (二进制管理 │ │   │  │
│  │  │  sshpass)    │  │  API 客户端) │  │  模板渲染) │ │   │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬─────┘ │   │  │
│  └─────────┼─────────────────┼──────────────────┘       │   │
│            │                 │                          │   │
│  ┌─────────▼─────────────────▼────────────────────┐     │   │
│  │              存储层                              │     │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐ │     │   │
│  │  │ 节点清单  │  │ 部署记录  │  │ gpt-load     │ │     │   │
│  │  │ (SQLite) │  │ (SQLite) │  │ 二进制包     │ │     │   │
│  │  └──────────┘  └──────────┘  │ (嵌入/缓存)  │ │     │   │
│  │                              └──────────────┘ │     │   │
│  └────────────────────────────────────────────────┘     │   │
└──────────────────────────────────────────────────────────┘   │
                                                               │
         ┌──────────────────────┼────────────────────┐
         ▼                      ▼                    ▼
   ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
   │ 目标服务器 A  │      │ 目标服务器 B  │      │ 目标服务器 C  │
   │ 公网可达      │      │ 公网可达      │      │ NAT 节点     │
   │              │      │              │      │              │
   │ Panel ──SSH──→│      │ Panel ──SSH──→│      │ 本地         │
   │ Panel ◀─HTTP──│      │ Panel ◀─HTTP──│      │ bash 脚本   │
   └──────────────┘      └──────────────┘      │ Slave主动上报 │
                                               └──────────────┘
```

### 组件职责

| 组件 | 职责 | 技术 |
|------|------|------|
| **Web UI** | 服务器管理、部署操作、监控看板 | Vue 3 + Element Plus |
| **API** | 面板自身 REST API | Gin |
| **SSH 连接器** | 连接目标服务器，执行命令/传输文件 | golang.org/x/crypto/ssh + sshpass fallback |
| **HTTP 轮询器** | 定时采集 gpt-load 节点指标 | net/http |
| **部署引擎** | 管理 gpt-load 版本包，生成 .env，触发安装 | Go |
| **存储** | 节点清单、部署历史、配置模板 | SQLite (gorm) |

---

## 三、分步实施计划

### P0：骨架搭建（可审核）

**目标**：项目跑起来，看到页面，能连接一台服务器并展示 /health

| 步骤 | 内容 | 说明 |
|------|------|------|
| 0.1 | Go 后端骨架 | gin 路由、启动入口、配置文件加载 |
| 0.2 | SQLite 存储 | gorm 初始化、服务器清单表 (servers) |
| 0.3 | SSH 连接器 | 通过密码或密钥连接到目标服务器 |
| 0.4 | HTTP 轮询器 | 请求目标服务器的 /health，返回状态 |
| 0.5 | Node API | CRUD 服务器清单、触发健康检查 |
| 0.6 | Vue 3 前端骨架 | 登录页、服务器列表页、服务器详情页 |
| 0.7 | 前后端对接 | 前端调用后端 API，展示服务器状态 |
| 0.8 | 构建+启动 | go:embed 前端、单二进制启动 |

**验证方式**：在面板中添加一台服务器（填写 IP + SSH 凭据），面板能连接上去并显示 `/health` 状态。

### P1：部署能力（核心价值）

**目标**：通过面板一键部署 gpt-load 到目标服务器

| 步骤 | 内容 | 说明 |
|------|------|------|
| 1.1 | gpt-load 版本管理 | 面板编译/下载 gpt-load 二进制，管理版本列表 |
| 1.2 | 配置模板 | .env 模板引擎，支持变量替换（NODE_NAME, AUTH_KEY 等） |
| 1.3 | 部署引擎 | 编排部署流程：传二进制 → 写 .env → 写 systemd → 启动 |
| 1.4 | 部署策略配置 | 选择模式(master/slave/standalone)、填写参数 |
| 1.5 | 部署任务管理 | 异步任务执行 + 实时日志输出（WebSocket/SSE） |
| 1.6 | 一键部署 UI | 部署表单 + 执行日志 + 结果展示 |

**验证方式**：在面板中选择一台新服务器，选择模式为 slave，填写 Master 地址，点击部署。观察面板 SSH 过去自动完成安装，节点启动后注册到集群。

### P2：监控与看板（看得见）

**目标**：集群状态一目了然

| 步骤 | 内容 | 说明 |
|------|------|------|
| 2.1 | 节点指标采集 | 定时轮询各节点 API (dashboard/stats, health) |
| 2.2 | 数据显示 | 节点在线/离线、请求量、错误率、RPM |
| 2.3 | 集群总览页面 | 所有节点状态卡片 + 聚合指标 |
| 2.4 | 节点详情页面 | 单节点时间线、请求趋势图 |
| 2.5 | 日志查看 | 远程拉取节点日志文件 |

**验证方式**：部署 2-3 个节点后，面板自动展示集群状态，能看到每个节点的实时指标。

### P3：安装脚本（NAT 节点接入）

**目标**：NAT 后面的节点也能通过 `curl | bash` 部署

| 步骤 | 内容 | 说明 |
|------|------|------|
| 3.1 | 部署令牌生成 | 限时令牌 API，面板生成一次有效令牌 |
| 3.2 | /install 端点 | Master 侧动态生成 install.sh，嵌入令牌验证 |
| 3.3 | /download 端点 | 提供二进制下载 |
| 3.4 | install.sh 脚本 | 零依赖安装脚本（不含 jq，纯 bash） |
| 3.5 | 部署 UI 集成 | 面板显示部署命令，复制即用 |

**验证方式**：在 NAT 服务器上运行 `curl -sSL https://PANEL:PORT/install | sudo bash -s -- <token>`，自动安装并注册到集群。

### P4：生产加固

**目标**：可放心用于生产环境

| 步骤 | 内容 | 说明 |
|------|------|------|
| 4.1 | 面板自身认证 | 用户名密码登录 + Session/JWT |
| 4.2 | 操作审计日志 | SSH 执行记录、部署历史可追溯 |
| 4.3 | 版本升级管理 | 批量推送新版本，滚动升级 |
| 4.4 | 异常告警 | 节点离线通知、错误率过高告警 |
| 4.5 | HTTPS | 面板自身支持 TLS |

---

## 四、测试策略

### 测试层级

```
┌──────────────────────────────┐
│  E2E 测试（手动 + 自动化）    │
│  在真实服务器上部署验证       │
├──────────────────────────────┤
│  集成测试                    │
│  SSH 连接测试（本地 SSH 模拟）│
│  API 集成测试                 │
├──────────────────────────────┤
│  单元测试                    │
│  部署引擎逻辑、模板渲染、     │
│  配置解析                    │
└──────────────────────────────┘
```

### 各阶段测试方法

#### P0 骨架阶段

| 测试内容 | 方式 |
|---------|------|
| SSH 连接 | 连接本地 localhost SSH server |
| /health 轮询 | mock HTTP server 返回模拟数据 |
| API CRUD | go test + SQLite 内存模式 |
| 前端展示 | Vue 组件单元测试，mock API |

#### P1 部署阶段

| 测试内容 | 方式 |
|---------|------|
| 二进制传输 | scp 到本地临时目录验证 |
| .env 渲染 | 模板 + 变量 → 验证输出内容 |
| systemd 安装 | 容器内测试（避免影响宿主机） |
| 完整部署流程 | DigitalOcean/Linode API 创建临时机器 → 部署 → 验证 → 销毁 |

#### P2 监控阶段

| 测试内容 | 方式 |
|---------|------|
| 指标采集 | 启动一个本地 gpt-load 实例，验证轮询器能正确解析响应 |
| 数据聚合 | 多个 mock 节点，验证聚合计算正确性 |
| 趋势图表 | 构造历史数据，验证前端图表渲染 |

#### P3 安装脚本

| 测试内容 | 方式 |
|---------|------|
| install.sh 兼容性 | Docker 容器测试（Ubuntu 20.04/22.04/24.04, Debian 11/12, CentOS 7/8） |
| 令牌机制 | 验证令牌生成 → 验证 → 使用 → 失效 完整周期 |
| NAT 场景 | 无公网 IP 的 Docker 容器模拟 |

### 持续验证方法

```
每个 PR 自动执行:
├── go test ./...        # 后端单元测试
├── npm run test:unit    # 前端组件测试
├── npm run build        # 前端构建验证
├── go build ./...       # 后端编译验证
└── golangci-lint run    # 代码规范检查
```

### 真实环境验证周期

| 阶段 | 验证环境 |
|------|---------|
| 开发期 | 本地 Docker Compose（1 panel + 2 mock nodes） |
| PR 合并前 | 开发者的测试服务器 |
| 版本发布前 | 完整的 3 节点真实集群（master + 2 slaves） |

---

## 五、技术选型

| 维度 | 选择 | 原因 |
|------|------|------|
| 后端语言 | Go | 与 gpt-load 一致，单二进制，交叉编译方便 |
| Web 框架 | Gin | 用户熟悉，与 gpt-load 一致 |
| SSH | golang.org/x/crypto/ssh | 标准库，纯 Go，无 CGO |
| 密码认证 | sshpass（外部调用） | 避免原生 Go SSH 库对密码交互支持弱的问题 |
| 数据库 | SQLite (gorm) | 面板自身数据简单，无需单独部署数据库 |
| 前端框架 | Vue 3 | 与 gpt-load 一致 |
| UI 组件 | Element Plus | 成熟、文档好 |
| 前端构建 | Vite | 快，HMR 体验好 |
| 嵌入式 | go:embed | 前端构建产物嵌入 Go 二进制，单文件发布 |
| 任务队列 | 内存 + goroutine | 初期部署任务量小，无需独立消息队列 |

---

## 六、项目结构

```
gpt-load-manager/
├── main.go                      # 入口
├── internal/
│   ├── api/
│   │   ├── router.go            # 路由注册
│   │   ├── server.go            # 服务器管理 API
│   │   ├── deploy.go            # 部署 API
│   │   └── monitor.go           # 监控 API
│   ├── db/
│   │   ├── db.go                # 数据库初始化
│   │   └── models.go            # 数据模型
│   ├── deploy/
│   │   ├── engine.go            # 部署引擎
│   │   ├── template.go          # .env 模板渲染
│   │   └── task.go              # 部署任务管理
│   ├── monitor/
│   │   └── poller.go            # HTTP 轮询器
│   ├── server/
│   │   ├── connector.go         # SSH 连接器
│   │   └── file.go              # 文件传输
│   ├── config/
│   │   └── config.go            # 面板配置
│   └── panel/
│       └── handler.go           # 面板自身管理
├── pkg/
│   └── gpt-load/                # gpt-load 版本包（空，放 .gitkeep）
├── scripts/
│   └── install.sh               # 安装脚本模板
├── web/                         # Vue 3 前端
│   ├── src/
│   │   ├── views/
│   │   │   ├── Servers.vue      # 服务器列表
│   │   │   ├── ServerDetail.vue # 服务器详情
│   │   │   ├── Deploy.vue       # 一键部署
│   │   │   ├── Dashboard.vue    # 监控看板
│   │   │   └── Settings.vue     # 面板设置
│   │   ├── api/                 # API 调用
│   │   └── components/          # 通用组件
│   └── ...
├── README.md
└── ROADMAP.md                   # 本文件
```

---

## 七、数据模型

```sql
-- 服务器清单
CREATE TABLE servers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL,             -- 显示名称
    host            TEXT NOT NULL,             -- IP 或域名
    port            INTEGER DEFAULT 22,
    auth_type       TEXT NOT NULL DEFAULT 'password',  -- password | key
    auth_credential TEXT NOT NULL,             -- 密码或密钥路径/内容
    gpt_mode        TEXT DEFAULT 'standalone', -- standalone | master | slave
    gpt_port        INTEGER DEFAULT 3001,      -- gpt-load 端口
    gpt_version     TEXT,                      -- 部署的 gpt-load 版本
    status          TEXT DEFAULT 'unknown',    -- online | offline | unknown
    last_health_at  DATETIME,                  -- 最后一次健康检查时间
    created_at      DATETIME DEFAULT NOW(),
    updated_at      DATETIME DEFAULT NOW()
);

-- 部署记录
CREATE TABLE deploy_logs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id       INTEGER REFERENCES servers(id),
    gpt_version     TEXT NOT NULL,
    action          TEXT NOT NULL,             -- install | upgrade | restart | rollback
    status          TEXT NOT NULL DEFAULT 'pending', -- pending | running | success | failed
    log             TEXT,                      -- 部署日志
    created_at      DATETIME DEFAULT NOW()
);

-- 监控快照（历史）
CREATE TABLE monitor_snapshots (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id       INTEGER REFERENCES servers(id),
    status          TEXT NOT NULL,             -- online | offline
    uptime          TEXT,
    response_time_ms INTEGER,
    created_at      DATETIME DEFAULT NOW()
);
```

---

## 八、边界与约束

**明确不做：**

1. **不改 gpt-load 代码** — gpt-load 原版什么样，部署完还是什么样。面板只是 SSH + HTTP 的工具
2. **不做端口映射/反向代理** — 那是 FRP/ngrok 的事，不是面板的职责
3. **不替代监控系统** — 面板展示状态，但不做告警通知、PagerDuty 集成（P4 可加简单通知）
4. **不做配置中心** — 不改 gpt-load 配置，只做批量推送 .env

**明确要做：**

1. **SSH 连接是第一优先级** — 连不上就什么都不用谈
2. **部署流程要可靠** — 每一步有日志，失败能回滚
3. **监控信息来自 gpt-load 自身 API** — 不额外开端口，不额外装软件
4. **单二进制发布** — 前端嵌入后端，一个文件搞定
