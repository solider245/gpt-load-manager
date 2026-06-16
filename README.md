# GPT-Load Manager

> 预制菜式一键部署 GPT-Load 集群——通过浏览器管理所有节点。

```bash
# 启动面板
./gpt-load-manager

# 打开浏览器 → 添加服务器 → 一键部署 → 完成
```

---

## 这是什么

GPT-Load Manager 是一个独立的 Web 面板，通过 **SSH 部署** + **HTTP 监控** 来管理 GPT-Load 集群。

**不碰 gpt-load 一行代码。** 原版 gpt-load 怎么跑，面板部署完还是怎么跑。拆了面板，集群照常工作。

## 核心理念

| 理念 | 说明 |
|------|------|
| **预制菜** | 面板内置 gpt-load 二进制 + 配置模板，SSH 过去解压即用 |
| **零侵入** | 只通过 SSH 传文件 + systemctl，不修改 gpt-load |
| **渐进式** | 先 SSH 部署，再 HTTP 监控，后 Agent 进阶（按需） |
| **单文件** | 前端嵌入后端，一个二进制发布 |

## 快速开始

> 即将推出

## 架构

```
面板 ──SSH──→ 服务器 (部署/升级/重启)
面板 ──HTTP──→ gpt-load API (健康检查/指标采集)
NAT节点 ──curl | bash──→ 自部署 (令牌认证)
```

## 路线图

| 阶段 | 目标 | 时间 |
|------|------|------|
| P0 | 骨架搭建 — 添加服务器，显示健康状态 | 进行中 |
| P1 | 部署能力 — 一键部署 gpt-load | 规划 |
| P2 | 监控看板 — 集群状态可视化 | 规划 |
| P3 | 安装脚本 — NAT 节点 `curl \| bash` | 规划 |
| P4 | 生产加固 — 认证、审计、告警 | 规划 |

详见 [ROADMAP.md](ROADMAP.md)。

## 技术栈

- 后端：Go + Gin + GORM + SQLite
- 前端：Vue 3 + Element Plus + Vite
- SSH：golang.org/x/crypto/ssh + sshpass
- 发布：单二进制 (go:embed)

## License

MIT
