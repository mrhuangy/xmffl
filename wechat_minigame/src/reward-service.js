class RewardService {
  constructor(progressStore, apiClient) {
    this.progressStore = progressStore;
    this.apiClient = apiClient || null;
  }

  loadProgress() {
    return this.progressStore.load();
  }

  applyLevelResult(result) {
    return this.progressStore.applyResult(result);
  }

  exchangeCoinsForStamina(cost, amount, productKey) {
    const targetProductKey = productKey || this.staminaProductKey(cost, amount);
    if (this.apiClient && this.apiClient.purchaseProduct && targetProductKey) {
      return this.apiClient.purchaseProduct(targetProductKey).then((progress) => ({
        success: true,
        progress,
        toastText: `\u83b7\u5f97${amount}\u4f53\u529b`
      })).catch((error) => {
        const message = error && error.message ? error.message : "";
        return {
          success: false,
          progress: this.loadProgress(),
          toastText: message.includes("insufficient coins") ? "\u91d1\u5e01\u4e0d\u8db3" : "\u5151\u6362\u5931\u8d25"
        };
      });
    }

    const result = this.progressStore.exchangeCoinsForStamina(cost, amount);
    return Promise.resolve({
      ...result,
      toastText: result.success ? `\u83b7\u5f97${amount}\u4f53\u529b` : "\u91d1\u5e01\u4e0d\u8db3"
    });
  }

  grantAdReward(reward) {
    if (!reward || !reward.type) {
      return {
        progress: this.loadProgress(),
        toastText: ""
      };
    }

    if (reward.type === "coin") {
      const amount = reward.amount || 100;
      return {
        progress: this.progressStore.addCoins(amount),
        toastText: `\u83b7\u5f97${amount}\u91d1\u5e01`
      };
    }

    if (reward.type === "stamina") {
      const amount = reward.amount || 3;
      return {
        progress: this.progressStore.addStamina(amount),
        toastText: `\u83b7\u5f97${amount}\u4f53\u529b`
      };
    }

    if (reward.type === "tool" && reward.toolType === "hint") {
      return this.changeToolReward(reward.toolType, 1);
    }

    if (reward.type === "tool") {
      return this.changeToolReward(reward.toolType, 1);
    }

    return {
      progress: this.loadProgress(),
      toastText: ""
    };
  }

  changeToolReward(toolType, amount) {
    if (this.apiClient && this.apiClient.changeToolCount) {
      return this.apiClient.changeToolCount(toolType, amount).then((progress) => ({
        progress,
        toastText: "\u6b21\u6570+1",
        toolType
      }));
    }

    const progress = this.progressStore.load();
    if (toolType === "hint") {
      return {
        progress: this.progressStore.addHints(amount),
        toastText: "\u6b21\u6570+1",
        toolType
      };
    }

    if (toolType === "previewAgain") {
      progress.previewAgainCount += amount;
      this.progressStore.save(progress);
      return {
        progress,
        toastText: "\u6b21\u6570+1",
        toolType
      };
    }

    if (toolType === "removePair") {
      progress.removePairCount += amount;
      this.progressStore.save(progress);
      return {
        progress,
        toastText: "\u6b21\u6570+1",
        toolType
      };
    }

    return {
      progress: this.loadProgress(),
      toastText: ""
    };
  }

  exchangeCoinsForTool(toolType) {
    if (this.apiClient && this.apiClient.purchaseToolCount) {
      return this.apiClient.purchaseToolCount(toolType).then((progress) => ({
        success: true,
        progress,
        toastText: "\u6b21\u6570+1",
        toolType
      })).catch((error) => {
        const message = error && error.message ? error.message : "";
        return {
          success: false,
          progress: this.loadProgress(),
          toastText: message.includes("insufficient coins") ? "\u91d1\u5e01\u4e0d\u8db3" : "\u5151\u6362\u5931\u8d25",
          toolType
        };
      });
    }

    return Promise.resolve({
      success: false,
      progress: this.loadProgress(),
      toastText: "\u5151\u6362\u5931\u8d25",
      toolType
    });
  }

  staminaProductKey(cost, amount) {
    if (amount === 1 && cost === 99) {
      return "stamina_1_by_coins";
    }
    if (amount === 3 && cost === 266) {
      return "stamina_3_by_coins";
    }
    if (amount === 5 && cost === 388) {
      return "stamina_5_by_coins";
    }
    return "";
  }
}

module.exports = {
  RewardService
};
