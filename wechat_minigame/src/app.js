const { AssetLoader } = require("./assets");
const { imageAssets, audioAssets, levelConfigs } = require("./config");
const { MatchGame } = require("./game-logic");
const { Renderer, clamp } = require("./renderer");
const { ProgressStore } = require("./storage");

const COIN_REWARD_AD_UNIT_ID = "";
const STAMINA_REWARD_AD_UNIT_ID = "";
const TOOL_REWARD_AD_UNIT_ID = "";

class GameApp {
  constructor() {
    this.canvas = wx.createCanvas();
    this.ctx = this.canvas.getContext("2d");
    this.assets = new AssetLoader(this.canvas);
    this.renderer = new Renderer(this.ctx, this.assets);
    this.progressStore = new ProgressStore();
    this.scene = "loading";
    this.progress = this.progressStore.load();
    this.game = null;
    this.cardRects = [];
    this.removedAnimations = [];
    this.buttons = {};
    this.levelCards = [];
    this.levelPage = 0;
    this.hintIndexes = [];
    this.hintEndAt = 0;
    this.previewAgainIndexes = [];
    this.previewAgainEndAt = 0;
    this.previewAgainCount = 3;
    this.removePairCount = 3;
    this.mismatchTimerAt = 0;
    this.previewEndAt = 0;
    this.gamePaused = false;
    this.pauseStartedAt = 0;
    this.toolAdDialog = null;
    this.toolRewardAd = null;
    this.pendingToolReward = null;
    this.loadingProgress = 0;
    this.toast = null;
    this.settingsOpen = false;
    this.coinAdDialogOpen = false;
    this.staminaShopOpen = false;
    this.staminaExchangeConfirm = null;
    this.musicEnabled = true;
    this.sfxEnabled = true;
    this.coinRewardAd = null;
    this.staminaRewardAd = null;
    this.width = 0;
    this.height = 0;
    this.dpr = 1;
  }

  start() {
    this.resize();
    wx.onTouchStart((event) => this.handleTouch(event));
    if (wx.onWindowResize) {
      wx.onWindowResize(() => this.resize());
    }

    this.assets.loadAudio(audioAssets);
    this.assets.loadImages(imageAssets, (loaded, total) => {
      this.loadingProgress = total > 0 ? loaded / total : 1;
    }).then(() => {
      this.loadingProgress = 1;
      this.scene = "home";
      this.playBgm("allBgm");
    });

    this.loop();
  }

  resize() {
    const info = wx.getSystemInfoSync();
    this.width = info.windowWidth;
    this.height = info.windowHeight;
    this.dpr = info.pixelRatio || 1;
    this.canvas.width = Math.floor(this.width * this.dpr);
    this.canvas.height = Math.floor(this.height * this.dpr);
    this.ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
  }

  loop() {
    this.update();
    this.render();
    const nextFrame = wx.requestAnimationFrame || requestAnimationFrame;
    nextFrame(() => this.loop());
  }

  update() {
    const now = Date.now();
    if (this.toast && now >= this.toast.until) {
      this.toast = null;
    }

    if (!this.game) {
      return;
    }

    if (this.scene !== "game") {
      return;
    }

    if (this.gamePaused) {
      return;
    }

    if (this.game.state === "previewing_cards" && now >= this.previewEndAt) {
      this.game.finishPreview(now);
    }

    if (this.game.state === "resolving_mismatch" && now >= this.mismatchTimerAt) {
      this.game.finishMismatch();
    }

    if (this.hintIndexes.length > 0 && now >= this.hintEndAt) {
      this.hintIndexes = [];
    }

    if (this.previewAgainIndexes.length > 0 && now >= this.previewAgainEndAt) {
      this.game.hideRevealed(this.previewAgainIndexes);
      this.previewAgainIndexes = [];
    }

    if (this.removedAnimations.length > 0) {
      this.removedAnimations = this.removedAnimations.filter((item) => now < item.endAt);
    }

    const wasEnded = this.game.isEnded();
    this.game.tick(now);
    if (!wasEnded && this.game.isEnded() && this.game.lastResult) {
      this.onLevelEnded(this.game.lastResult);
    }
  }

