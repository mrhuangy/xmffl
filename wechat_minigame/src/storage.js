const STORAGE_KEY = "fpxxl.player_progress";
const AUTH_KEY = "fpxxl.auth";
const SETTINGS_KEY = "fpxxl.settings";

function defaultProgress() {
  return {
    version: 1,
    currentLevel: 1,
    coins: 0,
    hints: 3,
    stamina: 5,
    maxStamina: 5,
    nextStaminaRecoverAt: null,
    previewAgainCount: 3,
    removePairCount: 3,
    levelStars: {},
    completedLevels: [],
    updatedAt: Date.now()
  };
}

function normalizeProgress(progress) {
  const fallback = defaultProgress();
  const source = progress || {};
  return {
    ...fallback,
    ...source,
    version: 1,
    currentLevel: typeof source.currentLevel === "number" ? source.currentLevel : fallback.currentLevel,
    coins: typeof source.coins === "number" ? source.coins : fallback.coins,
    hints: typeof source.hints === "number" ? source.hints : fallback.hints,
    stamina: typeof source.stamina === "number" ? source.stamina : fallback.stamina,
    maxStamina: typeof source.maxStamina === "number" ? source.maxStamina : fallback.maxStamina,
    nextStaminaRecoverAt: typeof source.nextStaminaRecoverAt === "number" ? source.nextStaminaRecoverAt : null,
    previewAgainCount: typeof source.previewAgainCount === "number" ? source.previewAgainCount : fallback.previewAgainCount,
    removePairCount: typeof source.removePairCount === "number" ? source.removePairCount : fallback.removePairCount,
    levelStars: source.levelStars || {},
    completedLevels: source.completedLevels || [],
    updatedAt: typeof source.updatedAt === "number" ? source.updatedAt : fallback.updatedAt
  };
}

function defaultSettings() {
  return {
    musicEnabled: true,
    sfxEnabled: true,
    updatedAt: Date.now()
  };
}

function normalizeSettings(settings) {
  const fallback = defaultSettings();
  const source = settings || {};
  return {
    ...fallback,
    musicEnabled: typeof source.musicEnabled === "boolean" ? source.musicEnabled : fallback.musicEnabled,
    sfxEnabled: typeof source.sfxEnabled === "boolean" ? source.sfxEnabled : fallback.sfxEnabled,
    updatedAt: typeof source.updatedAt === "number" ? source.updatedAt : fallback.updatedAt
  };
}

class ProgressStore {
  load() {
    try {
      const raw = wx.getStorageSync(STORAGE_KEY);
      if (!raw) {
        return defaultProgress();
      }

      const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
      return normalizeProgress(parsed);
    } catch (error) {
      return defaultProgress();
    }
  }

  save(progress) {
    wx.setStorageSync(STORAGE_KEY, {
      ...normalizeProgress(progress),
      version: 1,
      updatedAt: Date.now()
    });
  }

  saveRemote(progress) {
    const normalized = normalizeProgress(progress);
    wx.setStorageSync(STORAGE_KEY, normalized);
    return normalized;
  }

  loadSettings() {
    try {
      const raw = wx.getStorageSync(SETTINGS_KEY);
      if (!raw) {
        return defaultSettings();
      }

      const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
      return normalizeSettings(parsed);
    } catch (error) {
      return defaultSettings();
    }
  }

  saveSettings(settings) {
    wx.setStorageSync(SETTINGS_KEY, {
      ...normalizeSettings(settings),
      updatedAt: Date.now()
    });
  }

  loadAuth() {
    try {
      return wx.getStorageSync(AUTH_KEY) || null;
    } catch (error) {
      return null;
    }
  }

  saveAuth(auth) {
    wx.setStorageSync(AUTH_KEY, {
      token: auth && auth.token ? auth.token : "",
      player: auth && auth.player ? auth.player : null,
      updatedAt: Date.now()
    });
  }

  applyResult(result) {
    const progress = this.load();

    if (result.success) {
      const levelKey = String(result.levelId);
      progress.levelStars[levelKey] = Math.max(progress.levelStars[levelKey] || 0, result.stars);
      progress.coins += result.coinsEarned;

      if (!progress.completedLevels.includes(result.levelId)) {
        progress.completedLevels.push(result.levelId);
      }

      progress.currentLevel = Math.max(progress.currentLevel, result.levelId + 1);
    }

    this.save(progress);
    return progress;
  }

  addHints(amount) {
    const progress = this.load();
    progress.hints += amount;
    this.save(progress);
    return progress;
  }

  addCoins(amount) {
    const progress = this.load();
    progress.coins += amount;
    this.save(progress);
    return progress;
  }

  addStamina(amount) {
    const progress = this.load();
    progress.stamina += amount;
    this.save(progress);
    return progress;
  }

  exchangeCoinsForStamina(cost, amount) {
    const progress = this.load();
    if (progress.coins < cost) {
      return { success: false, progress };
    }

    progress.coins -= cost;
    progress.stamina += amount;
    this.save(progress);
    return { success: true, progress };
  }
}

module.exports = {
  ProgressStore
};
