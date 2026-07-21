const defaultAdUnitIds = {
  coin: "",
  stamina: "",
  tool: ""
};

class AdService {
  constructor(options = {}) {
    this.adUnitIds = {
      ...defaultAdUnitIds,
      ...(options.adUnitIds || {})
    };
    this.rewardedAds = {};
    this.pendingResolvers = {};
  }

  showRewardedVideo(placement) {
    const adUnitId = this.adUnitIds[placement];
    if (!adUnitId || typeof wx.createRewardedVideoAd !== "function") {
      return Promise.resolve({ status: "unconfigured", placement });
    }

    const ad = this.getRewardedAd(placement, adUnitId);
    return new Promise((resolve) => {
      this.pendingResolvers[placement] = resolve;
      ad.show().catch(() => {
        ad.load()
          .then(() => ad.show())
          .catch(() => this.resolvePlacement(placement, { status: "failed", placement }));
      });
    });
  }

  getRewardedAd(placement, adUnitId) {
    if (this.rewardedAds[placement]) {
      return this.rewardedAds[placement];
    }

    const ad = wx.createRewardedVideoAd({ adUnitId });
    ad.onClose((result) => {
      this.resolvePlacement(placement, {
        status: result && result.isEnded ? "completed" : "closed",
        placement
      });
    });
    ad.onError(() => {
      this.resolvePlacement(placement, { status: "failed", placement });
    });

    this.rewardedAds[placement] = ad;
    return ad;
  }

  resolvePlacement(placement, result) {
    const resolve = this.pendingResolvers[placement];
    if (!resolve) {
      return;
    }

    delete this.pendingResolvers[placement];
    resolve(result);
  }
}

module.exports = {
  AdService
};