  render() {
    this.renderer.clear(this.width, this.height);

    if (this.scene === "loading") {
      this.drawLoading();
      return;
    }

    if (this.scene === "home") {
      this.layoutHome();
      this.renderer.home({
        width: this.width,
        height: this.height,
        buttons: this.buttons,
        progress: this.progress,
        toast: this.toast,
        settingsOpen: this.settingsOpen,
        coinAdDialogOpen: this.coinAdDialogOpen,
        staminaShopOpen: this.staminaShopOpen,
        staminaExchangeConfirm: this.staminaExchangeConfirm,
        settings: {
          musicEnabled: this.musicEnabled,
          sfxEnabled: this.sfxEnabled
        }
      });
      return;
    }

    if (this.scene === "levels") {
      this.layoutLevels();
      this.renderer.levels({
        width: this.width,
        height: this.height,
        buttons: this.buttons,
        levels: this.levelCards,
        currentPage: this.levelPage,
        pageCount: Math.ceil(levelConfigs.length / 20),
        progress: this.progress,
        toast: this.toast,
        staminaShopOpen: this.staminaShopOpen,
        staminaExchangeConfirm: this.staminaExchangeConfirm
      });
      return;
    }

    if (this.scene === "game" && this.game) {
      this.layoutGame();
      this.renderer.game({
        width: this.width,
        height: this.height,
        buttons: this.buttons,
        game: this.game,
        progress: this.progress,
        previewAgainCount: this.previewAgainCount,
        removePairCount: this.removePairCount,
        cardRects: this.cardRects,
        hintIndexes: this.hintIndexes,
        removedAnimations: this.removedAnimations,
        gamePaused: this.gamePaused,
        toolAdDialog: this.toolAdDialog,
        settings: {
          musicEnabled: this.musicEnabled,
          sfxEnabled: this.sfxEnabled
        }
      });
    }
  }

  drawLoading() {
    this.renderer.loading({
      width: this.width,
      height: this.height,
      progress: this.loadingProgress
    });
  }

  layoutHome() {
    const selectW = clamp(this.width * 0.55, 210, 270);
    const selectH = selectW * 0.31;
    const startW = selectW;
    const startH = startW * (538 / 584);
    const startX = (this.width - startW) / 2;
    const selectX = (this.width - selectW) / 2;
    const startY = this.height * 0.42;
    const selectY = startY + startH * 1.05;
    const timeModeY = selectY + selectH * 1.08;
    const hudY = Math.max(34, this.height * 0.05);
    const operationSize = clamp(this.width * 0.09, 34, 40);
    const hudPlayerH = Math.max(38, Math.min(46, this.width * 0.11));
    const hudPlayerW = Math.max(94, Math.min(116, this.width * 0.29));
    const hudGap = Math.max(18, Math.min(22, this.width * 0.052));
    const hudStatH = Math.max(26, Math.min(30, hudPlayerH * 0.68));
    const staminaW = Math.max(62, Math.min(70, this.width * 0.165));
    const staminaX = 14 + hudPlayerW + Math.max(8, this.width * 0.02);
    const coinX = staminaX + staminaW + hudGap;
    const safeRight = this.width - 92;
    const coinW = Math.max(62, Math.min(Math.max(66, Math.min(74, this.width * 0.175)), safeRight - coinX));
    const statY = hudY + Math.max(2, this.height * 0.008);
    const plusRadius = hudStatH * 0.25;
    const coinPlusX = coinX + coinW - plusRadius * 0.9;
    const coinPlusY = statY + hudStatH * 0.5;
    const staminaPlusX = staminaX + staminaW - plusRadius * 0.9;
    const staminaPlusY = statY + hudStatH * 0.5;
    const shopX = this.width * 0.18;
    const shopW = this.width * 0.64;
    const shopY = this.height * 0.31;
    const rowH = 38;
    const rowGap = 10;
    this.buttons = {
      start: { x: startX, y: startY, w: startW, h: startH },
      levels: { x: selectX, y: selectY, w: selectW, h: selectH },
      timeMode: { x: selectX, y: timeModeY, w: selectW, h: selectH },
      operation: { x: this.width - operationSize - 18, y: hudY + 44, w: operationSize, h: operationSize },
      coinPlus: { x: coinPlusX - 16, y: coinPlusY - 16, w: 32, h: 32 },
      staminaPlus: { x: staminaPlusX - 16, y: staminaPlusY - 16, w: 32, h: 32 },
      settingsMusic: { x: this.width * 0.28, y: this.height * 0.405, w: this.width * 0.44, h: 42 },
      settingsSfx: { x: this.width * 0.28, y: this.height * 0.475, w: this.width * 0.44, h: 42 },
      settingsClose: { x: this.width * 0.34, y: this.height * 0.56, w: this.width * 0.32, h: 38 },
      coinAdConfirm: { x: this.width * 0.52, y: this.height * 0.52, w: this.width * 0.24, h: 38 },
      coinAdCancel: { x: this.width * 0.24, y: this.height * 0.52, w: this.width * 0.24, h: 38 },
      staminaAd: { x: shopX, y: shopY + 58, w: shopW, h: rowH },
      staminaBuy1: { x: shopX, y: shopY + 58 + (rowH + rowGap), w: shopW, h: rowH },
      staminaBuy3: { x: shopX, y: shopY + 58 + (rowH + rowGap) * 2, w: shopW, h: rowH },
      staminaBuy5: { x: shopX, y: shopY + 58 + (rowH + rowGap) * 3, w: shopW, h: rowH },
      staminaShopClose: { x: this.width * 0.34, y: shopY + 58 + (rowH + rowGap) * 4 + 8, w: this.width * 0.32, h: 36 },
      staminaExchangeConfirm: { x: this.width * 0.52, y: this.height * 0.52, w: this.width * 0.24, h: 38 },
      staminaExchangeCancel: { x: this.width * 0.24, y: this.height * 0.52, w: this.width * 0.24, h: 38 }
    };
  }

