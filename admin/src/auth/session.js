const TOKEN_KEY = "fpxxl_admin_token";
const USER_KEY = "fpxxl_admin_user";
const EXPIRES_KEY = "fpxxl_admin_expires_at";

export function getToken() { return localStorage.getItem(TOKEN_KEY) || ""; }
export function getUser() {
  try { return JSON.parse(localStorage.getItem(USER_KEY) || "null"); } catch { return null; }
}
export function isAuthenticated() {
  return Boolean(getToken()) && Number(localStorage.getItem(EXPIRES_KEY) || 0) > Math.floor(Date.now() / 1000);
}
export function saveSession(data) {
  localStorage.setItem(TOKEN_KEY, data.token);
  localStorage.setItem(USER_KEY, JSON.stringify(data.user));
  localStorage.setItem(EXPIRES_KEY, String(data.expiresAt));
}
export function clearSession() {
  localStorage.removeItem(TOKEN_KEY); localStorage.removeItem(USER_KEY); localStorage.removeItem(EXPIRES_KEY);
}
