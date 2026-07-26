package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fpxxl/minigame-api/internal/domain"
	"fpxxl/minigame-api/internal/repository"
	"fpxxl/minigame-api/internal/service"
)

type Handler struct {
	repo          repository.Repository
	playerService *service.PlayerService
	gameService   *service.GameService
	shopService   *service.ShopService
}

func New(repo repository.Repository, playerService *service.PlayerService, gameService *service.GameService, shopService *service.ShopService) *Handler {
	return &Handler{
		repo:          repo,
		playerService: playerService,
		gameService:   gameService,
		shopService:   shopService,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Code      string `json:"code" binding:"required"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	player, progress, token, err := h.playerService.Login(c.Request.Context(), req.Code, req.Nickname, req.AvatarURL)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"player":   player,
		"progress": domain.NewClientProgress(progress),
	})
}

func (h *Handler) Levels(c *gin.Context) {
	levels, err := h.repo.ListLevels(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"levels": levels})
}

func (h *Handler) AdConfig(c *gin.Context) {
	cfg, err := h.repo.AdConfig(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *Handler) InitConfig(c *gin.Context) {
	controls, err := h.repo.PublicSystemControls(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	levels, err := h.repo.ListLevels(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	ads, err := h.repo.AdConfig(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"systemControls": controls,
		"levels":         levels,
		"ads":            ads,
	})
}

func (h *Handler) Progress(c *gin.Context) {
	progress, err := h.repo.Progress(c.Request.Context(), currentPlayer(c).ID)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.NewClientProgress(progress))
}

func (h *Handler) StartLevel(c *gin.Context) {
	var req struct {
		LevelID int `json:"levelId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	progress, err := h.gameService.StartLevel(c.Request.Context(), currentPlayer(c).ID, req.LevelID)
	if err != nil {
		handleKnownError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"progress": domain.NewClientProgress(progress)})
}

func (h *Handler) SubmitLevelResult(c *gin.Context) {
	var result domain.LevelResult
	if err := c.ShouldBindJSON(&result); err != nil {
		badRequest(c, err.Error())
		return
	}
	progress, rewards, err := h.gameService.SubmitLevelResult(c.Request.Context(), currentPlayer(c), result)
	if err != nil {
		handleKnownError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"progress": domain.NewClientProgress(progress), "rewards": rewards})
}

func (h *Handler) ChangeToolCount(c *gin.Context) {
	var req struct {
		ToolType string `json:"toolType" binding:"required"`
		Delta    int    `json:"delta" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	if req.Delta == 0 {
		badRequest(c, "delta must not be zero")
		return
	}

	source := "ad_reward"
	if req.Delta < 0 {
		source = "use"
	}

	progress, err := h.repo.ChangeToolCount(c.Request.Context(), currentPlayer(c).ID, req.ToolType, req.Delta, source)
	if err != nil {
		handleKnownError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"progress": domain.NewClientProgress(progress)})
}

func (h *Handler) PurchaseToolCount(c *gin.Context) {
	var req struct {
		ToolType string `json:"toolType" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	progress, err := h.repo.PurchaseToolCount(c.Request.Context(), currentPlayer(c).ID, normalizeToolType(req.ToolType), 300, 1)
	if err != nil {
		handleKnownError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"progress": domain.NewClientProgress(progress)})
}

func normalizeToolType(toolType string) string {
	switch toolType {
	case "previewAgain":
		return "preview_again"
	case "removePair":
		return "remove_pair"
	default:
		return toolType
	}
}

func (h *Handler) ShopProducts(c *gin.Context) {
	products, err := h.shopService.Products(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h *Handler) Purchase(c *gin.Context) {
	var req struct {
		ProductKey string `json:"productKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	progress, orderNo, err := h.shopService.Purchase(c.Request.Context(), currentPlayer(c).ID, req.ProductKey)
	if err != nil {
		handleKnownError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"orderNo": orderNo, "progress": domain.NewClientProgress(progress)})
}

func (h *Handler) Leaderboard(c *gin.Context) {
	levelID := queryInt(c, "levelId", 0)
	limit := queryInt(c, "limit", 50)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	entries, err := h.repo.ListLeaderboard(c.Request.Context(), levelID, limit)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (h *Handler) Events(c *gin.Context) {
	var req struct {
		Events []domain.GameEvent `json:"events" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	if err := h.repo.SaveEvents(c.Request.Context(), currentPlayer(c).ID, req.Events); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"accepted": len(req.Events)})
}

func currentPlayer(c *gin.Context) domain.Player {
	value, exists := c.Get("player")
	if !exists {
		return domain.Player{}
	}
	player, _ := value.(domain.Player)
	return player
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func handleKnownError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrInsufficientCoin):
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient coins"})
	case errors.Is(err, repository.ErrInsufficientStamina):
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient stamina"})
	case errors.Is(err, repository.ErrInsufficientTool):
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient tool charges"})
	case errors.Is(err, repository.ErrProductUnavailable):
		c.JSON(http.StatusNotFound, gin.H{"error": "product unavailable"})
	default:
		serverError(c, err)
	}
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
