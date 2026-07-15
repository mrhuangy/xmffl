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
	ExcellentStepThreshold int       `json:"excellentStepThreshold"`
	NormalStepThreshold    int       `json:"normalStepThreshold"`
	ExcellentTimeThreshold *int      `json:"excellentTimeThreshold,omitempty"`
	NormalTimeThreshold    *int      `json:"normalTimeThreshold,omitempty"`
	TimeLimitSeconds       *int      `json:"timeLimitSeconds,omitempty"`
	StepLimit              *int      `json:"stepLimit,omitempty"`
	Enabled                bool      `json:"enabled"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type AdFrequencyConfig struct {
	NoInterstitialBeforeLevel int      `json:"noInterstitialBeforeLevel"`
	InterstitialEveryLevels   int      `json:"interstitialEveryLevels"`
	MaxInterstitialPerDay     int      `json:"maxInterstitialPerDay"`
	MaxRevivePerLevel         int      `json:"maxRevivePerLevel"`
	BannerEnabledScenes       []string `json:"bannerEnabledScenes"`
	UpdatedAt                 string   `json:"updatedAt"`
}

type PlayerProgress struct {
	OpenID          string         `json:"openId"`
	CurrentLevel    int            `json:"currentLevel"`
	Coins           int            `json:"coins"`
	Hints           int            `json:"hints"`
	LevelStars      map[string]int `json:"levelStars"`
	CompletedLevels []int          `json:"completedLevels"`
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
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"createdAt"`
}
