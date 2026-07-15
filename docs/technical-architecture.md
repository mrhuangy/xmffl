# 微信记忆翻牌消除小游戏技术架构

## 架构目标

- 支持微信小游戏快速开发、调试、提交审核和上线。
- 核心玩法逻辑和 UI 表现解耦，降低状态错乱风险。
- 关卡、主题、奖励和广告配置可扩展。
- 广告逻辑集中封装，便于控制频率、奖励和数据上报。
- 第一版可单机运行，后续平滑扩展到云端配置、排行榜和运营分析。

## 技术栈推荐

| 层级 | 推荐方案 | 说明 |
| --- | --- | --- |
| 游戏引擎 | Cocos Creator 3.x | 适合 2D 微信小游戏，组件化成熟 |
| 语言 | TypeScript | 类型明确，便于维护状态机和配置 |
| 平台 | 微信小游戏 | 目标发布平台 |
| 包管理 | npm | 管理构建工具和辅助依赖 |
| 本地存储 | 微信本地存储 API | 保存进度、设置、金币和道具 |
| 云端能力 | 微信云开发优先 | 后续用于远程配置、排行榜和事件数据 |
| 广告 | 微信小游戏广告 API | 激励视频、插屏、Banner/模板广告 |

## 系统总览

```text
微信小游戏运行时
  -> GameScene
  -> CardGrid / Card
  -> MatchSystem
  -> LevelManager
  -> RewardService
  -> AdService
  -> ProgressService
  -> AnalyticsService
  -> ConfigService
```

## 目录建议

```text
assets/
  themes/                 牌面主题资源
src/
  scenes/                 场景脚本
  components/             Card、CardGrid 等组件
  systems/                MatchSystem、LevelManager
  services/               AdService、RewardService、ProgressService
  configs/                本地关卡和广告配置
  models/                 类型定义
  utils/                  随机、时间、存储工具
tests/
  gameplay/               玩法逻辑测试
  fixtures/               固定牌组和关卡配置
docs/
  project-plan.md
  gameplay-design.md
  technical-architecture.md
```

## 前端架构

| 模块 | 职责 |
| --- | --- |
| `HomeScene` | 首页、开始游戏、入口广告展示 |
| `GameScene` | 游戏主场景，协调 UI、牌阵和关卡状态 |
| `ResultPanel` | 展示星级、步数、用时、奖励和广告入口 |
| `CardGrid` | 根据关卡配置生成和布局牌阵 |
| `CardView` | 单张卡牌的显示、翻转、消除动画 |
| `GameHud` | 步数、用时、提示、暂停等 UI |

## 核心系统

| 模块 | 职责 |
| --- | --- |
| `MatchSystem` | 处理翻牌输入、匹配判断、状态流转 |
| `LevelManager` | 读取关卡配置、生成牌组、切换下一关 |
| `RewardService` | 计算星级、金币、道具和广告奖励 |
| `ProgressService` | 本地保存和读取玩家进度 |
| `AdService` | 加载和展示激励视频、插屏、Banner |
| `AnalyticsService` | 记录关卡事件、广告事件、奖励事件 |
| `ConfigService` | 读取本地配置，后续接入远程配置 |

## 游戏状态机

| 状态 | 说明 |
| --- | --- |
| `ready` | 关卡生成完成，等待玩家开始 |
| `idle` | 等待玩家点击 |
| `first_card_opened` | 已翻开第一张牌 |
| `checking_match` | 已翻开两张，等待系统判断 |
| `resolving_match` | 匹配成功，执行消除 |
| `resolving_mismatch` | 匹配失败，延迟盖回 |
| `level_complete` | 全部消除，进入结算 |
| `paused` | 暂停或广告播放中 |

## 数据模型

### CardConfig

```ts
type CardConfig = {
  id: string;
  pairId: string;
  iconKey: string;
};
```

### RuntimeCard

```ts
type RuntimeCard = CardConfig & {
  state: "face_down" | "face_up" | "locked" | "removed";
  index: number;
};
```

### LevelConfig

```ts
type LevelConfig = {
  levelId: number;
  rows: number;
  cols: number;
  pairCount: number;
  mode: "normal" | "time_limit" | "step_limit";
  themeId: string;
  flipBackDelayMs: number;
  excellentStepThreshold: number;
  normalStepThreshold: number;
  excellentTimeThreshold?: number;
  normalTimeThreshold?: number;
  timeLimitSeconds?: number;
  stepLimit?: number;
};
```

### LevelResult

```ts
type LevelResult = {
  levelId: number;
  steps: number;
  elapsedMs: number;
  stars: 1 | 2 | 3;
  coinsEarned: number;
  usedHints: number;
  completedAt: number;
};
```

### PlayerProgress

```ts
type PlayerProgress = {
  currentLevel: number;
  coins: number;
  hints: number;
  levelStars: Record<string, number>;
  completedLevels: number[];
  updatedAt: number;
};
```

### AdEvent

```ts
type AdEvent = {
  eventId: string;
  adType: "rewarded_video" | "interstitial" | "banner";
  placement:
    | "hint"
    | "double_reward"
    | "extra_coins"
    | "add_time"
    | "add_steps"
    | "revive"
    | "between_levels"
    | "home";
  status: "requested" | "shown" | "completed" | "closed" | "failed" | "reward_granted";
  levelId?: number;
  rewardId?: string;
  createdAt: number;
};
```

## 广告服务设计

