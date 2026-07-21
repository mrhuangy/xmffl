const ASSET_ROOT = "assets";

const levelConfigs = Array.from({ length: 100 }, (_, index) => {
  const levelId = index + 1;
  const base = {
    levelId,
    mode: "normal",
    themeId: "animal",
    initialPreviewMs: levelId <= 10 ? 2500 : (levelId <= 40 ? 2200 : 2000),
    flipBackDelayMs: levelId <= 30 ? 800 : (levelId <= 70 ? 700 : 600),
    showSteps: true,
    showTimer: true,
    showMismatch: true,
    hintHighlightMs: 1300,
    coinRewardBase: 10,
    coinRewardStar1: 10,
    coinRewardStar2: 20,
    coinRewardStar3: 30,
    staminaCost: 1
  };

  if (levelId === 1) {
    return {
      ...base,
      rows: 2,
      cols: 2,
      pairCount: 2,
      levelTimeLimitSeconds: 45,
      maxMismatchCount: 6,
      excellentStepThreshold: 3,
      normalStepThreshold: 5,
      excellentTimeThreshold: 20,
      normalTimeThreshold: 35
    };
  }

  if (levelId <= 3) {
    return {
      ...base,
      rows: 3,
      cols: 3,
      pairCount: 4,
      emptySlots: [4],
      levelTimeLimitSeconds: 70,
      maxMismatchCount: 8,
      excellentStepThreshold: 6,
      normalStepThreshold: 10,
      excellentTimeThreshold: 35,
      normalTimeThreshold: 58
    };
  }

  if (levelId <= 5) {
    return {
      ...base,
      rows: 3,
      cols: 4,
      pairCount: 6,
      levelTimeLimitSeconds: 90,
      maxMismatchCount: 10,
      excellentStepThreshold: 9,
      normalStepThreshold: 14,
      excellentTimeThreshold: 45,
      normalTimeThreshold: 75
    };
  }

  return {
    ...base,
    rows: 4,
    cols: 4,
    pairCount: 8,
    levelTimeLimitSeconds: levelId <= 20 ? 120 : (levelId <= 50 ? 110 : (levelId <= 80 ? 100 : 90)),
    maxMismatchCount: levelId <= 30 ? 12 : (levelId <= 60 ? 11 : (levelId <= 85 ? 10 : 9)),
    excellentStepThreshold: levelId <= 30 ? 12 : (levelId <= 70 ? 11 : 10),
    normalStepThreshold: levelId <= 30 ? 18 : (levelId <= 70 ? 17 : 16),
    excellentTimeThreshold: levelId <= 30 ? 70 : (levelId <= 70 ? 65 : 60),
    normalTimeThreshold: levelId <= 30 ? 105 : (levelId <= 70 ? 98 : 90)
  };
});

const animalIcons = [
  "panda",
  "tiger",
  "rabbit",
  "bear",
  "fox",
  "cat",
  "dog",
  "penguin",
  "koala",
  "cow",
  "chicken",
  "dinosaur",
  "elefants",
  "frog",
  "girafa",
  "hedgehog",
  "monkey",
  "owl",
  "seal",
  "sheep"
];

const imageAssets = {
  gameBg: `${ASSET_ROOT}/backgrounds/game-bg.png`,
  logo: `${ASSET_ROOT}/ui/logo.png`,
  loading: `${ASSET_ROOT}/ui/loading.png`,
  homeBg: `${ASSET_ROOT}/backgrounds/home-bg.png`,
  startButton: `${ASSET_ROOT}/ui/game-start.png`,
  selectButton: `${ASSET_ROOT}/ui/select-button.png`,
  timeModeButton: `${ASSET_ROOT}/ui/unlessmodel-button.png`,
  toolBulb: `${ASSET_ROOT}/ui/tool-bulb.png`,
  toolMagnifier: `${ASSET_ROOT}/ui/tool-magnifier.png`,
  toolEraser: `${ASSET_ROOT}/ui/tool-eraser.png`,
  cardBack: `${ASSET_ROOT}/cards/card-back.png`,
  back: `${ASSET_ROOT}/select/back.png`,
  selectTitle: `${ASSET_ROOT}/select/title.png`,
  blockNow: `${ASSET_ROOT}/select/block-now.png`,
  blockNo: `${ASSET_ROOT}/select/block-no.png`,
  blockAlready: `${ASSET_ROOT}/select/block-aleady.png`,
  lock: `${ASSET_ROOT}/select/lock.png`,
  winStar: `${ASSET_ROOT}/select/win-star.png`,
  lossStar: `${ASSET_ROOT}/select/loss-star.png`
};

for (const icon of animalIcons) {
  imageAssets[`animal:${icon}`] = `${ASSET_ROOT}/animals/${icon}.png`;
}

const audioAssets = {
  click: `${ASSET_ROOT}/audio/click.mp3`,
  right: `${ASSET_ROOT}/audio/right.mp3`,
  wrong: `${ASSET_ROOT}/audio/wrong.mp3`,
  allBgm: `${ASSET_ROOT}/audio/all-bgm.mp3`,
  gameBgm: `${ASSET_ROOT}/audio/game-bgm.mp3`
};

module.exports = {
  ASSET_ROOT,
  levelConfigs,
  animalIcons,
  imageAssets,
  audioAssets
};