  layoutLevels() {
    const backSize = clamp(this.width * 0.13, 46, 58);
    const panelX = this.width * 0.075;
    const panelY = this.height * 0.255;
    const panelW = this.width * 0.85;
    const gridPadX = panelW * 0.065;
    const gap = Math.max(9, Math.min(14, this.width * 0.028));
    const cols = 5;
    const side = Math.floor((panelW - gridPadX * 2 - gap * (cols - 1)) / cols);
    const cardH = Math.floor(side * (410 / 330));
    const startX = panelX + gridPadX;
    const startY = panelY + this.height * 0.115;
    const statH = Math.max(26, Math.min(30, this.width * 0.075));
    const staminaW = Math.max(62, Math.min(70, this.width * 0.165));
    const staminaX = this.width - staminaW - 112;
    const staminaY = Math.max(44, this.height * 0.055);
    const plusRadius = statH * 0.25;
    const staminaPlusX = staminaX + staminaW - plusRadius * 0.9;
    const staminaPlusY = staminaY + statH * 0.5;
    const shopX = this.width * 0.18;
    const shopW = this.width * 0.64;
    const shopY = this.height * 0.31;
    const rowH = 38;
    const rowGap = 10;

    this.buttons = {
      back: { x: 18, y: Math.max(28, this.height * 0.045), w: backSize, h: backSize },
      levelPrev: { x: this.width * 0.23, y: panelY + Math.min(this.height * 0.56, this.height - panelY - 118) - 54, w: 44, h: 44 },
      levelNext: { x: this.width * 0.66, y: panelY + Math.min(this.height * 0.56, this.height - panelY - 118) - 54, w: 44, h: 44 },
      staminaPlus: { x: staminaPlusX - 16, y: staminaPlusY - 16, w: 32, h: 32 },
      staminaAd: { x: shopX, y: shopY + 58, w: shopW, h: rowH },
      staminaBuy1: { x: shopX, y: shopY + 58 + (rowH + rowGap), w: shopW, h: rowH },
      staminaBuy3: { x: shopX, y: shopY + 58 + (rowH + rowGap) * 2, w: shopW, h: rowH },
      staminaBuy5: { x: shopX, y: shopY + 58 + (rowH + rowGap) * 3, w: shopW, h: rowH },
      staminaShopClose: { x: this.width * 0.34, y: shopY + 58 + (rowH + rowGap) * 4 + 8, w: this.width * 0.32, h: 36 },
      staminaExchangeConfirm: { x: this.width * 0.52, y: this.height * 0.52, w: this.width * 0.24, h: 38 },
      staminaExchangeCancel: { x: this.width * 0.24, y: this.height * 0.52, w: this.width * 0.24, h: 38 }
    };

    const pageSize = 20;
    const pageCount = Math.ceil(levelConfigs.length / pageSize);
    this.levelPage = Math.max(0, Math.min(this.levelPage, pageCount - 1));
    const pageLevels = levelConfigs.slice(this.levelPage * pageSize, (this.levelPage + 1) * pageSize);

    this.levelCards = pageLevels.map((level, index) => {
      const progress = this.progressStore.load();
      const row = Math.floor(index / cols);
      const col = index % cols;
      return {
        level,
        unlocked: level.levelId <= progress.currentLevel,
        stars: progress.levelStars[String(level.levelId)] || 0,
        isCurrent: level.levelId === progress.currentLevel,
        rect: {
          x: startX + col * (side + gap),
          y: startY + row * (cardH + gap),
          w: side,
          h: cardH
        }
      };
    });
  }

