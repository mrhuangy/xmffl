const STORAGE_KEY = "fpxxl.player_progress";

function defaultProgress() {
  return {
    version: 1,
    currentLevel: 1,
    coins: 0,
    hints: 3,
    stamina: 5,
    maxStamina: 5,
    levelStars: {},
    completedLevels: [],
    updatedAt: Date.now()
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
      return {
        ...defaultProgress(),
        ...parsed,
        version: 1,
        stamina: typeof parsed.stamina === "number" ? parsed.stamina : defaultProgress().stamina,
        maxStamina: typeof parsed.maxStamina === "number" ? parsed.maxStamina : defaultProgress().maxStamina,
        levelStars: parsed.levelStars || {},
        completedLevels: parsed.completedLevels || []
      };
    } catch (error) {
      return defaultProgress();
    }
  }

  save(progress) {
    wx.setStorageSync(STORAGE_KEY, {
      ...progress,
      version: 1,
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

  spendHint() {
    const progress = this.load();
    if (progress.hints <= 0) {
      return false;
    }

    progress.hints -= 1;
    this.save(progress);
    return true;
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
