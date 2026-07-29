// const API_BASE = "https://fpxxl-gameapi.gaintmonster.cn/api/v1";
// const API_BASE = "http://127.0.0.1:8090/api/v1";
const API_BASE = "http://minigame-api.fpxxl.local/api/v1";
const REQUEST_TIMEOUT_MS = 30000;
const REQUEST_RETRY_COUNT = 1;

class ApiClient {
  constructor(progressStore) {
    this.progressStore = progressStore;
  }

  login() {
    return wxLogin()
      .then((loginResult) => request({
        url: `${API_BASE}/auth/login`,
        method: "POST",
        data: {
          code: loginResult.code,
          nickname: "",
          avatarUrl: ""
        }
      }))
      .then((data) => {
        if (data && data.token) {
          this.progressStore.saveAuth({
            token: data.token,
            player: data.player || null
          });
        }
        if (data && data.progress) {
          this.progressStore.saveRemote(data.progress);
        }
        return data;
      });
  }

  fetchProgress() {
    const auth = this.progressStore.loadAuth();
    if (!auth || !auth.token) {
      return Promise.reject(new Error("missing auth token"));
    }
    return request({
      url: `${API_BASE}/player/progress`,
      method: "GET",
      header: {
        Authorization: `Bearer ${auth.token}`
      }
    }).then((progress) => this.progressStore.saveRemote(progress));
  }

  fetchLevels() {
    return request({
      url: `${API_BASE}/config/levels`,
      method: "GET"
    }).then((data) => (data && Array.isArray(data.levels) ? data.levels : []));
  }

  fetchInitConfig() {
    return request({
      url: `${API_BASE}/config/init`,
      method: "GET"
    });
  }

  startLevel(levelId) {
    const auth = this.progressStore.loadAuth();
    if (!auth || !auth.token) {
      return Promise.reject(new Error("missing auth token"));
    }
    return request({
      url: `${API_BASE}/levels/start`,
      method: "POST",
      data: {
        levelId
      },
      header: {
        Authorization: `Bearer ${auth.token}`
      }
    }).then((data) => {
      if (data && data.progress) {
        return this.progressStore.saveRemote(data.progress);
      }
      return this.progressStore.load();
    });
  }

  submitLevelResult(result) {
    const auth = this.progressStore.loadAuth();
    if (!auth || !auth.token) {
      return Promise.reject(new Error("missing auth token"));
    }
    return request({
      url: `${API_BASE}/levels/results`,
      method: "POST",
      data: result,
      header: {
        Authorization: `Bearer ${auth.token}`
      }
    }).then((data) => {
      if (data && data.progress) {
        return {
          progress: this.progressStore.saveRemote(data.progress),
          rewards: data.rewards || {}
        };
      }
      return {
        progress: this.progressStore.load(),
        rewards: {}
      };
    });
  }

  changeToolCount(toolType, delta) {
    const auth = this.progressStore.loadAuth();
    if (!auth || !auth.token) {
      return Promise.reject(new Error("missing auth token"));
    }
    return request({
      url: `${API_BASE}/tools/change`,
      method: "POST",
      data: {
        toolType: normalizeToolType(toolType),
        delta
      },
      header: {
        Authorization: `Bearer ${auth.token}`
      }
    }).then((data) => {
      if (data && data.progress) {
        return this.progressStore.saveRemote(data.progress);
      }
      return this.progressStore.load();
    });
  }

  purchaseToolCount(toolType) {
    const auth = this.progressStore.loadAuth();
    if (!auth || !auth.token) {
      return Promise.reject(new Error("missing auth token"));
    }
    return request({
      url: `${API_BASE}/tools/purchase`,
      method: "POST",
      data: {
        toolType: normalizeToolType(toolType)
      },
      header: {
        Authorization: `Bearer ${auth.token}`
      }
    }).then((data) => {
      if (data && data.progress) {
        return this.progressStore.saveRemote(data.progress);
      }
      return this.progressStore.load();
    });
  }

  purchaseProduct(productKey) {
    const auth = this.progressStore.loadAuth();
    if (!auth || !auth.token) {
      return Promise.reject(new Error("missing auth token"));
    }
    return request({
      url: `${API_BASE}/shop/purchase`,
      method: "POST",
      data: {
        productKey
      },
      header: {
        Authorization: `Bearer ${auth.token}`
      }
    }).then((data) => {
      if (data && data.progress) {
        return this.progressStore.saveRemote(data.progress);
      }
      return this.progressStore.load();
    });
  }
}

function normalizeToolType(toolType) {
  if (toolType === "previewAgain") {
    return "preview_again";
  }
  if (toolType === "removePair") {
    return "remove_pair";
  }
  return toolType;
}

function wxLogin() {
  return new Promise((resolve, reject) => {
    if (!wx.login) {
      resolve({ code: `local_${Date.now()}` });
      return;
    }
    wx.login({
      success: (res) => {
        if (res && res.code) {
          resolve(res);
        } else {
          reject(new Error("wx.login returned empty code"));
        }
      },
      fail: reject
    });
  });
}

function request(options) {
  return requestWithRetry(options, 0);
}

function requestWithRetry(options, attempt) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: options.url,
      method: options.method || "GET",
      data: options.data,
      timeout: REQUEST_TIMEOUT_MS,
      header: {
        "Content-Type": "application/json",
        ...(options.header || {})
      },
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data);
          return;
        }
        const message = res.data && res.data.error ? res.data.error : `HTTP ${res.statusCode}`;
        reject(new Error(message));
      },
      fail: (error) => {
        if (shouldRetry(error) && attempt < REQUEST_RETRY_COUNT) {
          setTimeout(() => {
            requestWithRetry(options, attempt + 1).then(resolve).catch(reject);
          }, 800);
          return;
        }
        reject(error);
      }
    });
  });
}

function shouldRetry(error) {
  const message = error && error.errMsg ? error.errMsg : "";
  return message.includes("timeout") || message.includes("fail");
}

module.exports = {
  ApiClient
};
