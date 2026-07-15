const API_BASE = import.meta.env.VITE_API_BASE || "/api";

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {})
    },
    ...options
  });

  const text = await response.text();
  const data = text ? JSON.parse(text) : null;
  if (!response.ok) {
    throw new Error(data?.error || `HTTP ${response.status}`);
  }
  return data;
}

export function fetchLevels() {
  return request("/config/levels?includeDisabled=true");
}

export function saveLevel(level) {
  return request(`/admin/levels/${level.levelId}`, {
    method: "PUT",
    body: JSON.stringify(level)
  });
}

export function fetchAdConfig() {
  return request("/config/ads");
}

export function saveAdConfig(config) {
  return request("/admin/config/ads", {
    method: "PUT",
    body: JSON.stringify(config)
  });
}
