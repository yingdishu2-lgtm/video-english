---
schema-version: 1
knowledge-id: local-start-make-run
status: active
confidence: verified
owner: project
verified-at: 2026-10-04T12:15:08.238437+00:00
freshness-ttl-seconds: 2592000
source-sha256: cb6d0bfacbbd87d64796153671c15bb6f8e2db3f10f6f7201a75efdf0e719060
---

# 本地统一服务启动方式

在项目根目录执行 make run：先构建 frontend，再使用 GOTOOLCHAIN=go1.25.0 启动 backend/cmd/server。统一入口 http://localhost:8080，健康检查 /health。首次使用需安装 frontend 的 npm 依赖。
