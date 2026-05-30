package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func (s *Service) ListPublishedChallenges(ctx context.Context, input api.ListChallengesInput) ([]domain.Challenge, error) {
	skillID, err := domain.ParseSkillID(input.SkillID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	challenges, err := s.challengeRepo.ListPublishedChallenges(ctx, skillID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return challenges, nil
}

func (s *Service) ListAllChallenges(ctx context.Context, input api.ListChallengesInput) ([]domain.Challenge, error) {
	skillID, err := domain.ParseSkillID(input.SkillID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	challenges, err := s.challengeRepo.ListAllChallenges(ctx, skillID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return challenges, nil
}

func (s *Service) CreateChallenge(ctx context.Context, input api.ChallengeWriteInput) (domain.Challenge, error) {
	actorID, err := parseActor(input.ActorID)
	if err != nil {
		return domain.Challenge{}, err
	}
	skillID, err := domain.ParseSkillID(input.SkillID)
	if err != nil {
		return domain.Challenge{}, mapDomainError(err)
	}
	challengeType, difficulty, tags, err := parseChallengeMeta(input.Type, input.Difficulty, input.Tags)
	if err != nil {
		return domain.Challenge{}, err
	}
	if err := validateChallengeText(challengeType, input.Prompt, input.Explanation); err != nil {
		return domain.Challenge{}, err
	}
	now := time.Now().UTC()
	challenge := domain.Challenge{
		ID:          domain.ChallengeID(uuid.New()),
		SkillID:     skillID,
		Type:        challengeType,
		Difficulty:  difficulty,
		Tags:        tags,
		Level:       input.Level,
		LessonCount: input.LessonCount,
		Prompt:      strings.TrimSpace(input.Prompt),
		Body:        input.Body,
		Payload:     jsonOrDefault(input.Payload, "{}"),
		Options:     jsonOrDefault(input.Options, "[]"),
		Answers:     jsonOrDefault(input.Answers, "[]"),
		Explanation: input.Explanation,
		Position:    input.Position,
		Status:      domain.ContentStatusDraft,
		Audit:       newAudit(actorID, now),
	}
	created, err := s.challengeRepo.CreateChallenge(ctx, challenge)
	return created, mapDomainError(err)
}

func (s *Service) UpdateChallenge(ctx context.Context, input api.ChallengeWriteInput) (domain.Challenge, error) {
	actorID, id, err := parseActorAndID(input.ActorID, input.ID)
	if err != nil {
		return domain.Challenge{}, err
	}
	skillID, err := domain.ParseSkillID(input.SkillID)
	if err != nil {
		return domain.Challenge{}, mapDomainError(err)
	}
	challengeType, difficulty, tags, err := parseChallengeMeta(input.Type, input.Difficulty, input.Tags)
	if err != nil {
		return domain.Challenge{}, err
	}
	if err := validateChallengeText(challengeType, input.Prompt, input.Explanation); err != nil {
		return domain.Challenge{}, err
	}
	challenge := domain.Challenge{
		ID:          domain.ChallengeID(id),
		SkillID:     skillID,
		Type:        challengeType,
		Difficulty:  difficulty,
		Tags:        tags,
		Level:       input.Level,
		LessonCount: input.LessonCount,
		Prompt:      strings.TrimSpace(input.Prompt),
		Body:        input.Body,
		Payload:     jsonOrDefault(input.Payload, "{}"),
		Options:     jsonOrDefault(input.Options, "[]"),
		Answers:     jsonOrDefault(input.Answers, "[]"),
		Explanation: input.Explanation,
		Position:    input.Position,
		Audit:       updateAudit(actorID),
	}
	updated, err := s.challengeRepo.UpdateChallenge(ctx, challenge)
	return updated, mapDomainError(err)
}