  layoutGame() {
    const topButtonSize = clamp(this.width * 0.13, 48, 58);
    const boardX = this.width * 0.045;
    const boardY = this.height * 0.225;
    const boardW = this.width * 0.91;
    const boardH = this.height * 0.44;
    const toolY = boardY + boardH + 30;
    const pausePanelW = Math.min(this.width * 0.78, 320);
    const pausePanelX = (this.width - pausePanelW) / 2;
    const pausePanelY = this.height * 0.27;
    const pauseCloseSize = 30;
    this.buttons = {
      back: { x: 24, y: Math.max(24, this.height * 0.04), w: topButtonSize, h: topButtonSize },
      hint: { x: this.width * 0.2, y: toolY, w: this.width * 0.18, h: 58 },
      shuffle: { x: this.width * 0.41, y: toolY, w: this.width * 0.18, h: 58 },
      pair: { x: this.width * 0.62, y: toolY, w: this.width * 0.18, h: 58 },
      pauseMusic: { x: this.width * 0.28, y: this.height * 0.345, w: this.width * 0.44, h: 42 },
      pauseSfx: { x: this.width * 0.28, y: this.height * 0.415, w: this.width * 0.44, h: 42 },
      pauseRetry: { x: this.width * 0.24, y: this.height * 0.495, w: this.width * 0.22, h: 38 },
      pauseHome: { x: this.width * 0.54, y: this.height * 0.495, w: this.width * 0.22, h: 38 },
      pauseClose: { x: pausePanelX + pausePanelW - pauseCloseSize - 10, y: pausePanelY + 10, w: pauseCloseSize, h: pauseCloseSize },
      toolAdCancel: { x: this.width * 0.24, y: this.height * 0.52, w: this.width * 0.24, h: 38 },
      toolAdConfirm: { x: this.width * 0.52, y: this.height * 0.52, w: this.width * 0.24, h: 38 },
      retry: { x: this.width * 0.18, y: this.height * 0.28 + 208, w: this.width * 0.28, h: 42 },
      next: { x: this.width * 0.54, y: this.height * 0.28 + 208, w: this.width * 0.28, h: 42 }
    };

    const level = this.game.level;
    const gap = level.cols >= 5 ? Math.max(8, Math.min(12, this.width * 0.025)) : 12;
    const padX = boardW * 0.09;
    const padY = boardH * 0.08;
    const maxCardW = (boardW - padX * 2 - gap * (level.cols - 1)) / level.cols;
    const maxCardH = (boardH - padY * 2 - gap * (level.rows - 1)) / level.rows;
    const cardSize = Math.floor(Math.min(maxCardW, maxCardH));
    const cardsW = cardSize * level.cols + gap * (level.cols - 1);
    const boardCardsH = cardSize * level.rows + gap * (level.rows - 1);
    const startX = boardX + (this.width * 0.91 - cardsW) / 2;
    const startY = boardY + (this.height * 0.44 - boardCardsH) / 2;

    const emptySlots = level.emptySlots || [];
    const boardSlots = Array.from({ length: level.rows * level.cols }, (_, index) => index)
      .filter((index) => !emptySlots.includes(index));

    this.cardRects = this.game.cards.map((card, index) => {
      const boardIndex = boardSlots[index] ?? index;
      const row = Math.floor(boardIndex / level.cols);
      const col = boardIndex % level.cols;
      return {
        x: startX + col * (cardSize + gap),
        y: startY + row * (cardSize + gap),
        w: cardSize,
        h: cardSize
      };
    });
  }

