package usecase

import (
	"encoding/json"
	"strings"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func validateChallengeText(challengeType domain.ChallengeType, prompt, explanation string) error {
	if strings.TrimSpace(prompt) == "" {
		return api.ErrInvalidInput
	}
	if challengeType != domain.ChallengeTypeTheory && strings.TrimSpace(explanation) == "" {
		return api.ErrInvalidInput
	}
	return nil
}

func parseChallengeMeta(rawType, rawDifficulty string, rawTags json.RawMessage) (domain.ChallengeType, domain.Difficulty, json.RawMessage, error) {
	challengeType, err := domain.NewChallengeType(strings.TrimSpace(rawType))
	if err != nil {
		return "", "", nil, api.ErrInvalidInput
	}
	difficultyRaw := strings.TrimSpace(rawDifficulty)
	if difficultyRaw == "" {
		difficultyRaw = string(domain.DifficultyEasy)
	}
	difficulty, err := domain.NewDifficulty(difficultyRaw)
	if err != nil {
		return "", "", nil, api.ErrInvalidInput
	}
	tags := jsonOrDefault(rawTags, "[]")
	if err := domain.ValidateTags(tags); err != nil {
		return "", "", nil, api.ErrInvalidInput
	}
	return challengeType, difficulty, tags, nil
}

func parseActor(raw string) (domain.UserID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return domain.UserID(uuid.Nil), api.ErrInvalidInput
	}
	return domain.UserID(id), nil
}

func parseActorAndID(actorRaw, idRaw string) (domain.UserID, uuid.UUID, error) {
	actorID, err := parseActor(actorRaw)
	if err != nil {
		return domain.UserID(uuid.Nil), uuid.Nil, err
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return domain.UserID(uuid.Nil), uuid.Nil, api.ErrInvalidInput
	}
	return actorID, id, nil
}

func parseTransition(input api.StatusTransitionInput) (domain.UserID, uuid.UUID, error) {
	if strings.TrimSpace(input.Entity) == "" {
		return domain.UserID(uuid.Nil), uuid.Nil, api.ErrInvalidInput
	}
	return parseActorAndID(input.ActorID, input.ID)
}

func jsonOrDefault(raw []byte, fallback string) []byte {
	if len(raw) == 0 {
		return []byte(fallback)
	}
	return raw
}
