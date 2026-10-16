---
schema-version: 1
knowledge-id: player-responsive-layout-and-follow-20261015
status: active
confidence: verified
owner: video-english/frontend
verified-at: 2026-10-15T05:07:27.902446+00:00
freshness-ttl-seconds: 2592000
source-sha256: 9c68def4f4e1ad5a7753e34379ffb6f9854287adf61e224740ea84b5c4a4a8ec
---

# 播放器响应布局与字幕跟随的交互边界

播放器采用受可视高度约束的工作区：竖屏上下排列视频和字幕，宽度至少960px或宽度至少640px的横屏采用双栏。字幕跟随只滚动字幕列表；手动触摸、滚轮或键盘浏览会暂停跟随，按钮可恢复。发音面板为非模态，Escape关闭并归还单词焦点。英文轨切到现有中文轨时交换轨道，字幕查询排除英中相同ID。实体移动设备和iOS Safari兼容性仍待验证。
