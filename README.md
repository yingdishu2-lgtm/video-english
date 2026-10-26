# English Media Learner

本地优先的影视英语学习播放器。当前版本已完成规格中的 Phase 1～7，并提前加入 Phase 9 的单词切分与音标查看能力。

## 已实现

- React、TypeScript、Vite、TanStack Query、Zustand 前端骨架。
- Go、Gin、Ent、SQLite 后端骨架。
- 扫描 `MEDIA_ROOT` 下的 MP4、M4V、MOV、MKV、WebM。
- 使用 `ffprobe` 保存容器、时长、分辨率及音频/视频/字幕轨信息。
- 基于相对路径、文件大小和修改时间跳过未变化媒体。
- 从真实本地视频生成并缓存封面，媒体变化后刷新；媒体库支持标题、文件与编码搜索。
- MP4/M4V/MOV + H.264 + AAC/MP3 Direct Play。
- MKV + H.264 按所选音轨无损 remux 为 MP4，并缓存到 `data/cache/media/{id}/`。
- 播放接口支持 `audioStream` 参数；非 AAC/MP3 音轨会回退转码为 AAC。
- 使用 HTTP Range 返回媒体，支持浏览器 seek。
- 识别内嵌及外挂 SRT、ASS、SSA、WebVTT 文本字幕。
- FFmpeg 提取内嵌字幕并标准化为毫秒级 `SubtitleCue` 时间轴。
- 播放器字幕轨选择、当前句高亮、自动滚动和点击跳转。
- 英文主时间轴 + 中文字幕的时间重叠/中心距离对齐，支持 1:1、1:N、N:1 并缓存为 `SubtitleAlignment`。
- 双语学习字幕行；英文单词可点击，并显示当前双语语境。
- 内置 CMU Pronouncing Dictionary，常见英文词断网时也可显示美式 IPA；未收录词可由在线词典接口补充并缓存。
- 媒体路径清理、目录边界和符号链接边界检查。

当前前端还没有音轨切换控件，但后端播放接口已经支持选择音轨。下一阶段是单句播放、上一句/下一句、循环当前句与字幕 offset。

说明：IPA 是发音标注，能辅助确认读音；自然拼读侧重字母与发音规律，两者相关但不等同。本项目展示的是词典发音，不会按拼写臆造读音。离线词典数据来自 CMU Pronouncing Dictionary，因此默认反映美式发音。

## 环境要求

- Go 1.25 或更高版本。
- Node.js 22 或更高版本。
- FFmpeg 与 ffprobe。

本机若使用 Go 1.27，Ent 代码生成请通过 `make generate` 使用声明的兼容工具链；正常构建和运行不受影响。

## 本地运行

最简单的方式是构建前端并由 Go 服务统一托管：

```bash
make run
```

然后打开 <http://localhost:8080>。不要直接双击 `frontend/index.html`；Vite 应用不能通过 `file://` 协议运行。

首次运行前需要安装前端依赖：

   ```bash
   cd frontend
   npm install
   ```

将视频文件放入项目的 `media/` 目录。

需要前端热更新时，可改用双进程开发模式：

1. 在一个终端启动后端：

   ```bash
   make backend
   ```

2. 在另一个终端启动前端：

   ```bash
   make frontend
   ```

3. 打开 <http://localhost:5173>，点击“扫描媒体目录”。

默认设置见 [.env.example](./.env.example)。后端从进程环境读取配置；从 `backend/` 启动时，默认媒体目录为项目根目录的 `media/`，数据目录为 `data/`。

## API

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/health` | 健康检查 |
| `POST` | `/api/v1/media/scan` | 扫描并探测媒体 |
| `GET` | `/api/v1/media` | 媒体列表与轨道 |
| `GET` | `/api/v1/media/:id` | 媒体详情 |
| `GET` | `/api/v1/media/:id/poster` | 本地视频抽帧封面，支持缓存与 ETag |
| `POST` | `/api/v1/media/:id/probe` | 重新运行 ffprobe |
| `GET` | `/api/v1/media/:id/playback` | 获取直放信息 |
| `GET` | `/api/v1/media/:id/stream` | Range 媒体流 |
| `GET` | `/api/v1/media/:id/subtitles` | 字幕轨列表 |
| `POST` | `/api/v1/subtitle-tracks/:id/process` | 提取并解析字幕轨 |
| `GET` | `/api/v1/subtitle-tracks/:id/cues` | 标准化字幕时间轴 |
| `GET` | `/api/v1/media/:id/learning-cues` | 按 `englishTrackId`、`chineseTrackId` 返回双语学习时间轴 |
| `GET` | `/api/v1/dictionary/:word` | 查询并缓存在线 IPA/发音补充信息 |

## 验证

```bash
make test
make build
```

后端测试覆盖扫描、播放决策、HTTP Range、双语对齐和音标缓存；前端测试覆盖播放器状态、英文词元切分与 ARPABET → IPA 转换，构建命令同时执行 TypeScript 检查。

2026 年 10 月前端重做与字幕居中跟随的历史截图、视口测量和验证边界见[界面验收资料](./docs/qa/frontend-redesign-2026-10/README.md)。
