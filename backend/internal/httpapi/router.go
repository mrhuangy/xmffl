package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fpxxl/backend/internal/domain"
	"fpxxl/backend/internal/store"
)

type Handler struct {
	repo        store.Repository
	allowOrigin string
	jwtSecret   []byte
}

func NewRouter(repo store.Repository, allowOrigin, jwtSecret string) http.Handler {
	h := &Handler{repo: repo, allowOrigin: allowOrigin, jwtSecret: []byte(jwtSecret)}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("GET /api/auth/me", h.me)
	mux.HandleFunc("GET /api/config/levels", h.listLevels)
	mux.HandleFunc("PUT /api/admin/levels/", h.upsertLevel)
	mux.HandleFunc("GET /api/config/ads", h.getAdConfig)
	mux.HandleFunc("PUT /api/admin/config/ads", h.saveAdConfig)
	mux.HandleFunc("GET /api/admin/players", h.listAdminPlayers)
	mux.HandleFunc("GET /api/admin/players/", h.getAdminPlayer)
	mux.HandleFunc("GET /api/admin/users", h.listAdmins)
	mux.HandleFunc("POST /api/admin/users", h.createAdmin)
	mux.HandleFunc("GET /api/admin/users/", h.getAdmin)
	mux.HandleFunc("PUT /api/admin/users/", h.updateAdmin)
	mux.HandleFunc("DELETE /api/admin/users/", h.deleteAdmin)
	mux.HandleFunc("GET /api/admin/system-controls", h.listSystemControls)
	mux.HandleFunc("POST /api/admin/system-controls", h.createSystemControl)
	mux.HandleFunc("GET /api/admin/system-controls/", h.getSystemControl)
	mux.HandleFunc("PUT /api/admin/system-controls/", h.updateSystemControl)
	mux.HandleFunc("DELETE /api/admin/system-controls/", h.deleteSystemControl)
	mux.HandleFunc("GET /api/player/progress", h.getProgress)
	mux.HandleFunc("POST /api/player/progress", h.saveProgress)
	mux.HandleFunc("POST /api/player/level-results", h.saveLevelResult)
	mux.HandleFunc("POST /api/leaderboard/submit", h.submitLeaderboard)
	mux.HandleFunc("GET /api/leaderboard", h.listLeaderboard)
	mux.HandleFunc("POST /api/events/batch", h.saveEvents)

	return h.middleware(mux)
}

func (h *Handler) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", h.allowOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Openid")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if h.requiresAdmin(r) {
			claims, err := h.authenticate(r)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "登录状态已失效，请重新登录"})
				return
			}
			r = r.WithContext(withAdminClaims(r.Context(), claims))
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listLevels(w http.ResponseWriter, r *http.Request) {
	includeDisabled := r.URL.Query().Get("includeDisabled") == "true"
	levels, err := h.repo.ListLevels(r.Context(), includeDisabled)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"levels": levels})
}

