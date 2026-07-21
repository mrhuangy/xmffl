const { AssetLoader } = require("./assets");
const { ApiClient } = require("./api-client");
const { AdService } = require("./ad-service");
const { imageAssets, audioAssets, levelConfigs: defaultLevelConfigs } = require("./config");
const { MatchGame } = require("./game-logic");
const { Renderer, clamp } = require("./renderer");
const { ProgressStore } = require("./storage");
const { RewardService } = require("./reward-service");

const STAMINA_RECOVER_INTERVAL_MS = 2 * 60 * 1000;

class GameApp {
  constructor() {
    this.canvas = wx.createCanvas();
    this.ctx = this.canvas.getContext("2d");
    this.assets = new AssetLoader(this.canvas);
    this.renderer = new Renderer(this.ctx, this.assets);
    this.progressStore = new ProgressStore();
    this.apiClient = new ApiClient(this.progressStore);
    this.rewardService = new RewardService(this.progressStore, this.apiClient);
    this.adService = new AdService();
    this.levelConfigs = defaultLevelConfigs;
    this.systemControls = {};
    this.unlimitedStamina = false;
    this.levelConfigPromise = null;
    this.scene = "loading";
    this.progress = this.rewardService.loadProgress();
    this.player = (this.progressStore.loadAuth() || {}).player || null;
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
    this.previewAgainCount = this.progress.previewAgainCount || 3;
    this.removePairCount = this.progress.removePairCount || 3;
    this.mismatchTimerAt = 0;
    this.previewEndAt = 0;
    this.gamePaused = false;
    this.pauseStartedAt = 0;
    this.toolAdDialog = null;
    this.loadingProgress = 0;
    this.toast = null;
    this.settingsOpen = false;
    this.coinAdDialogOpen = false;
    this.staminaShopOpen = false;
    this.staminaExchangeConfirm = null;
    this.audioSettings = this.progressStore.loadSettings();
    this.musicEnabled = this.audioSettings.musicEnabled;
    this.sfxEnabled = this.audioSettings.sfxEnabled;
    this.width = 0;
    this.height = 0;
    this.dpr = 1;
    this.sessionReady = false;
    this.sessionPromise = null;
    this.startingLevel = false;
    this.staminaRefreshPromise = null;
    this.needsStaminaSync = false;
    this.nextStaminaRefreshRetryAt = 0;
    this.toolUsePending = false;
  }

  start() {
    this.resize();
    wx.onTouchStart((event) => this.handleTouch(event));
    if (wx.onWindowResize) {
      wx.onWindowResize(() => this.resize());
    }

    this.assets.loadAudio(audioAssets);
    this.levelConfigPromise = this.syncInitConfig();
    this.sessionPromise = this.syncSession();
    this.assets.loadImages(imageAssets, (loaded, total) => {
      this.loadingProgress = total > 0 ? loaded / total : 1;
    }).then(() => {
      this.loadingProgress = 1;
      this.scene = "home";
      this.playBgm("allBgm");
    });

    this.loop();
  }

  syncSession() {
    this.sessionPromise = this.apiClient.login()
      .then((data) => {
        if (data && data.player) {
          this.player = data.player;
        }
        if (data && data.progress) {
          this.progress = this.progressStore.load();
          this.syncToolCountsFromProgress(this.progress);
          this.sessionReady = true;
        }
        return this.apiClient.fetchProgress();
      })
      .then((progress) => {
        this.progress = progress;
        this.syncToolCountsFromProgress(progress);
        this.sessionReady = true;
      })
      .catch((error) => {
        this.sessionReady = false;
        if (typeof console !== "undefined" && console.warn) {
          console.warn("sync session failed", error);
        }
      });
    return this.sessionPromise;
  }

  syncInitConfig() {
    this.levelConfigPromise = this.apiClient.fetchInitConfig()
      .then((config) => {
        if (config && Array.isArray(config.levels) && config.levels.length > 0) {
          this.levelConfigs = config.levels;
        }
        if (config && config.systemControls) {
          this.applySystemControls(config.systemControls);
        }
        return this.levelConfigs;
      })
      .catch((error) => {
        if (typeof console !== "undefined" && console.warn) {
          console.warn("sync init config failed", error);
        }
        return this.syncLevelConfigs();
      });
    return this.levelConfigPromise;
  }

