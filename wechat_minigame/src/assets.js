class AssetLoader {
  constructor(canvas) {
    this.canvas = canvas;
    this.images = {};
    this.audio = {};
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
      const audio = wx.createInnerAudioContext();
      audio.src = src;
      audio.obeyMuteSwitch = false;
      this.audio[key] = audio;
    }
  }

  image(key) {
    return this.images[key] || null;
  }

  play(key) {
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
    }
  }

  bgm(key) {
    const audio = this.audio[key];
    if (!audio) {
      return;
    }

    audio.loop = true;
    try {
      audio.play();
    } catch (error) {
      console.warn(`BGM failed: ${key}`, error);
    }
  }

  stopBgm(key) {
    const audio = this.audio[key];
    if (audio) {
      audio.stop();
    }
  }
}

module.exports = {
  AssetLoader
};
