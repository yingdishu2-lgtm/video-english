---
name: "English Media Learner"
description: "以真实影片画面和双语字幕为核心的午夜蓝本地影音学习界面。"
colors:
  canvas: "#0b1019"
  surface: "#141c29"
  surface-raised: "#1b2535"
  divider: "#2a3445"
  text-primary: "#f1f4fa"
  text-muted: "#a3afc2"
  accent: "#7393ff"
  action: "#3d63e4"
  action-hover: "#4265d6"
  selection: "#3857a0"
  success-surface: "#163529"
  success-text: "#d8f7e7"
  error-surface: "#482324"
  error-text: "#ffdeda"
typography:
  display:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif"
    fontSize: "clamp(2rem, 4.4vw, 4.2rem)"
    fontWeight: 760
    lineHeight: 1.02
    letterSpacing: "-0.035em"
  headline:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif"
    fontSize: "clamp(2rem, 4vw, 3.5rem)"
    fontWeight: 760
    lineHeight: 0.98
    letterSpacing: "-0.035em"
  title:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif"
    fontSize: "18px"
    fontWeight: 650
    lineHeight: 1.3
  body:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif"
    fontSize: "16px"
    fontWeight: 400
    lineHeight: 1.65
  label:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif"
    fontSize: "14px"
    fontWeight: 600
    lineHeight: 1.5
  phonetic:
    fontFamily: "'Times New Roman', serif"
    fontSize: "22px"
    fontWeight: 400
    lineHeight: 1.25
rounded:
  word: "3px"
  compact: "4px"
  badge: "6px"
  control: "8px"
  button: "10px"
  card: "12px"
  empty-icon: "14px"
  panel: "16px"
  circular: "50%"
spacing:
  xxs: "4px"
  xs: "8px"
  sm: "10px"
  md: "12px"
  lg: "16px"
  xl: "20px"
  2xl: "24px"
  3xl: "28px"
  4xl: "32px"
  touch: "44px"
components:
  button-primary:
    backgroundColor: "{colors.action}"
    textColor: "#ffffff"
    typography: "{typography.label}"
    rounded: "{rounded.button}"
    padding: "10px 18px"
    height: "44px"
  button-primary-hover:
    backgroundColor: "{colors.action-hover}"
    textColor: "#ffffff"
    typography: "{typography.label}"
    rounded: "{rounded.button}"
    padding: "10px 18px"
    height: "44px"
  button-secondary:
    backgroundColor: "#1b283d"
    textColor: "#dce5f5"
    typography: "{typography.label}"
    rounded: "{rounded.button}"
    padding: "10px 18px"
    height: "44px"
  search-field:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text-primary}"
    typography: "{typography.body}"
    rounded: "{rounded.card}"
    padding: "0 12px"
    height: "46px"
  media-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text-primary}"
    typography: "{typography.body}"
    rounded: "{rounded.card}"
    width: "100%"
  subtitle-active:
    backgroundColor: "#20365b"
    textColor: "#f4f7ff"
    typography: "{typography.body}"
    rounded: "{rounded.button}"
    padding: "10px 7px"
  word-token-touch:
    backgroundColor: "transparent"
    textColor: "{colors.text-primary}"
    typography: "{typography.body}"
    rounded: "{rounded.word}"
    padding: "0 2px"
    height: "44px"
---

# Design System: English Media Learner

## Overview

**Creative North Star: "午夜媒体阅览室"**

界面像一间专注于个人片库的低光阅览室：午夜蓝画布与分层石板蓝面板退到背景，真实影片画面、播放入口和双语字幕承担视觉主角。参考 Infuse 的内容层级与完成度，但图标、布局和交互均由本项目实现，不复制其品牌资产或逐像素构图。

这是一个面向操作的系统。媒体库先帮助用户认出和选择自己的影片；播放器把画面、原生播放控件、字幕时间轴和点词发音组织为连续任务。视觉语言保持克制、清晰和触控友好，状态颜色只表达真实状态。

**Key Characteristics:**

- 午夜蓝画布、冷白正文、石板蓝分层面板与单一实蓝行动色。
- 真实本地媒体抽帧保持原色；不可用时显示明确的无图回退。
- 平台无衬线字体、紧凑信息密度和统一的线性 SVG 图标。
- 手机与平板优先，主要操作和触屏字幕目标保持至少 44px 高。
- 影片画面、字幕与当前学习语境优先于路径和技术元数据。