  syncLevelConfigs() {
    this.levelConfigPromise = this.apiClient.fetchLevels()
      .then((levels) => {
        if (levels && levels.length > 0) {
          this.levelConfigs = levels;
        }
        return this.levelConfigs;
      })
      .catch((error) => {
        if (typeof console !== "undefined" && console.warn) {
          console.warn("sync level configs failed", error);
        }
        this.levelConfigPromise = null;
        return this.levelConfigs;
      });
    return this.levelConfigPromise;
  }

  ensureLevelConfigs() {
    return this.levelConfigPromise || this.syncInitConfig();
  }

  applySystemControls(controls) {
    this.systemControls = controls || {};
    this.unlimitedStamina = this.systemControls["game.unlimited_stamina"] === true;
  }

  syncToolCountsFromProgress(progress) {
    if (!progress) {
      return;
    }
    if (typeof progress.previewAgainCount === "number") {
      this.previewAgainCount = progress.previewAgainCount;
    }
    if (typeof progress.removePairCount === "number") {
      this.removePairCount = progress.removePairCount;
    }
  }

  consumeTool(toolType, label) {
    if (this.toolUsePending) {
      return Promise.resolve(false);
    }
    this.toolUsePending = true;
    return this.apiClient.changeToolCount(toolType, -1)
      .then((progress) => {
        this.progress = progress;
        this.syncToolCountsFromProgress(progress);
        return true;
      })
      .catch((error) => {
        const message = error && error.message ? error.message : "";
        if (message.includes("insufficient tool charges")) {
          this.openToolAdDialog(toolType, label);
        } else {
          this.showToast("\u6b21\u6570\u540c\u6b65\u5931\u8d25\uff0c\u8bf7\u91cd\u8bd5");
        }
        if (typeof console !== "undefined" && console.warn) {
          console.warn("consume tool failed", error);
        }
        return false;
      })
      .finally(() => {
        this.toolUsePending = false;
      });
  }

  refreshStaminaWhenDue(now) {
    if (this.unlimitedStamina) {
      return;
    }
    if (!this.progress) {
      return;
    }

    const stamina = this.progress.stamina ?? 0;
    const maxStamina = this.progress.maxStamina ?? 0;
    const nextRecoverAt = this.progress.nextStaminaRecoverAt || 0;
    if (stamina < maxStamina && !nextRecoverAt) {
      this.needsStaminaSync = true;
    }

    if (this.projectStaminaRecovery(now)) {
      this.needsStaminaSync = true;
    }

    if (!this.needsStaminaSync || now < this.nextStaminaRefreshRetryAt || this.staminaRefreshPromise) {
      return;
    }

    this.staminaRefreshPromise = this.apiClient.fetchProgress()
      .then((progress) => {
        this.progress = progress;
        this.syncToolCountsFromProgress(progress);
        this.needsStaminaSync = false;
        this.nextStaminaRefreshRetryAt = 0;
      })
      .catch((error) => {
        this.nextStaminaRefreshRetryAt = Date.now() + 10000;
        if (typeof console !== "undefined" && console.warn) {
          console.warn("refresh stamina failed", error);
        }
      })
      .finally(() => {
        this.staminaRefreshPromise = null;
      });
  }

