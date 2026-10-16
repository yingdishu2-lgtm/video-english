---
schema-version: 1
knowledge-id: local-media-poster-api-20261015
status: active
confidence: verified
owner: video-english/backend
verified-at: 2026-10-15T05:53:10.206059+00:00
freshness-ttl-seconds: 2592000
source-sha256: dc1313e403044796604f3704c938261d39be6eb2c0bcd79a46f8a39d18e89893
---

# 媒体库真实封面接口与缓存失效约定

GET /api/v1/media/:id/poster 从已验证边界内的本地视频生成640×360 JPEG；复用FFmpeg并限制单次抽帧为15秒，请求取消向子进程传播。缓存键绑定媒体文件大小、mtime和输出尺寸，响应含ETag与private max-age=86400。前端封面URL带媒体updatedAt版本参数，使扫描更新后浏览器重新取图。无图或提取失败由前端明确回退，不调用外部电影海报服务。