## Colors

色彩以低光可读性为中心：深蓝黑承担背景，递进的石板蓝划分层级，冷白与灰蓝形成正文层次，饱和蓝只用于行动、焦点和当前状态。

### Primary

- **实蓝行动色**（`action`）：用于主播放入口和明确的主要操作；悬停使用单独的 `action-hover` 状态。
- **冷蓝强调色**（`accent`）：用于焦点、输入光标、加载指示与受控的交互强调，不与行动色混用为装饰。

### Neutral

- **午夜画布**（`canvas`）：全局背景。
- **石板面板**（`surface`）：搜索框、媒体卡片与空状态的基础表面。
- **抬升石板**（`surface-raised`）：图标容器与局部强调表面。
- **克制分隔线**（`divider`）：面板、导航和播放器区域之间的边界。
- **冷白正文**（`text-primary`）：主要标题、正文和可操作内容。
- **灰蓝次文**（`text-muted`）：说明、元数据和次要状态。

### Secondary

- **成功状态组**（`success-surface` / `success-text`）：扫描成功提示。
- **错误状态组**（`error-surface` / `error-text`）：扫描和读取失败提示。

### Named Rules

**The One Action Blue Rule.** 实蓝只标识主要操作；焦点、选中字幕和信息层级使用各自的蓝色表面，不把整屏染成强调色。

**The Honest Image Rule.** 影片画面只来自用户本地媒体的运行时抽帧；封面不可用时使用明确回退，不虚构海报、评分或观看进度。

## Typography

**Display Font:** 平台无衬线字体栈（系统 UI 字体优先）\
**Body Font:** 平台无衬线字体栈（系统 UI 字体优先）\
**Label/Mono Font:** 标签沿用平台无衬线；时长与时间戳启用等宽数字\

**Character:** 标题以紧凑字距和较高字重提供影音产品的自信感，正文保持朴素清晰。IPA 独立使用衬线字体，以便与界面标签区分。

### Hierarchy

- **Display**（760，流体字号，1.02）：媒体库特选影片标题，允许平衡换行。
- **Headline**（760，流体字号，0.98）：页面级媒体库标题。
- **Title**（650，18px，1.3）：字幕面板标题与同级区块标题。
- **Body**（400，16px，1.65）：字幕正文、词义语境和可读状态说明。
- **Label**（600，14px，1.5）：按钮、导航与紧凑控制标签。
- **Phonetic**（400，22px，1.25）：发音面板中的 IPA；不扩展为通用装饰字体。

**The Content Before Metadata Rule.** 影片标题、画面和台词保持最高可读层级；路径、编码和来源始终作为可截断或次要信息。

## Layout

页面以移动优先的单列结构开始。媒体库内容区使用 24px 页边距和 28px 主节奏；在不超过 640px 时收紧为 22px，并将标题、扫描操作和工具栏纵向排列。影片网格默认三列，在 980px 以下变两列，在 640px 以下变单列；480–720px 的中等宽度恢复两列以充分利用小平板和宽手机空间。媒体缩略图保持 16:9。

媒体库在 1100px 以上切换为 220px 固定侧栏与弹性内容区。播放器在竖向视口中采用画面在上、字幕在下的受限高度工作区；在 960px 以上，或 640px 以上的横屏中，切换为约 1.65:1 的画面/字幕双栏。高度不超过 650px 的竖屏进一步压缩画面高度。所有主要容器尊重安全区内边距，并使用 `svh` / `dvh` 约束避免移动地址栏导致溢出。

**The Viewport-Bound Player Rule.** 播放器工作区必须留在当前动态视口内；视频和字幕各自管理尺寸与滚动，页面不依赖无限纵向增长完成核心任务。

## Elevation & Depth

系统以色调分层和细分隔线为主，阴影只出现在需要从画布中抬升的特选画面、媒体缩略图、轨道设置浮层和发音面板。阴影均为无色黑色环境阴影，不使用霓虹光晕；搜索焦点的蓝色外环表达交互焦点，不表达空间高度。

### Shadow Vocabulary

- **卡片低层**：媒体缩略图使用紧凑的柔和阴影，保持图像边缘可辨。
- **特选中层**：大幅特选画面使用更宽的环境阴影，形成首要内容层。
- **浮层高层**：轨道设置与发音面板使用明显但无色的深阴影，保证覆盖关系清楚。
- **焦点环**：输入与控件使用 2px 轮廓或 3px 半透明强调环，并保留外偏移。

