package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"fpxxl/minigame-api/internal/domain"
	"fpxxl/minigame-api/internal/repository"
)

type PlayerService struct {
	repo repository.Repository
}

func NewPlayerService(repo repository.Repository) *PlayerService {
	return &PlayerService{repo: repo}
}

func (s *PlayerService) Login(ctx context.Context, code string, nickname string, avatarURL string) (domain.Player, domain.PlayerProgress, string, error) {
	openID := devOpenID(code)
	player, progress, err := s.repo.UpsertPlayer(ctx, openID, nickname, avatarURL)
	if err != nil {
		return domain.Player{}, domain.PlayerProgress{}, "", err
	}
	return player, progress, player.OpenID, nil
}

func devOpenID(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		code = "anonymous"
	}
	sum := sha1.Sum([]byte(code))
	return "dev_" + hex.EncodeToString(sum[:])[:24]
}
