package domain

import "time"

type Player struct {
	ID          uint64     `json:"id"`
	OpenID      string     `json:"openId"`
	UnionID     *string    `json:"unionId,omitempty"`
	Nickname    string     `json:"nickname"`
	AvatarURL   string     `json:"avatarUrl"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type PlayerProgress struct {
	PlayerID             uint64         `json:"playerId"`
	CurrentLevel         int            `json:"currentLevel"`
	Coins                int            `json:"coins"`
	Stamina              int            `json:"stamina"`
	MaxStamina           int            `json:"maxStamina"`
	Hints                int            `json:"hints"`
	PreviewAgainCount    int            `json:"previewAgainCount"`
	RemovePairCount      int            `json:"removePairCount"`
	LevelStars           map[string]int `json:"levelStars"`
	CompletedLevels      []int          `json:"completedLevels"`
	NextStaminaRecoverAt *time.Time     `json:"nextStaminaRecoverAt,omitempty"`
	UpdatedAt            time.Time      `json:"updatedAt"`
}

type ClientProgress struct {
	Version              int            `json:"version"`
	CurrentLevel         int            `json:"currentLevel"`
	Coins                int            `json:"coins"`
	Hints                int            `json:"hints"`
	PreviewAgainCount    int            `json:"previewAgainCount"`
	RemovePairCount      int            `json:"removePairCount"`
	Stamina              int            `json:"stamina"`
	MaxStamina           int            `json:"maxStamina"`
	NextStaminaRecoverAt *int64         `json:"nextStaminaRecoverAt,omitempty"`
	LevelStars           map[string]int `json:"levelStars"`
	CompletedLevels      []int          `json:"completedLevels"`
	UpdatedAt            int64          `json:"updatedAt"`
}

func NewClientProgress(progress PlayerProgress) ClientProgress {
	updatedAt := progress.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	levelStars := progress.LevelStars
	if levelStars == nil {
		levelStars = map[string]int{}
	}
	completedLevels := progress.CompletedLevels
	if completedLevels == nil {
		completedLevels = []int{}
	}
	var nextStaminaRecoverAt *int64
	if progress.NextStaminaRecoverAt != nil {
		value := progress.NextStaminaRecoverAt.UnixMilli()
		nextStaminaRecoverAt = &value
	}
	return ClientProgress{
		Version:              1,
		CurrentLevel:         progress.CurrentLevel,
		Coins:                progress.Coins,
		Hints:                progress.Hints,
		PreviewAgainCount:    progress.PreviewAgainCount,
		RemovePairCount:      progress.RemovePairCount,
		Stamina:              progress.Stamina,
		MaxStamina:           progress.MaxStamina,
		NextStaminaRecoverAt: nextStaminaRecoverAt,
		LevelStars:           levelStars,
		CompletedLevels:      completedLevels,
		UpdatedAt:            updatedAt.UnixMilli(),
	}
}

type LevelConfig struct {
	LevelID                int       `json:"levelId"`
	Rows                   int       `json:"rows"`
	Cols                   int       `json:"cols"`
	PairCount              int       `json:"pairCount"`
	Mode                   string    `json:"mode"`
	ThemeID                string    `json:"themeId"`
	InitialPreviewMs       int       `json:"initialPreviewMs"`
	FlipBackDelayMs        int       `json:"flipBackDelayMs"`
	LevelTimeLimitSeconds  int       `json:"levelTimeLimitSeconds"`
	MaxMismatchCount       int       `json:"maxMismatchCount"`
	ShowSteps              bool      `json:"showSteps"`
	ShowTimer              bool      `json:"showTimer"`
	ShowMismatch           bool      `json:"showMismatch"`
	HintHighlightMs        int       `json:"hintHighlightMs"`
	CoinRewardBase         int       `json:"coinRewardBase"`
	CoinRewardStar1        int       `json:"coinRewardStar1"`
	CoinRewardStar2        int       `json:"coinRewardStar2"`
	CoinRewardStar3        int       `json:"coinRewardStar3"`
	StaminaCost            int       `json:"staminaCost"`
	ExcellentStepThreshold int       `json:"excellentStepThreshold"`
	NormalStepThreshold    int       `json:"normalStepThreshold"`
	ExcellentTimeThreshold *int      `json:"excellentTimeThreshold,omitempty"`
	NormalTimeThreshold    *int      `json:"normalTimeThreshold,omitempty"`
	TimeLimitSeconds       *int      `json:"timeLimitSeconds,omitempty"`
	StepLimit              *int      `json:"stepLimit,omitempty"`
	Version                int       `json:"version"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type AdFrequencyConfig struct {
	NoInterstitialBeforeLevel int      `json:"noInterstitialBeforeLevel"`
	InterstitialEveryLevels   int      `json:"interstitialEveryLevels"`
	MaxInterstitialPerDay     int      `json:"maxInterstitialPerDay"`
	MaxRevivePerLevel         int      `json:"maxRevivePerLevel"`
	BannerEnabledScenes       []string `json:"bannerEnabledScenes"`
	Version                   int      `json:"version"`
	UpdatedAt                 string   `json:"updatedAt"`
}

type PublicSystemControls map[string]any

type ShopProduct struct {
	ID             uint64    `json:"id"`
	ProductKey     string    `json:"productKey"`
	Name           string    `json:"name"`
	ProductType    string    `json:"productType"`
	CurrencyType   string    `json:"currencyType"`
	CurrencyAmount int       `json:"currencyAmount"`
	GrantType      string    `json:"grantType"`
	GrantAmount    int       `json:"grantAmount"`
	DailyBuyLimit  *int      `json:"dailyBuyLimit,omitempty"`
	SortOrder      int       `json:"sortOrder"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type LevelResult struct {
	LevelID       int    `json:"levelId" binding:"required"`
	Success       bool   `json:"success"`
	Reason        string `json:"reason" binding:"required"`
	Steps         int    `json:"steps"`
	MismatchCount int    `json:"mismatchCount"`
	ElapsedMs     int    `json:"elapsedMs"`
	Stars         int    `json:"stars"`
	CoinsEarned   int    `json:"coinsEarned"`
	UsedHints     int    `json:"usedHints"`
}

type LeaderboardEntry struct {
	OpenID      string    `json:"openId"`
	Nickname    string    `json:"nickname"`
	LevelID     int       `json:"levelId"`
	Stars       int       `json:"stars"`
	Steps       int       `json:"steps"`
	ElapsedMs   int       `json:"elapsedMs"`
	SubmittedAt time.Time `json:"submittedAt"`
}

type GameEvent struct {
	EventID   string         `json:"eventId" binding:"required"`
	EventType string         `json:"eventType" binding:"required"`
	LevelID   *int           `json:"levelId,omitempty"`
	Payload   map[string]any `json:"payload"`
}
