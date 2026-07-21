package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fpxxl/minigame-api/internal/config"
	"fpxxl/minigame-api/internal/domain"
	"fpxxl/minigame-api/internal/repository"
)

type PlayerService struct {
	httpClient     *http.Client
	repo           repository.Repository
	wechatMiniGame config.WechatMiniGameConfig
}

func NewPlayerService(repo repository.Repository, wechatMiniGame config.WechatMiniGameConfig) *PlayerService {
	return &PlayerService{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		repo:           repo,
		wechatMiniGame: wechatMiniGame,
	}
}

func (s *PlayerService) Login(ctx context.Context, code string, nickname string, avatarURL string) (domain.Player, domain.PlayerProgress, string, error) {
	session, err := s.code2Session(ctx, code)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, "", err
	}
	player, progress, err := s.repo.UpsertPlayer(ctx, session.OpenID, session.UnionID, nickname, avatarURL)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, "", err
	}
	return player, progress, player.OpenID, nil
}

type wechatSession struct {
	OpenID     string  `json:"openid"`
	SessionKey string  `json:"session_key"`
	UnionID    *string `json:"unionid,omitempty"`
	ErrCode    int     `json:"errcode,omitempty"`
	ErrMsg     string  `json:"errmsg,omitempty"`
}

func (s *PlayerService) code2Session(ctx context.Context, code string) (wechatSession, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return wechatSession{}, errors.New("wx login code is required")
	}
	if s.wechatMiniGame.AppID == "" || s.wechatMiniGame.Secret == "" {
		return wechatSession{OpenID: devOpenID(code)}, nil
	}

	values := url.Values{}
	values.Set("appid", s.wechatMiniGame.AppID)
	values.Set("secret", s.wechatMiniGame.Secret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	endpoint := "https://api.weixin.qq.com/sns/jscode2session?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return wechatSession{}, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return wechatSession{}, fmt.Errorf("wechat jscode2session request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return wechatSession{}, fmt.Errorf("wechat jscode2session http status %d", resp.StatusCode)
	}

	var session wechatSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return wechatSession{}, fmt.Errorf("decode wechat jscode2session response: %w", err)
	}
	if session.ErrCode != 0 {
		return wechatSession{}, fmt.Errorf("wechat jscode2session error %d: %s", session.ErrCode, session.ErrMsg)
	}
	if session.OpenID == "" {
		return wechatSession{}, errors.New("wechat jscode2session returned empty openid")
	}
	return session, nil
}

func devOpenID(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		code = "anonymous"
	}
	sum := sha1.Sum([]byte(code))
	return "dev_" + hex.EncodeToString(sum[:])[:24]
}