func (h *Handler) upsertLevel(w http.ResponseWriter, r *http.Request) {
	levelID, err := pathInt(r.URL.Path, "/api/admin/levels/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	var level domain.LevelConfig
	if err := decodeJSON(r, &level); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	level.LevelID = levelID
	if err := validateLevel(level); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if err := h.repo.UpsertLevel(r.Context(), level); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *Handler) getAdConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.repo.GetAdConfig(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (h *Handler) saveAdConfig(w http.ResponseWriter, r *http.Request) {
	var cfg domain.AdFrequencyConfig
	if err := decodeJSON(r, &cfg); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if cfg.NoInterstitialBeforeLevel < 0 || cfg.InterstitialEveryLevels <= 0 || cfg.MaxInterstitialPerDay < 0 || cfg.MaxRevivePerLevel < 0 {
		writeBadRequest(w, "invalid ad frequency config")
		return
	}
	if err := h.repo.SaveAdConfig(r.Context(), cfg); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *Handler) listAdminPlayers(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	status := r.URL.Query().Get("status")
	if status != "" && status != "active" && status != "blocked" && status != "deleted" {
		writeBadRequest(w, "invalid player status")
		return
	}
	players, total, err := h.repo.ListAdminPlayers(r.Context(), strings.TrimSpace(r.URL.Query().Get("keyword")), status, (page-1)*pageSize, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"players": players, "total": total, "page": page, "pageSize": pageSize})
}

func (h *Handler) getAdminPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := pathUint64(r.URL.Path, "/api/admin/players/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	detail, err := h.repo.GetAdminPlayer(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) getProgress(w http.ResponseWriter, r *http.Request) {
	openID := openIDFromRequest(r)
	if openID == "" {
		writeBadRequest(w, "openId is required")
		return
	}
	progress, err := h.repo.GetProgress(r.Context(), openID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func (h *Handler) saveProgress(w http.ResponseWriter, r *http.Request) {
	var progress domain.PlayerProgress
	if err := decodeJSON(r, &progress); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if progress.OpenID == "" {
		progress.OpenID = openIDFromRequest(r)
	}
	if progress.OpenID == "" {
		writeBadRequest(w, "openId is required")
		return
	}
	if progress.LevelStars == nil {
		progress.LevelStars = map[string]int{}
	}
	if progress.CompletedLevels == nil {
		progress.CompletedLevels = []int{}
	}
	if err := h.repo.SaveProgress(r.Context(), progress); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (h *Handler) saveLevelResult(w http.ResponseWriter, r *http.Request) {
	var result domain.LevelResult
	if err := decodeJSON(r, &result); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if result.OpenID == "" {
		result.OpenID = openIDFromRequest(r)
	}
	if result.OpenID == "" || result.LevelID <= 0 {
		writeBadRequest(w, "openId and levelId are required")
		return
	}
	if err := h.repo.SaveLevelResult(r.Context(), result); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handler) submitLeaderboard(w http.ResponseWriter, r *http.Request) {
	var entry domain.LeaderboardEntry
	if err := decodeJSON(r, &entry); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if entry.OpenID == "" {
		entry.OpenID = openIDFromRequest(r)
	}
	if entry.OpenID == "" || entry.LevelID <= 0 {
		writeBadRequest(w, "openId and levelId are required")
		return
	}
	if err := h.repo.SubmitLeaderboard(r.Context(), entry); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handler) listLeaderboard(w http.ResponseWriter, r *http.Request) {
	levelID := queryInt(r, "levelId", 0)
	limit := queryInt(r, "limit", 50)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	entries, err := h.repo.ListLeaderboard(r.Context(), levelID, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

func (h *Handler) saveEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Events []domain.GameEvent `json:"events"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	openID := openIDFromRequest(r)
	for i := range req.Events {
		if req.Events[i].EventID == "" || req.Events[i].EventType == "" {
			writeBadRequest(w, "eventId and eventType are required")
			return
		}
		if req.Events[i].OpenID == "" {
			req.Events[i].OpenID = openID
		}
		if req.Events[i].CreatedAt.IsZero() {
			req.Events[i].CreatedAt = time.Now()
		}
	}
	if err := h.repo.SaveEvents(r.Context(), req.Events); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"accepted": len(req.Events)})
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

func openIDFromRequest(r *http.Request) string {
	if value := r.Header.Get("X-Openid"); value != "" {
		return value
	}
	return r.URL.Query().Get("openId")
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func pathInt(path, prefix string) (int, error) {
	value := strings.TrimPrefix(path, prefix)
	if value == "" || value == path {
		return 0, errors.New("missing id in path")
	}
	return strconv.Atoi(value)
}

func pathUint64(path, prefix string) (uint64, error) {
	value := strings.TrimPrefix(path, prefix)
	if value == "" || value == path {
		return 0, errors.New("missing id in path")
	}
	return strconv.ParseUint(value, 10, 64)
}

func validateLevel(level domain.LevelConfig) error {
	if level.LevelID <= 0 {
		return errors.New("levelId must be positive")
	}
	if level.Rows <= 0 || level.Cols <= 0 || level.PairCount <= 0 {
		return errors.New("rows, cols and pairCount must be positive")
	}
	if !validLevelGrid(level.Rows, level.Cols, level.PairCount) {
		return errors.New("rows * cols must equal pairCount * 2, or equal pairCount * 2 + 1 for center empty slot")
	}
	if level.Mode == "" {
		return errors.New("mode is required")
	}
	if level.ThemeID == "" {
		return errors.New("themeId is required")
	}
	return nil
}

func validLevelGrid(rows int, cols int, pairCount int) bool {
	totalSlots := rows * cols
	cardSlots := pairCount * 2
	return totalSlots == cardSlots || (totalSlots == cardSlots+1 && totalSlots%2 == 1)
}