  projectStaminaRecovery(now) {
    const stamina = this.progress.stamina ?? 0;
    const maxStamina = this.progress.maxStamina ?? 0;
    const nextRecoverAt = this.progress.nextStaminaRecoverAt || 0;
    if (stamina >= maxStamina || !nextRecoverAt || now < nextRecoverAt) {
      return false;
    }

    const recoverCount = Math.min(
      maxStamina - stamina,
      1 + Math.floor((now - nextRecoverAt) / STAMINA_RECOVER_INTERVAL_MS)
    );
    if (recoverCount <= 0) {
      return false;
    }

    const nextProgress = {
      ...this.progress,
      stamina: stamina + recoverCount,
      nextStaminaRecoverAt: stamina + recoverCount >= maxStamina
        ? null
        : nextRecoverAt + recoverCount * STAMINA_RECOVER_INTERVAL_MS,
      updatedAt: now
    };
    this.progress = this.progressStore.saveRemote(nextProgress);
    return true;
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
    this.refreshStaminaWhenDue(now);

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
        player: this.player,
        toast: this.toast,
        settingsOpen: this.settingsOpen,
        coinAdDialogOpen: this.coinAdDialogOpen,
        staminaShopOpen: this.staminaShopOpen,
        staminaExchangeConfirm: this.staminaExchangeConfirm,
        settings: {
          musicEnabled: this.musicEnabled,
          sfxEnabled: this.sfxEnabled
        },
        unlimitedStamina: this.unlimitedStamina
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
        pageCount: Math.ceil(this.levelConfigs.length / 20),
        progress: this.progress,
        toast: this.toast,
        staminaShopOpen: this.staminaShopOpen,
        staminaExchangeConfirm: this.staminaExchangeConfirm,
        unlimitedStamina: this.unlimitedStamina
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
    const coinX = this.unlimitedStamina ? staminaX : staminaX + staminaW + hudGap;
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
    const pageCount = Math.ceil(this.levelConfigs.length / pageSize);
    this.levelPage = Math.max(0, Math.min(this.levelPage, pageCount - 1));
    const pageLevels = this.levelConfigs.slice(this.levelPage * pageSize, (this.levelPage + 1) * pageSize);

    this.levelCards = pageLevels.map((level, index) => {
      const progress = this.progress;
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
      retry: { x: this.width * 0.13, y: this.height * 0.28 + 208, w: this.width * 0.22, h: 42 },
      exit: { x: this.width * 0.39, y: this.height * 0.28 + 208, w: this.width * 0.22, h: 42 },
      next: { x: this.width * 0.65, y: this.height * 0.28 + 208, w: this.width * 0.22, h: 42 }
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

    const emptySlots = this.emptySlotsForLevel(level);
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
        this.ensureLevelConfigs();
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
        if (this.unlimitedStamina) {
          return;
        }
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
        this.levelPage = Math.min(Math.ceil(this.levelConfigs.length / 20) - 1, this.levelPage + 1);
        return;
      }

      if (hit(this.buttons.staminaPlus, x, y)) {
        if (this.unlimitedStamina) {
          return;
        }
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
      this.saveAudioSettings();
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
      this.saveAudioSettings();
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

  showToolRewardAd(type) {
    this.adService.showRewardedVideo("tool").then((result) => {
      if (result.status === "completed") {
        return Promise.resolve(this.rewardService.grantAdReward({
          type: "tool",
          toolType: type
        })).then((reward) => {
          this.progress = reward.progress;
          this.syncToolCountsFromProgress(reward.progress);
          this.showToast(reward.toastText);
        });
      }
      this.showAdFailureToast(result.status);
      return null;
    }).catch((error) => {
      this.showToast("\u6b21\u6570\u589e\u52a0\u5931\u8d25\uff0c\u8bf7\u91cd\u8bd5");
      if (typeof console !== "undefined" && console.warn) {
        console.warn("grant tool reward failed", error);
      }
    }).finally(() => {
      this.resumeGame();
    });
  }

  handleResultTouch(x, y) {
    if (hit(this.buttons.retry, x, y)) {
      this.playSfx("click");
      this.startLevel(this.game.level.levelId);
      return;
    }

    if (hit(this.buttons.exit, x, y)) {
      this.playSfx("click");
      this.scene = "levels";
      this.assets.stopBgm("gameBgm");
      this.playBgm("allBgm");
      return;
    }

    if (hit(this.buttons.next, x, y)) {
      this.playSfx("click");
      if (this.game.lastResult && this.game.lastResult.success) {
        this.startLevel(Math.min(this.game.level.levelId + 1, this.levelConfigs.length));
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
      this.saveAudioSettings();
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
      this.saveAudioSettings();
      this.playSfx("click");
      return;
    }

    if (hit(this.buttons.settingsClose, x, y)) {
      this.playSfx("click");
      this.settingsOpen = false;
    }
  }

  saveAudioSettings() {
    this.audioSettings = {
      musicEnabled: !!this.musicEnabled,
      sfxEnabled: !!this.sfxEnabled
    };
    this.progressStore.saveSettings(this.audioSettings);
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
      this.openStaminaExchangeConfirm(99, 1, "stamina_1_by_coins");
      return;
    }

    if (hit(this.buttons.staminaBuy3, x, y)) {
      this.openStaminaExchangeConfirm(266, 3, "stamina_3_by_coins");
      return;
    }

    if (hit(this.buttons.staminaBuy5, x, y)) {
      this.openStaminaExchangeConfirm(388, 5, "stamina_5_by_coins");
      return;
    }

    if (hit(this.buttons.staminaShopClose, x, y)) {
      this.playSfx("click");
      this.staminaShopOpen = false;
    }
  }

  openStaminaExchangeConfirm(cost, amount, productKey) {
    this.playSfx("click");
    this.staminaExchangeConfirm = { cost, amount, productKey };
  }

  handleStaminaExchangeConfirmTouch(x, y) {
    if (hit(this.buttons.staminaExchangeCancel, x, y)) {
      this.playSfx("click");
      this.staminaExchangeConfirm = null;
      return;
    }

    if (hit(this.buttons.staminaExchangeConfirm, x, y)) {
      this.playSfx("click");
      const { cost, amount, productKey } = this.staminaExchangeConfirm;
      const result = this.rewardService.exchangeCoinsForStamina(cost, amount, productKey);
      Promise.resolve(result).then((data) => {
        this.progress = data.progress;
        this.staminaExchangeConfirm = null;
        this.staminaShopOpen = false;
        this.showToast(data.toastText);
      });
    }
  }

  showCoinRewardAd() {
    this.adService.showRewardedVideo("coin").then((result) => {
      if (result.status === "completed") {
        const reward = this.rewardService.grantAdReward({ type: "coin", amount: 100 });
        this.progress = reward.progress;
        this.showToast(reward.toastText);
      } else {
        this.showAdFailureToast(result.status);
      }
    });
  }

  showStaminaRewardAd() {
    this.adService.showRewardedVideo("stamina").then((result) => {
      if (result.status === "completed") {
        const reward = this.rewardService.grantAdReward({ type: "stamina", amount: 3 });
        this.progress = reward.progress;
        this.showToast(reward.toastText);
      } else {
        this.showAdFailureToast(result.status);
      }
    });
  }

  startLevel(levelId) {
    if (this.startingLevel) {
      return;
    }

    this.startingLevel = true;
    this.showToast("\u6b63\u5728\u8fdb\u5165\u5173\u5361");
    const auth = this.progressStore.loadAuth();
    const session = auth && auth.token ? Promise.resolve() : (this.sessionPromise || this.syncSession());
    Promise.all([session, this.ensureLevelConfigs()])
      .then(() => this.apiClient.startLevel(levelId))
      .then((progress) => {
        if (progress) {
          this.progress = progress;
          this.syncToolCountsFromProgress(progress);
        }
        this.enterLevel(levelId);
      })
      .catch((error) => {
        const message = error && error.message ? error.message : "";
        if (message.includes("insufficient stamina")) {
          this.scene = this.scene === "levels" ? "levels" : "home";
          this.staminaShopOpen = true;
          this.showToast("\u4f53\u529b\u4e0d\u8db3");
        } else if (message.includes("missing auth token")) {
          this.showToast("\u767b\u5f55\u5931\u8d25\uff0c\u8bf7\u91cd\u8bd5");
        } else {
          this.showToast("\u7f51\u7edc\u5f02\u5e38\uff0c\u8bf7\u91cd\u8bd5");
        }
        if (typeof console !== "undefined" && console.warn) {
          console.warn("start level failed", error);
        }
      })
      .finally(() => {
        this.startingLevel = false;
      });
  }

  enterLevel(levelId) {
    this.scene = "game";
    this.game = new MatchGame(levelId, this.levelConfigs);
    this.gamePaused = false;
    this.pauseStartedAt = 0;
    this.toolAdDialog = null;
    this.previewEndAt = Date.now() + this.game.level.initialPreviewMs;
    this.mismatchTimerAt = 0;
    this.hintIndexes = [];
    this.removedAnimations = [];
    this.previewAgainIndexes = [];
    this.previewAgainEndAt = 0;
    this.syncToolCountsFromProgress(this.progress);
    this.assets.stopBgm("allBgm");
    this.playBgm("gameBgm");
  }

  emptySlotsForLevel(level) {
    const explicitSlots = Array.isArray(level.emptySlots) ? level.emptySlots : [];
    const totalSlots = level.rows * level.cols;
    if (totalSlots % 2 === 0) {
      return explicitSlots;
    }
    const centerSlot = Math.floor(totalSlots / 2);
    return explicitSlots.includes(centerSlot) ? explicitSlots : [...explicitSlots, centerSlot];
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
    if (!this.game.canUseHint()) {
      this.playSfx("wrong");
      return;
    }
    if (this.progress.hints <= 0) {
      this.openToolAdDialog("hint", "\u63d0\u793a");
      return;
    }

    this.consumeTool("hint", "\u63d0\u793a").then((success) => {
      if (!success) {
        return;
      }
      this.hintIndexes = this.game.useHint();
      this.hintEndAt = Date.now() + (this.game.level.hintHighlightMs || 1300);
      this.playSfx(this.hintIndexes.length > 0 ? "right" : "wrong");
    });
  }

  previewAgain() {
    if (this.previewAgainIndexes.length > 0) {
      this.playSfx("wrong");
      return;
    }

    if (!this.game.canRevealUnmatched()) {
      this.playSfx("wrong");
      return;
    }

    if (this.previewAgainCount <= 0) {
      this.openToolAdDialog("previewAgain", "\u518d\u770b\u4e00\u6b21");
      return;
    }

    this.consumeTool("previewAgain", "\u518d\u770b\u4e00\u6b21").then((success) => {
      if (!success) {
        return;
      }
      const indexes = this.game.revealUnmatched();
      if (indexes.length === 0) {
        this.playSfx("wrong");
        return;
      }
      this.previewAgainIndexes = indexes;
      this.previewAgainEndAt = Date.now() + 2000;
      this.playSfx("right");
    });
  }

  removeOnePair() {
    if (!this.game.canRemovePair()) {
      this.playSfx("wrong");
      return;
    }

    if (this.removePairCount <= 0) {
      this.openToolAdDialog("removePair", "\u6d88\u9664\u4e00\u5bf9");
      return;
    }

    this.consumeTool("removePair", "\u6d88\u9664\u4e00\u5bf9").then((success) => {
      if (!success) {
        return;
      }
      const outcome = this.game.removeOnePair();
      if (outcome.type === "ignored") {
        this.playSfx("wrong");
        return;
      }
      this.playSfx("right");
      if (outcome.type === "level_complete") {
        this.onLevelEnded(outcome.result);
      }
    });
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

  showToast(text) {
    if (!text) {
      return;
    }

    this.toast = {
      text,
      until: Date.now() + 2000
    };
  }

  showAdFailureToast(status) {
    if (status === "closed") {
      return;
    }

    this.showToast(status === "unconfigured" ? "\u5e7f\u544a\u6682\u672a\u914d\u7f6e" : "\u5e7f\u544a\u6682\u65f6\u65e0\u6cd5\u64ad\u653e");
  }

  onLevelEnded(result) {
    if (!result || result._saved) {
      return;
    }

    result._saved = true;
    this.apiClient.submitLevelResult({
      levelId: result.levelId,
      success: result.success,
      reason: result.reason,
      steps: result.steps,
      mismatchCount: result.mismatchCount,
      elapsedMs: result.elapsedMs,
      stars: result.stars,
      coinsEarned: result.coinsEarned,
      usedHints: result.usedHints
    })
      .then((progress) => {
        this.progress = progress;
        this.syncToolCountsFromProgress(progress);
      })
      .catch((error) => {
        result._saved = false;
        this.showToast("\u6210\u7ee9\u4fdd\u5b58\u5931\u8d25\uff0c\u8bf7\u68c0\u67e5\u7f51\u7edc");
        if (typeof console !== "undefined" && console.warn) {
          console.warn("submit level result failed", error);
        }
      });
  }
}

function hit(rect, x, y) {
  return rect && x >= rect.x && x <= rect.x + rect.w && y >= rect.y && y <= rect.y + rect.h;
}

module.exports = {
  GameApp
};
