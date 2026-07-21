package domain

import "time"

type LevelMode string

const (
	LevelModeNormal    LevelMode = "normal"
	LevelModeTimeLimit LevelMode = "time_limit"
	LevelModeStepLimit LevelMode = "step_limit"
)

type LevelConfig struct {
	LevelID                int       `json:"levelId"`
	Rows                   int       `json:"rows"`
	Cols                   int       `json:"cols"`
	PairCount              int       `json:"pairCount"`
	Mode                   LevelMode `json:"mode"`
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
	StaminaCost            int       `json:"staminaCost"`
	ExcellentStepThreshold int       `json:"excellentStepThreshold"`
	NormalStepThreshold    int       `json:"normalStepThreshold"`
	ExcellentTimeThreshold *int      `json:"excellentTimeThreshold,omitempty"`
	NormalTimeThreshold    *int      `json:"normalTimeThreshold,omitempty"`
	TimeLimitSeconds       *int      `json:"timeLimitSeconds,omitempty"`
	StepLimit              *int      `json:"stepLimit,omitempty"`
	Enabled                bool      `json:"enabled"`
	Version                uint      `json:"version"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type AdFrequencyConfig struct {
	NoInterstitialBeforeLevel int      `json:"noInterstitialBeforeLevel"`
	InterstitialEveryLevels   int      `json:"interstitialEveryLevels"`
	MaxInterstitialPerDay     int      `json:"maxInterstitialPerDay"`
	MaxRevivePerLevel         int      `json:"maxRevivePerLevel"`
	BannerEnabledScenes       []string `json:"bannerEnabledScenes"`
	Version                   uint     `json:"version"`
	UpdatedAt                 string   `json:"updatedAt"`
}

type PlayerProgress struct {
	PlayerID        uint64         `json:"playerId"`
	OpenID          string         `json:"openId"`
	CurrentLevel    int            `json:"currentLevel"`
	Coins           int            `json:"coins"`
	Stamina         int            `json:"stamina"`
	MaxStamina      int            `json:"maxStamina"`
	Hints           int            `json:"hints"`
	PreviewAgain    int            `json:"previewAgainCount"`
	RemovePair      int            `json:"removePairCount"`
	LevelStars      map[string]int `json:"levelStars"`
	CompletedLevels []int          `json:"completedLevels"`
	NextRecoverAt   *time.Time     `json:"nextStaminaRecoverAt,omitempty"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

type LevelResult struct {
	OpenID        string `json:"openId"`
	LevelID       int    `json:"levelId"`
	Success       bool   `json:"success"`
	Reason        string `json:"reason"`
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
	EventID   string         `json:"eventId"`
	OpenID    string         `json:"openId"`
	EventType string         `json:"eventType"`
	LevelID   *int           `json:"levelId,omitempty"`
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"createdAt"`
}

type AdminUser struct {
	ID                  uint64         `json:"id"`
	Username            string         `json:"username"`
	Email               *string        `json:"email,omitempty"`
	PasswordHash        string         `json:"-"`
	DisplayName         string         `json:"displayName"`
	Role                string         `json:"role"`
	Permissions         map[string]any `json:"permissions,omitempty"`
	Status              string         `json:"status"`
	FailedLoginAttempts uint           `json:"failedLoginAttempts"`
	LockedUntil         *time.Time     `json:"lockedUntil,omitempty"`
	PasswordChangedAt   *time.Time     `json:"passwordChangedAt,omitempty"`
	LastLoginAt         *time.Time     `json:"lastLoginAt,omitempty"`
	LastLoginIP         string         `json:"lastLoginIp"`
	CreatedBy           *uint64        `json:"createdBy,omitempty"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
}

type AdminPlayerListItem struct {
	ID              uint64     `json:"id"`
	OpenID          string     `json:"openId"`
	Nickname        string     `json:"nickname"`
	AvatarURL       string     `json:"avatarUrl"`
	Status          string     `json:"status"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	CurrentLevel    int        `json:"currentLevel"`
	Coins           int        `json:"coins"`
	Stamina         int        `json:"stamina"`
	MaxStamina      int        `json:"maxStamina"`
	Hints           int        `json:"hints"`
	CompletedCount  int        `json:"completedCount"`
	TotalGames      int        `json:"totalGames"`
	SuccessfulGames int        `json:"successfulGames"`
}

type AdminPlayerProgress struct {
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
	UpdatedAt            *time.Time     `json:"updatedAt,omitempty"`
}

type AdminLevelResult struct {
	ID            uint64    `json:"id"`
	LevelID       int       `json:"levelId"`
	Success       bool      `json:"success"`
	Reason        string    `json:"reason"`
	Steps         int       `json:"steps"`
	MismatchCount int       `json:"mismatchCount"`
	ElapsedMs     int       `json:"elapsedMs"`
	Stars         int       `json:"stars"`
	CoinsEarned   int       `json:"coinsEarned"`
	CompletedAt   time.Time `json:"completedAt"`
}

type AdminPlayerDetail struct {
	Player        AdminPlayerListItem `json:"player"`
	Progress      AdminPlayerProgress `json:"progress"`
	RecentResults []AdminLevelResult  `json:"recentResults"`
}

type SystemControl struct {
	ID               uint64     `json:"id"`
	ControlKey       string     `json:"controlKey"`
	ControlGroup     string     `json:"controlGroup"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	ValueType        string     `json:"valueType"`
	ValueText        *string    `json:"valueText,omitempty"`
	ValueJSON        any        `json:"valueJson,omitempty"`
	DefaultValueText *string    `json:"defaultValueText,omitempty"`
	DefaultValueJSON any        `json:"defaultValueJson,omitempty"`
	Enabled          bool       `json:"enabled"`
	IsPublic         bool       `json:"isPublic"`
	SortOrder        int        `json:"sortOrder"`
	Version          uint       `json:"version"`
	EffectiveFrom    *time.Time `json:"effectiveFrom,omitempty"`
	EffectiveUntil   *time.Time `json:"effectiveUntil,omitempty"`
	CreatedBy        *uint64    `json:"createdBy,omitempty"`
	UpdatedBy        *uint64    `json:"updatedBy,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
