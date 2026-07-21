const API_BASE = import.meta.env.VITE_API_BASE || "/api";
import { clearSession, getToken } from "../auth/session";

async function request(path, options = {}) {
  const token = getToken();
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {})
    },
    ...options
  });

  const text = await response.text();
  const data = text ? JSON.parse(text) : null;
  if (response.status === 401 && path !== "/auth/login") {
    clearSession();
    window.dispatchEvent(new CustomEvent("admin-session-expired"));
  }
  if (!response.ok) {
    throw new Error(data?.error || `HTTP ${response.status}`);
  }
  return data;
}

export function login(credentials) {
  return request("/auth/login", { method: "POST", body: JSON.stringify(credentials) });
}

export function fetchCurrentAdmin() { return request("/auth/me"); }

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

export function fetchPlayers(params = {}) {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== "" && value != null) query.set(key, value);
  });
  return request(`/admin/players?${query.toString()}`);
}

export function fetchPlayerDetail(id) {
  return request(`/admin/players/${id}`);
}

export function fetchAdmins() { return request("/admin/users"); }
export function createAdmin(data) { return request("/admin/users", { method: "POST", body: JSON.stringify(data) }); }
export function updateAdmin(id, data) { return request(`/admin/users/${id}`, { method: "PUT", body: JSON.stringify(data) }); }
export function deleteAdmin(id) { return request(`/admin/users/${id}`, { method: "DELETE" }); }
export function fetchSystemControls() { return request("/admin/system-controls"); }
export function createSystemControl(data) { return request("/admin/system-controls", { method: "POST", body: JSON.stringify(data) }); }
export function updateSystemControl(id, data) { return request(`/admin/system-controls/${id}`, { method: "PUT", body: JSON.stringify(data) }); }
export function deleteSystemControl(id) { return request(`/admin/system-controls/${id}`, { method: "DELETE" }); }
