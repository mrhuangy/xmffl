package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type adminClaims struct {
	Subject     uint64 `json:"sub"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	ExpiresAt   int64  `json:"exp"`
}

type claimsContextKey struct{}

func withAdminClaims(ctx context.Context, claims adminClaims) context.Context {
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

func (h *Handler) requiresAdmin(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/admin/") || r.URL.Path == "/api/auth/me" ||
		(r.URL.Path == "/api/config/levels" && r.URL.Query().Get("includeDisabled") == "true")
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &request); err != nil || strings.TrimSpace(request.Username) == "" || request.Password == "" {
		writeBadRequest(w, "请输入用户名和密码")
		return
	}
	admin, err := h.repo.FindAdminByUsername(r.Context(), strings.TrimSpace(request.Username))
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	if admin.Status == "disabled" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "账号已被禁用"})
		return
	}
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		writeJSON(w, http.StatusLocked, map[string]string{"error": "登录失败次数过多，请稍后再试"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(request.Password)) != nil {
		var lockedUntil *time.Time
		if admin.FailedLoginAttempts+1 >= 5 {
			value := time.Now().Add(15 * time.Minute)
			lockedUntil = &value
		}
		_ = h.repo.RecordAdminLoginFailure(r.Context(), admin.ID, lockedUntil)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
		return
	}
	if err := h.repo.RecordAdminLoginSuccess(r.Context(), admin.ID, clientIP(r)); err != nil {
		writeError(w, err)
		return
	}
	claims := adminClaims{Subject: admin.ID, Username: admin.Username, DisplayName: admin.DisplayName, Role: admin.Role, ExpiresAt: time.Now().Add(8 * time.Hour).Unix()}
	token, err := h.signToken(claims)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "expiresAt": claims.ExpiresAt, "user": claims})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(claimsContextKey{}).(adminClaims)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "未登录"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": claims})
}

func (h *Handler) signToken(claims adminClaims) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := encodeSegment(header) + "." + encodeSegment(payload)
	mac := hmac.New(sha256.New, h.jwtSecret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + encodeSegment(mac.Sum(nil)), nil
}

func (h *Handler) authenticate(r *http.Request) (adminClaims, error) {
	var claims adminClaims
	value := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	parts := strings.Split(value, ".")
	if len(parts) != 3 || value == r.Header.Get("Authorization") {
		return claims, errors.New("invalid token")
	}
	mac := hmac.New(sha256.New, h.jwtSecret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return claims, errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(payload, &claims) != nil || claims.ExpiresAt <= time.Now().Unix() {
		return claims, errors.New("expired token")
	}
	return claims, nil
}

func encodeSegment(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func clientIP(r *http.Request) string {
	if value := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); value != "" {
		return value
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