`AdService` 只负责广告加载、展示、回调和事件记录，不直接发放奖励。奖励由 `RewardService` 根据广告完成结果统一发放。

建议方法：

```ts
interface AdService {
  preloadRewardedVideo(placement: AdPlacement): Promise<void>;
  showRewardedVideo(placement: AdPlacement): Promise<AdShowResult>;
  showInterstitial(placement: AdPlacement): Promise<AdShowResult>;
  showBanner(scene: "home" | "result" | "level_select"): void;
  hideBanner(): void;
}
```

奖励发放流程：

1. UI 请求某广告位奖励。
2. `AdService` 记录 `requested` 并展示广告。
3. 玩家完整观看后返回 `completed`。
4. `RewardService` 校验 `rewardId` 是否已发放。
5. 发放奖励并记录 `reward_granted`。
6. 如果广告关闭或失败，不发奖励，不消耗玩家机会。

## 广告频控配置

```ts
type AdFrequencyConfig = {
  noInterstitialBeforeLevel: number;
  interstitialEveryLevels: number;
  maxInterstitialPerDay: number;
  maxRevivePerLevel: number;
  bannerEnabledScenes: Array<"home" | "result" | "level_select">;
};
```

建议默认值：

```ts
const defaultAdFrequencyConfig: AdFrequencyConfig = {
  noInterstitialBeforeLevel: 4,
  interstitialEveryLevels: 4,
  maxInterstitialPerDay: 10,
  maxRevivePerLevel: 1,
  bannerEnabledScenes: ["home", "result"],
};
```

## 数据库和扩展原则

- 第一期使用本地 JSON 配置关卡，后续迁移到远程配置。
- 玩家进度、关卡结果、广告事件分离存储。
- 奖励来源要可追踪，例如通关、广告、每日奖励、活动。
- 关卡模式使用枚举字段，不用多个布尔字段表达复杂状态。
- 广告频控独立配置，不写死在 UI 组件。
- 主题资源使用 `themeId` 管理，避免关卡逻辑依赖具体图片路径。

## API 草案

第一期可不需要后端 API。后续接入微信云开发或轻量后端时，建议接口如下：

| API | 方法 | 用途 |
| --- | --- | --- |
| `/config/levels` | GET | 获取远程关卡配置 |
| `/config/ads` | GET | 获取广告频控配置 |
| `/player/progress` | GET/POST | 同步玩家进度 |
| `/leaderboard/submit` | POST | 提交排行榜成绩 |
| `/events/batch` | POST | 批量上报关卡、奖励和广告事件 |

## 第三方集成

| 集成 | 阶段 | 说明 |
| --- | --- | --- |
| 微信小游戏 API | 第一期 | 生命周期、存储、震动、分享 |
| 微信广告 API | 第 3 期 | 激励视频优先，插屏后置 |
| 微信云开发 | 第 5 期 | 远程配置、事件上报、排行榜 |
| 数据分析工具 | 第 5 期 | 统计留存、关卡和广告指标 |

## 安全与隐私

- 不在客户端硬编码敏感密钥。
- 广告位 ID 按环境区分测试和正式。
- 数据上报避免收集不必要的个人敏感信息。
- 存储用户数据前遵守微信小游戏隐私和用户授权要求。
- 奖励发放使用 `rewardId` 去重，避免重复领取。

## 测试策略

| 类型 | 内容 |
| --- | --- |
| 单元测试 | 洗牌、配对、匹配规则、星级计算、奖励发放 |
| 场景测试 | 翻牌、盖回、消除、通关、暂停、重开 |
| 广告测试 | 完成、关闭、失败、重复点击、奖励去重 |
| 存储测试 | 进度保存、版本升级、异常恢复 |
| 真机测试 | 屏幕适配、触摸响应、性能、音效、震动 |

## 部署与环境

| 环境 | 用途 |
| --- | --- |
| 本地开发 | Cocos Creator 预览和调试 |
| 微信开发者工具 | 平台 API、广告、包体检查 |
| 体验版 | 内部试玩、难度和广告调参 |
| 正式版 | 微信小游戏上线 |

当前仓库尚未初始化具体 Cocos 项目，安装、开发、测试和构建命令需要在引入工具链后补充到 `AGENTS.md` 和本节。

## 可观测性

重点记录：

- 关卡开始、通关、失败、重玩。
- 步数、耗时、使用提示次数、星级。
- 广告请求、展示、完成、关闭、失败、奖励发放。
- 新手关流失点。
- 不同广告位的点击率和完成率。
- 插屏展示后的下一关进入率。

## 性能与扩展性

- 4x4 到 6x4 牌阵均应保持流畅。
- 牌面资源使用图集或主题包，减少加载碎片。
- 关卡内避免频繁创建销毁节点，可复用卡牌节点。
- 广告预加载应发生在首页、关卡间或结算页，不阻塞翻牌操作。
- 云端配置失败时回退本地配置。

## 关键风险和技术决策

| 决策/风险 | 当前建议 | 取舍 |
| --- | --- | --- |
| 引擎选择 | Cocos Creator + TypeScript | 开发效率高，适合 2D 微信小游戏 |
| 后端选择 | 第一期无后端，第 5 期再接云开发 | 降低早期复杂度，后续可扩展 |
| 广告策略 | 激励视频优先，插屏低频 | 收益起步稳，体验风险低 |
| 状态管理 | 显式游戏状态机 | 增加少量结构，显著降低点击错乱 |
| 配置方式 | 本地 JSON 起步，远程配置后置 | 快速上线，保留运营调参空间 |