  handleTouch(event) {
    const touch = event.touches && event.touches[0];
    if (!touch) {
      return;
    }

    const x = touch.clientX;
    const y = touch.clientY;

    if (this.scene === "home") {
      if (this.staminaExchangeConfirm) {
        this.handleStaminaExchangeConfirmTouch(x, y);
        return;
      }

      if (this.staminaShopOpen) {
        this.handleStaminaShopTouch(x, y);
        return;
      }

      if (this.coinAdDialogOpen) {
        this.handleCoinAdDialogTouch(x, y);
        return;
      }

      if (this.settingsOpen) {
        this.handleSettingsTouch(x, y);
        return;
      }

      if (hit(this.buttons.start, x, y)) {
        this.playSfx("click");
        this.startLevel(this.progress.currentLevel);
      } else if (hit(this.buttons.levels, x, y)) {
        this.playSfx("click");
        this.scene = "levels";
      } else if (hit(this.buttons.timeMode, x, y)) {
        this.playSfx("click");
        this.toast = {
          text: "\u5f00\u53d1\u4e2d\uff0c\u656c\u8bf7\u671f\u5f85",
          until: Date.now() + 2000
        };
      } else if (hit(this.buttons.operation, x, y)) {
        this.playSfx("click");
        this.settingsOpen = true;
      } else if (hit(this.buttons.staminaPlus, x, y)) {
        this.playSfx("click");
        this.staminaShopOpen = true;
      } else if (hit(this.buttons.coinPlus, x, y)) {
        this.playSfx("click");
        this.coinAdDialogOpen = true;
      }
      return;
    }

    if (this.scene === "levels") {
      if (this.staminaExchangeConfirm) {
        this.handleStaminaExchangeConfirmTouch(x, y);
        return;
      }

      if (this.staminaShopOpen) {
        this.handleStaminaShopTouch(x, y);
        return;
      }

      if (hit(this.buttons.back, x, y)) {
        this.playSfx("click");
        this.scene = "home";
        return;
      }

      if (hit(this.buttons.levelPrev, x, y)) {
        this.playSfx("click");
        this.levelPage = Math.max(0, this.levelPage - 1);
        return;
      }

      if (hit(this.buttons.levelNext, x, y)) {
        this.playSfx("click");
        this.levelPage = Math.min(Math.ceil(levelConfigs.length / 20) - 1, this.levelPage + 1);
        return;
      }

      if (hit(this.buttons.staminaPlus, x, y)) {
        this.playSfx("click");
        this.staminaShopOpen = true;
        return;
      }

      const card = this.levelCards.find((item) => hit(item.rect, x, y));
      if (card && card.unlocked) {
        this.playSfx("click");
        this.startLevel(card.level.levelId);
      }
      return;
    }

    if (this.scene === "game" && this.game) {
      if (this.gamePaused) {
        if (this.toolAdDialog) {
          this.handleToolAdDialogTouch(x, y);
        } else {
          this.handlePauseMenuTouch(x, y);
        }
        return;
      }

      if (this.game.isEnded()) {
        this.handleResultTouch(x, y);
        return;
      }

      if (hit(this.buttons.back, x, y)) {
        this.playSfx("click");
        this.pauseGame();
        return;
      }

      if (hit(this.buttons.hint, x, y)) {
        this.useHint();
        return;
      }

      if (hit(this.buttons.shuffle, x, y)) {
        this.previewAgain();
        return;
      }

      if (hit(this.buttons.pair, x, y)) {
        this.removeOnePair();
        return;
      }

      const cardIndex = this.cardRects.findIndex((rect) => hit(rect, x, y));
      if (cardIndex >= 0) {
        this.flip(cardIndex);
      }
    }
  }

  pauseGame() {
    if (this.gamePaused) {
      return;
    }

    this.gamePaused = true;
    this.pauseStartedAt = Date.now();
  }

  resumeGame() {
    if (!this.gamePaused) {
      return;
    }

    const pausedMs = Date.now() - this.pauseStartedAt;
    if (this.game && this.game.startedAt > 0) {
      this.game.startedAt += pausedMs;
    }
    if (this.previewEndAt > 0) {
      this.previewEndAt += pausedMs;
    }
    if (this.mismatchTimerAt > 0) {
      this.mismatchTimerAt += pausedMs;
    }
    if (this.hintEndAt > 0) {
      this.hintEndAt += pausedMs;
    }
    if (this.previewAgainEndAt > 0) {
      this.previewAgainEndAt += pausedMs;
    }
    this.gamePaused = false;
    this.pauseStartedAt = 0;
  }