**The Tonal First Rule.** 默认表面依靠颜色和边界建立层级；只有真实浮层与主视觉图像获得阴影。

## Shapes

形状语言从紧凑控制到大型容器逐级放大：词按钮仅作轻微圆角，时间徽标使用小圆角，选择框与跟随按钮为 8px，常规按钮与字幕行是 10px，搜索框和媒体卡片为 12px，大型面板与发音卡片为 16px。播放覆盖按钮和加载指示器保持圆形。边框为 1px 克制分隔线；媒体图像与容器共同裁切，不额外套装饰框。

**The Radius by Scale Rule.** 圆角随组件体量递增，控制件不使用夸张胶囊形；圆形仅保留给播放和加载这类具明确语义的元素。

## Components

### Buttons

- **Shape:** 常规按钮使用 10px 圆角与至少 44px 高的触控尺寸。
- **Primary:** 白字配实蓝行动色，水平内边距 18px；仅用于主要行动，如“开始播放”。
- **Secondary:** 冷白字、深石板蓝底和可见边框，用于扫描、重试与清除等次级操作。
- **Hover / Focus:** 仅在精细指针设备启用悬停变色；键盘焦点始终用清晰的蓝色轮廓，减少动态偏好会把过渡压缩到近乎即时。

### Cards / Containers

- **Media Card:** 16:9 真实抽帧、12px 裁切圆角、下方标题与真实分辨率/音轨/字幕数量；桌面悬停时图像轻微缩放并显示播放按钮，触屏直接显示播放按钮。
- **Featured Media:** 大型真实抽帧配方向自适应暗色渐变，标题与播放入口置于可读区域；手机改用自下而上的遮罩。
- **State Container:** 加载、空、无结果和读取失败复用石板表面与居中文案，但各自保留真实状态和可执行下一步。

### Inputs / Fields

- **Search:** 12px 圆角、46px 总高度、44px 输入与清除目标；聚焦时边框切换为强调色并出现 3px 半透明焦点环。
- **Select:** 8px 圆角、至少 44px 高；播放器倍速和字幕轨道均保留原生选择行为。
- **Error / Disabled:** 错误使用专用红褐表面和浅红文字；禁用控件降低透明度并保持可辨文字。

### Navigation

移动与平板使用 76px 顶栏，桌面媒体库在 1100px 起切换为 220px 左侧栏。活动导航使用较亮蓝灰文字与深蓝选中面；品牌标记和所有功能图标均为圆角端点、1.75px 描边的内联 SVG，不使用字符或表情符号代替图标。

### Subtitle Timeline

当前字幕行使用蓝色底面和可见边框，时间戳独立为至少 44px 的跳转按钮。触屏词按钮保持自然横向宽度和 44px 纵向命中区；精细指针设备可将词按钮恢复为紧凑行高。中文台词按钮同样保持至少 44px 高，并可直接跳转播放。

### Pronunciation Panel

发音面板是非模态浮层：手机贴近安全区边缘，宽屏固定为不超过 440px 的右下浮层。关闭按钮为 44px 方形；单词、IPA、音频、查询状态、双语语境和来源按真实数据逐层出现。

## Do's and Don'ts

### Do:

- **Do** 让真实影片画面、片名、播放入口和双语字幕承担首要层级。
- **Do** 为主要按钮、关闭、搜索清除、轨道选择、时间戳和触屏词按钮保留至少 44px 的可触达高度。
- **Do** 在触屏上保持短词自然横向宽度，只扩大纵向命中区以维持正常词间距。
- **Do** 使用统一的线性 SVG 图标、清晰焦点、真实加载/空/错误状态和 `prefers-reduced-motion` 回退。
- **Do** 让手机、平板和横屏布局都尊重安全区与动态视口。

### Don't:

- **Don't** 虚构海报、评分、观看进度、学习数据或尚未实现的功能。
- **Don't** 把路径、编码和技术标签提升到影片与字幕之上。
- **Don't** 使用霓虹光晕、装饰性渐变文字、字符图标或系统外的展示字体建立视觉噪声。
- **Don't** 强制短词获得 44px 横向宽度，或让自动字幕跟随覆盖用户正在手动浏览的位置。
- **Don't** 把 Chromium 模拟结果写成实体设备或 iOS Safari 已验证的承诺。
