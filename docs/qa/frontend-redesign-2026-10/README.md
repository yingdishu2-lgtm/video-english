# 2026 年 10 月界面验收资料

本目录归档本地已有的前端重做与字幕居中跟随资料，对应实现已包含在提交 `c090f22699c4038c90bda947cafc9d61a3c4103d` 中。2026-10-26 整理时保留了原始文件内容，并记录在 [SHA256SUMS](./SHA256SUMS) 中。

这些资料用于回看当时的界面和测量结果。它们不代表对当前环境重新执行了全部浏览器验收；原始浏览器采集脚本未保存在资料包内。

## 验收记录

| 资料 | 内容与边界 |
| --- | --- |
| [前端重做验收说明](./summary.md) | 2026-10-15 的历史记录，包含当时报告的测试、交互检查和未验证事项 |
| [响应式视口测量](./qa-results.json) | 6 种 Chrome/Chromium 模拟视口中的媒体库、播放器、发音面板尺寸及页面错误记录 |
| [字幕居中测量](./subtitle-centering/qa-results.json) | 5 种视口中首句、换句、末句、尺寸变化和恢复跟随共 35 个测量样本 |
| [列表宽度变化测量](./subtitle-centering/width-resize-results.json) | 同一字幕在列表宽度变化前后的居中偏差 |

原始数据均标明未使用实体设备。实体手机、平板、iOS Safari、其他浏览器和受限网络表现仍需单独验证。`summary.md` 中的通过结论和测试数量属于当时记录。

## 截图索引

| 场景 | 截图 |
| --- | --- |
| 重做前媒体库 | [桌面](./before-library-desktop.png)、[手机](./before-library-mobile.png) |
| 媒体库 | [桌面](./library-desktop.png)、[手机](./library-mobile.png)、[小屏手机](./library-small-mobile.png)、[手机横屏](./library-phone-landscape.png)、[平板竖屏](./library-tablet-portrait.png)、[平板横屏](./library-tablet-landscape.png) |
| 媒体库状态 | [空态](./library-empty.png)、[错误态](./library-error.png) |
| 播放器 | [桌面](./player-desktop.png)、[手机](./player-mobile.png)、[小屏手机](./player-small-mobile.png)、[手机横屏](./player-phone-landscape.png)、[平板竖屏](./player-tablet-portrait.png)、[平板横屏](./player-tablet-landscape.png) |
| 点词发音 | [桌面](./word-desktop.png)、[手机](./word-mobile.png) |
| 字幕居中 | [桌面](./subtitle-centering/desktop.png)、[手机](./subtitle-centering/phone.png)、[手机横屏](./subtitle-centering/phone-landscape.png)、[平板竖屏](./subtitle-centering/tablet-portrait.png)、[平板横屏](./subtitle-centering/tablet-landscape.png) |
| 末句居中 | [桌面](./subtitle-centering/desktop-last.png)、[手机](./subtitle-centering/phone-last.png)、[手机横屏](./subtitle-centering/phone-landscape-last.png)、[平板竖屏](./subtitle-centering/tablet-portrait-last.png)、[平板横屏](./subtitle-centering/tablet-landscape-last.png) |

截图中的视频封面、播放画面和字幕来自当时的本地媒体，作为界面验收示例保存。媒体库空态与错误态使用受控测试响应；本目录不包含视频、完整字幕、数据库或播放缓存。

## 后续复查

在项目根目录运行已有自动化检查：

```bash
make test
make build
```

检查归档文件是否与本次整理时一致：

```bash
cd docs/qa/frontend-redesign-2026-10
shasum -a 256 -c SHA256SUMS
```

重新进行界面验收时，记录新的日期、提交、浏览器和视口范围，并另存结果，保留本目录的历史原件。本机安装、登录和推送回执，以及工具临时状态不属于这份验收资料。
