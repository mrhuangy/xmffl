class AssetLoader {
  constructor(canvas) {
    this.canvas = canvas;
    this.images = {};
    this.audio = {};
    this.audioSources = {};
    this.audioReady = {};
    this.audioPendingPlay = {};
    this.audioPreparing = {};
    this.audioRetryCount = {};
    this.audioLoopWanted = {};
  }

  loadImages(map, onProgress) {
    const entries = Object.entries(map);
    let loaded = 0;
    const tasks = entries.map(([key, src]) =>
      this.loadImage(key, src).then((image) => {
        loaded += 1;
        if (onProgress) {
          onProgress(loaded, entries.length, key);
        }
        return image;
      })
    );
    return Promise.all(tasks);
  }

  loadImage(key, src) {
    if (this.images[key]) {
      return Promise.resolve(this.images[key]);
    }

    return new Promise((resolve) => {
      const image = this.createImage();
      image.onload = () => {
        this.images[key] = image;
        resolve(image);
      };
      image.onerror = () => {
        console.warn(`Image failed to load: ${src}`);
        resolve(null);
      };
      image.src = src;
    });
  }

  createImage() {
    if (typeof wx.createImage === "function") {
      return wx.createImage();
    }

    if (this.canvas && typeof this.canvas.createImage === "function") {
      return this.canvas.createImage();
    }

    if (typeof Image !== "undefined") {
      return new Image();
    }

    throw new Error("No image factory is available in this WeChat runtime.");
  }

  loadAudio(map) {
    for (const [key, src] of Object.entries(map)) {
      this.audioSources[key] = src;
    }
  }

  prepareAudio(key) {
    const src = this.audioSources[key];
    if (!src || this.audio[key] || this.audioPreparing[key]) {
      return;
    }

    this.audioPreparing[key] = true;
    this.resolveAudioSource(key, src)
      .then((resolvedSrc) => {
        this.audioPreparing[key] = false;
        this.createAudio(key, resolvedSrc);
      })
      .catch((error) => {
        this.audioPreparing[key] = false;
        console.warn(`Audio failed to prepare: ${src}`, error);
        this.retryAudio(key);
      });
  }

  resolveAudioSource(key, src) {
    if (!/^https?:\/\//.test(src) || !wx.downloadFile) {
      return Promise.resolve(src);
    }

    const cached = this.getCachedAudioPath(key, src);
    if (cached) {
      return Promise.resolve(cached);
    }

    return this.downloadAudio(key, src);
  }

  getCachedAudioPath(key, src) {
    try {
      const cache = wx.getStorageSync("fpxxl.audio_cache") || {};
      const item = cache[key];
      if (item && item.src === src && item.path) {
        return item.path;
      }
    } catch (error) {
      console.warn("Audio cache read failed", error);
    }
    return "";
  }

  setCachedAudioPath(key, src, path) {
    try {
      const cache = wx.getStorageSync("fpxxl.audio_cache") || {};
      cache[key] = { src, path };
      wx.setStorageSync("fpxxl.audio_cache", cache);
    } catch (error) {
      console.warn("Audio cache write failed", error);
    }
  }

  clearCachedAudioPath(key) {
    try {
      const cache = wx.getStorageSync("fpxxl.audio_cache") || {};
      if (cache[key]) {
        delete cache[key];
        wx.setStorageSync("fpxxl.audio_cache", cache);
      }
    } catch (error) {
      console.warn("Audio cache clear failed", error);
    }
  }

  downloadAudio(key, src) {
    return new Promise((resolve, reject) => {
      wx.downloadFile({
        url: src,
        timeout: 30000,
        success: (res) => {
          if (res.statusCode && res.statusCode >= 400) {
            reject(new Error(`HTTP ${res.statusCode}`));
            return;
          }
          const tempFilePath = res.tempFilePath;
          if (!wx.saveFile) {
            resolve(tempFilePath);
            return;
          }
          wx.saveFile({
            tempFilePath,
            success: (saveRes) => {
              const savedPath = saveRes.savedFilePath || tempFilePath;
              this.setCachedAudioPath(key, src, savedPath);
              resolve(savedPath);
            },
            fail: () => resolve(tempFilePath)
          });
        },
        fail: reject
      });
    });
  }

  createAudio(key, src) {
    if (this.audio[key] && typeof this.audio[key].destroy === "function") {
      this.audio[key].destroy();
    }

    const audio = wx.createInnerAudioContext();
    audio.obeyMuteSwitch = false;
    audio.loop = !!this.audioLoopWanted[key];
    audio.onCanplay(() => {
      this.audioReady[key] = true;
      if (this.audioPendingPlay[key]) {
        this.audioPendingPlay[key] = false;
        this.playAudioContext(key, true);
      }
    });
    audio.onError((error) => {
      this.audioReady[key] = false;
      this.clearCachedAudioPath(key);
      console.warn(`Audio failed to load: ${src}`, error);
      this.retryAudio(key);
    });
    audio.src = src;
    this.audio[key] = audio;
  }

  image(key) {
    return this.images[key] || null;
  }

  play(key) {
    const audio = this.audio[key];
    if (!audio) {
      this.audioPendingPlay[key] = true;
      this.prepareAudio(key);
      return;
    }

    if (!this.audioReady[key]) {
      this.audioPendingPlay[key] = true;
    }

    this.playAudioContext(key, false);
  }

  playAudioContext(key, isRetry) {
    const audio = this.audio[key];
    if (!audio) {
      return;
    }

    try {
      audio.stop();
      audio.currentTime = 0;
      audio.play();
    } catch (error) {
      console.warn(`Audio failed: ${key}`, error);
      if (!isRetry) {
        this.audioPendingPlay[key] = true;
      }
    }
  }

  bgm(key) {
    const audio = this.audio[key];
    if (!audio) {
      this.audioPendingPlay[key] = true;
      this.audioLoopWanted[key] = true;
      this.prepareAudio(key);
      return;
    }

    audio.loop = true;
    this.audioLoopWanted[key] = true;
    if (!this.audioReady[key]) {
      this.audioPendingPlay[key] = true;
    }

    try {
      audio.play();
    } catch (error) {
      console.warn(`BGM failed: ${key}`, error);
      this.audioPendingPlay[key] = true;
    }
  }

  retryAudio(key) {
    const retryCount = this.audioRetryCount[key] || 0;
    if (retryCount >= 2) {
      return;
    }

    this.audioRetryCount[key] = retryCount + 1;
    setTimeout(() => {
      delete this.audio[key];
      this.prepareAudio(key);
    }, 1200 * this.audioRetryCount[key]);
  }

  stopBgm(key) {
    const audio = this.audio[key];
    this.audioPendingPlay[key] = false;
    this.audioLoopWanted[key] = false;
    if (audio) {
      audio.loop = false;
      audio.stop();
    }
  }
}

module.exports = {
  AssetLoader
};