  handlePauseMenuTouch(x, y) {
    if (hit(this.buttons.pauseMusic, x, y)) {
      this.playSfx("click");
      this.musicEnabled = !this.musicEnabled;
      if (this.musicEnabled) {
        this.playBgm("gameBgm");
      } else {
        this.assets.stopBgm("allBgm");
        this.assets.stopBgm("gameBgm");
      }
      return;
    }

    if (hit(this.buttons.pauseSfx, x, y)) {
      this.sfxEnabled = !this.sfxEnabled;
      this.playSfx("click");
      return;
    }

    if (hit(this.buttons.pauseRetry, x, y)) {
      this.playSfx("click");
      this.gamePaused = false;
      this.startLevel(this.game.level.levelId);
      return;
    }

    if (hit(this.buttons.pauseHome, x, y)) {
      this.playSfx("click");
      this.gamePaused = false;
      this.scene = "home";
      this.assets.stopBgm("gameBgm");
      this.playBgm("allBgm");
      return;
    }

    if (hit(this.buttons.pauseClose, x, y)) {
      this.playSfx("click");
      this.resumeGame();
    }
  }

  openToolAdDialog(type, label) {
    this.playSfx("wrong");
    this.pauseGame();
    this.toolAdDialog = { type, label };
  }

  handleToolAdDialogTouch(x, y) {
    if (hit(this.buttons.toolAdCancel, x, y)) {
      this.playSfx("click");
      this.toolAdDialog = null;
      this.resumeGame();
      return;
    }

    if (hit(this.buttons.toolAdConfirm, x, y)) {
      this.playSfx("click");
      const type = this.toolAdDialog.type;
      this.toolAdDialog = null;
      this.showToolRewardAd(type);
    }
  }

  addToolCount(type) {
    if (type === "hint") {
      this.progress = this.progressStore.addHints(1);
    } else if (type === "previewAgain") {
      this.previewAgainCount += 1;
    } else if (type === "removePair") {
      this.removePairCount += 1;
    }
  }

  showToolRewardAd(type) {
    if (!TOOL_REWARD_AD_UNIT_ID || typeof wx.createRewardedVideoAd !== "function") {
      this.toast = {
        text: "\u5e7f\u544a\u6682\u672a\u914d\u7f6e",
        until: Date.now() + 2000
      };
      this.resumeGame();
      return;
    }

    this.pendingToolReward = type;
    if (!this.toolRewardAd) {
      this.toolRewardAd = wx.createRewardedVideoAd({ adUnitId: TOOL_REWARD_AD_UNIT_ID });
      this.toolRewardAd.onClose((result) => {
        if (result && result.isEnded && this.pendingToolReward) {
          this.addToolCount(this.pendingToolReward);
          this.toast = {
            text: "\u6b21\u6570+1",
            until: Date.now() + 2000
          };
        }
        this.pendingToolReward = null;
        this.resumeGame();
      });
      this.toolRewardAd.onError(() => {
        this.toast = {
          text: "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e",
          until: Date.now() + 2000
        };
        this.pendingToolReward = null;
        this.resumeGame();
      });
    }

    this.toolRewardAd.show().catch(() => {
      this.toolRewardAd.load()
        .then(() => this.toolRewardAd.show())
        .catch(() => {
          this.toast = {
            text: "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e",
            until: Date.now() + 2000
          };
          this.pendingToolReward = null;
          this.resumeGame();
        });
    });
  }

  handleResultTouch(x, y) {
    if (hit(this.buttons.retry, x, y)) {
      this.playSfx("click");
      this.startLevel(this.game.level.levelId);
      return;
    }

    if (hit(this.buttons.next, x, y)) {
      this.playSfx("click");
      if (this.game.lastResult && this.game.lastResult.success) {
        this.startLevel(Math.min(this.game.level.levelId + 1, levelConfigs.length));
      } else {
        this.scene = "levels";
        this.assets.stopBgm("gameBgm");
        this.playBgm("allBgm");
      }
    }
  }

  handleSettingsTouch(x, y) {
    if (hit(this.buttons.settingsMusic, x, y)) {
      this.playSfx("click");
      this.musicEnabled = !this.musicEnabled;
      if (this.musicEnabled) {
        this.playBgm("allBgm");
      } else {
        this.assets.stopBgm("allBgm");
        this.assets.stopBgm("gameBgm");
      }
      return;
    }

    if (hit(this.buttons.settingsSfx, x, y)) {
      this.sfxEnabled = !this.sfxEnabled;
      this.playSfx("click");
      return;
    }

    if (hit(this.buttons.settingsClose, x, y)) {
      this.playSfx("click");
      this.settingsOpen = false;
    }
  }

