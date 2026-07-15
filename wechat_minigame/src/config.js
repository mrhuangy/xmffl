const ASSET_ROOT = "assets";

const levelConfigs = Array.from({ length: 40 }, (_, index) => {
  const levelId = index + 1;
  const base = {
    levelId,
    mode: "normal",
    themeId: "animal",
    initialPreviewMs: 2000,
    flipBackDelayMs: 700
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
      rows: 4,
      cols: 4,
      pairCount: 8,
      levelTimeLimitSeconds: 90,
      maxMismatchCount: 10,
      excellentStepThreshold: 10,
      normalStepThreshold: 15,
      excellentTimeThreshold: 50,
      normalTimeThreshold: 78
    };
  }

  return {
    ...base,
    rows: 4,
    cols: 4,
    pairCount: 8,
    levelTimeLimitSeconds: 120,
    maxMismatchCount: 12,
    excellentStepThreshold: 12,
    normalStepThreshold: 18,
    excellentTimeThreshold: 70,
    normalTimeThreshold: 105
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
