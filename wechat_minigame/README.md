# fpxxl 微信小游戏原生版

这是从 Cocos 工程资源迁出的微信小游戏原生实现。入口是 `game.js`，界面和玩法代码在 `src/`，图片和音频在 `assets/`。

## 如何运行

1. 打开微信开发者工具。
2. 选择“导入项目”。
3. 项目目录选择本目录：`wechat_minigame/`。
4. AppID 可先使用测试号或替换 `project.config.json` 里的 `appid`。
5. 编译运行。

## 当前已实现

- 首页：继续当前关卡、进入选关。
- 选关：20 关展示、锁定状态、星级展示。
- 游戏关卡：翻牌匹配、开局预览、倒计时、错误次数限制、提示、通关/失败弹窗。
- 本地存档：当前关卡、金币、提示次数、关卡星级。
- 音效和背景音乐：复用 Cocos 工程中的 mp3 资源。

## 目录说明

```text
game.js                 微信小游戏入口
game.json               小游戏运行配置
project.config.json     微信开发者工具项目配置
src/app.js              场景切换、触摸处理、主循环
src/game-logic.js       翻牌匹配和关卡规则
src/renderer.js         Canvas 2D 绘制
src/assets.js           图片和音频加载
src/storage.js          微信本地存档
src/config.js           关卡、资源路径配置
assets/                 从 Cocos 工程迁出的图片和音频
```

## 后续建议

- 真机确认背景音乐自动播放策略，必要时改成用户首次点击后播放。
- 接入正式 AppID 后再补广告、分享、排行榜等微信能力。
- 如果美术资源继续调整，只需要替换 `assets/` 下同名文件。