  handleCoinAdDialogTouch(x, y) {
    if (hit(this.buttons.coinAdCancel, x, y)) {
      this.playSfx("click");
      this.coinAdDialogOpen = false;
      return;
    }

    if (hit(this.buttons.coinAdConfirm, x, y)) {
      this.playSfx("click");
      this.coinAdDialogOpen = false;
      this.showCoinRewardAd();
    }
  }

  handleStaminaShopTouch(x, y) {
    if (hit(this.buttons.staminaAd, x, y)) {
      this.playSfx("click");
      this.staminaShopOpen = false;
      this.showStaminaRewardAd();
      return;
    }

    if (hit(this.buttons.staminaBuy1, x, y)) {
      this.openStaminaExchangeConfirm(99, 1);
      return;
    }

    if (hit(this.buttons.staminaBuy3, x, y)) {
      this.openStaminaExchangeConfirm(266, 3);
      return;
    }

    if (hit(this.buttons.staminaBuy5, x, y)) {
      this.openStaminaExchangeConfirm(388, 5);
      return;
    }

    if (hit(this.buttons.staminaShopClose, x, y)) {
      this.playSfx("click");
      this.staminaShopOpen = false;
    }
  }

  openStaminaExchangeConfirm(cost, amount) {
    this.playSfx("click");
    this.staminaExchangeConfirm = { cost, amount };
  }

  handleStaminaExchangeConfirmTouch(x, y) {
    if (hit(this.buttons.staminaExchangeCancel, x, y)) {
      this.playSfx("click");
      this.staminaExchangeConfirm = null;
      return;
    }

    if (hit(this.buttons.staminaExchangeConfirm, x, y)) {
      this.playSfx("click");
      const { cost, amount } = this.staminaExchangeConfirm;
      const result = this.progressStore.exchangeCoinsForStamina(cost, amount);
      this.progress = result.progress;
      this.staminaExchangeConfirm = null;
      this.staminaShopOpen = false;
      this.toast = {
        text: result.success ? `\u83b7\u5f97${amount}\u4f53\u529b` : "\u91d1\u5e01\u4e0d\u8db3",
        until: Date.now() + 2000
      };
    }
  }

  showCoinRewardAd() {
    if (!COIN_REWARD_AD_UNIT_ID || typeof wx.createRewardedVideoAd !== "function") {
      this.toast = {
        text: "\u5e7f\u544a\u6682\u672a\u914d\u7f6e",
        until: Date.now() + 2000
      };
      return;
    }

    if (!this.coinRewardAd) {
      this.coinRewardAd = wx.createRewardedVideoAd({ adUnitId: COIN_REWARD_AD_UNIT_ID });
      this.coinRewardAd.onClose((result) => {
        if (result && result.isEnded) {
          this.progress = this.progressStore.addCoins(100);
          this.toast = {
            text: "\u83b7\u5f97100\u91d1\u5e01",
            until: Date.now() + 2000
          };
        }
      });
      this.coinRewardAd.onError(() => {
        this.toast = {
          text: "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e",
          until: Date.now() + 2000
        };
      });
    }

    this.coinRewardAd.show().catch(() => {
      this.coinRewardAd.load()
        .then(() => this.coinRewardAd.show())
        .catch(() => {
          this.toast = {
            text: "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e",
            until: Date.now() + 2000
          };
        });
    });
  }

  showStaminaRewardAd() {
    if (!STAMINA_REWARD_AD_UNIT_ID || typeof wx.createRewardedVideoAd !== "function") {
      this.toast = {
        text: "\u5e7f\u544a\u6682\u672a\u914d\u7f6e",
        until: Date.now() + 2000
      };
      return;
    }

    if (!this.staminaRewardAd) {
      this.staminaRewardAd = wx.createRewardedVideoAd({ adUnitId: STAMINA_REWARD_AD_UNIT_ID });
      this.staminaRewardAd.onClose((result) => {
        if (result && result.isEnded) {
          this.progress = this.progressStore.addStamina(3);
          this.toast = {
            text: "\u83b7\u5f973\u4f53\u529b",
            until: Date.now() + 2000
          };
        }
      });
      this.staminaRewardAd.onError(() => {
        this.toast = {
          text: "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e",
          until: Date.now() + 2000
        };
      });
    }

    this.staminaRewardAd.show().catch(() => {
      this.staminaRewardAd.load()
        .then(() => this.staminaRewardAd.show())
        .catch(() => {
          this.toast = {
            text: "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e",
            until: Date.now() + 2000
          };
        });
    });
  }

