---
version: 1
slug: "frontend-src-features-library-librarypage-tsx"
primary_target: "frontend/src/features/library/LibraryPage.tsx"
related_targets: ["frontend/src/features/player/PlayerPage.tsx","frontend/src/components/AppShell.tsx","frontend/src/styles.css"]
---

# 媒体库与播放器重做

Mode: Operate
Scope: 媒体库、应用导航、播放器及发音面板；保留真实媒体、扫描、播放、字幕和点词能力，以及已完成的移动适配。

用户已确认按成熟视频产品的完成度制作，具体方案委托执行者选择。因此本轮直接执行选定方向，不再请求用户选择设计卡片。主要参考 https://firecore.com/infuse 的内容层级与完成度，不复制其品牌资产或源码。

## Direction contract

THESIS: 让真实影片画面、清晰的播放入口和可阅读的双语字幕构成产品，而不是让路径、技术标签与空卡片主导首屏。

OWN-WORLD: 午夜蓝底色、冷白正文、分层石板蓝面板、实蓝行动按钮，真实抽帧保持原色。平台无衬线字体，统一线性 SVG 图标，12px 控件圆角和克制分隔线。色彩区分导航、选中字幕与交互状态，无霓虹光晕。

STORY: 用户进入媒体库即可认出自己的影片，搜索或扫描后选择播放；播放器让画面与双语学习并列，点词不打断播放，手动浏览字幕不被拉回。

FIRST VIEWPORT: 桌面媒体库左侧约220px导航，右侧标题与扫描操作、真实影片横幅及16:9封面网格；手机为紧凑顶栏与单列内容。播放器沿用受视口约束的竖屏上下、宽横屏双栏工作区。标志交互是影片封面上的明确播放入口，以及字幕当前句在蓝色底面中保持可见；手机所有主操作可触达。

FORM: 选定成熟影音库的界面语言，以 Infuse 的画面优先与密度为制作标尺；自有候选池第3项，seed 58874a8b。其余文化参考包括 Criterion 影像收藏、电影节导视、双语剧本、影带租赁店、电视字幕规范与影院排期。抽签挑战者均未在受众辨识和产品清晰度两项上胜出：数据艺术/动物园地图/机票/星图/设计年鉴/短视频流分别贡献对齐、状态色纪律、完整状态、低光可读性、间距一致性与内容优先的纪律，不搬用装饰或导航模型。

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance

## Constraints

不虚构海报、评分、观看进度、账号、学习数据或尚未实现的生词本/复习能力。封面来自本地媒体抽帧，失败时提供明确的无图回退。保留键盘操作、44px主控件、行内单词至少44px纵向命中区、清晰焦点和减少动态效果；行内短词不强制扩宽，以保持自然词间距。真实移动设备与iOS Safari兼容性不由Chromium模拟替代。