  startLevel(levelId) {
    this.scene = "game";
    this.progress = this.progressStore.load();
    this.game = new MatchGame(levelId);
    this.gamePaused = false;
    this.pauseStartedAt = 0;
    this.toolAdDialog = null;
    this.pendingToolReward = null;
    this.previewEndAt = Date.now() + this.game.level.initialPreviewMs;
    this.mismatchTimerAt = 0;
    this.hintIndexes = [];
    this.removedAnimations = [];
    this.previewAgainIndexes = [];
    this.previewAgainEndAt = 0;
    this.previewAgainCount = 3;
    this.removePairCount = 3;
    this.assets.stopBgm("allBgm");
    this.playBgm("gameBgm");
  }

  flip(cardIndex) {
    const outcome = this.game.flipCard(cardIndex, Date.now());
    if (outcome.type === "matched") {
      this.addRemovedAnimations(outcome.cards);
      this.playSfx("right");
    } else if (outcome.type === "mismatched") {
      this.playSfx("wrong");
      this.mismatchTimerAt = Date.now() + this.game.level.flipBackDelayMs;
    } else if (outcome.type === "level_complete") {
      this.playSfx("right");
      this.onLevelEnded(outcome.result);
    } else if (outcome.type === "level_failed") {
      this.playSfx("wrong");
      this.onLevelEnded(outcome.result);
    }
  }

  addRemovedAnimations(cards) {
    const now = Date.now();
    this.removedAnimations.push(...(cards || []).map((card) => ({
      ...card,
      rect: this.cardRects[card.index],
      startAt: now,
      endAt: now + 280
    })).filter((item) => item.rect));
  }

  useHint() {
    if (this.progress.hints <= 0) {
      this.openToolAdDialog("hint", "\u63d0\u793a");
      return;
    }

    if (!this.progressStore.spendHint()) {
      this.openToolAdDialog("hint", "\u63d0\u793a");
      return;
    }

    this.progress = this.progressStore.load();
    this.hintIndexes = this.game.useHint();
    this.hintEndAt = Date.now() + 1300;
    this.playSfx(this.hintIndexes.length > 0 ? "right" : "wrong");
  }

  previewAgain() {
    if (this.previewAgainIndexes.length > 0) {
      this.playSfx("wrong");
      return;
    }

    if (this.previewAgainCount <= 0) {
      this.openToolAdDialog("previewAgain", "\u518d\u770b\u4e00\u6b21");
      return;
    }

    const indexes = this.game.revealUnmatched();
    if (indexes.length === 0) {
      this.playSfx("wrong");
      return;
    }

    this.previewAgainCount -= 1;
    this.previewAgainIndexes = indexes;
    this.previewAgainEndAt = Date.now() + 2000;
    this.playSfx("right");
  }

  removeOnePair() {
    if (this.removePairCount <= 0) {
      this.openToolAdDialog("removePair", "\u6d88\u9664\u4e00\u5bf9");
      return;
    }

    const outcome = this.game.removeOnePair();
    if (outcome.type === "ignored") {
      this.playSfx("wrong");
      return;
    }

    this.removePairCount -= 1;
    this.playSfx("right");
    if (outcome.type === "level_complete") {
      this.onLevelEnded(outcome.result);
    }
  }

  playSfx(key) {
    if (this.sfxEnabled) {
      this.assets.play(key);
    }
  }

  playBgm(key) {
    if (this.musicEnabled) {
      this.assets.bgm(key);
    }
  }

  onLevelEnded(result) {
    if (!result || result._saved) {
      return;
    }

    result._saved = true;
    this.progress = this.progressStore.applyResult(result);
  }
}

function hit(rect, x, y) {
  return rect && x >= rect.x && x <= rect.x + rect.w && y >= rect.y && y <= rect.y + rect.h;
}

module.exports = {
  GameApp
};
