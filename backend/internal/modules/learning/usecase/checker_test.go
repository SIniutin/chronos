package usecase

import (
	"encoding/json"
	"testing"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func TestCheckAnswerByChallengeType(t *testing.T) {
	tests := []struct {
		name      string
		challenge cd.Challenge
		answer    string
		correct   bool
	}{
		{
			name:      "single choice",
			challenge: challenge(cd.ChallengeTypeSingleChoice, `["a"]`),
			answer:    `"a"`,
			correct:   true,
		},
		{
			name:      "multiple choice ignores order",
			challenge: challenge(cd.ChallengeTypeMultiple, `["a","c"]`),
			answer:    `["c","a"]`,
			correct:   true,
		},
		{
			name:      "timeline preserves order",
			challenge: challenge(cd.ChallengeTypeTimeline, `["a","b"]`),
			answer:    `["b","a"]`,
			correct:   false,
		},
		{
			name:      "match pairs ignores pair order",
			challenge: challenge(cd.ChallengeTypeMatchPairs, `[{"left_id":"l1","right_id":"r1"},{"left_id":"l2","right_id":"r2"}]`),
			answer:    `[{"left_id":"l2","right_id":"r2"},{"left_id":"l1","right_id":"r1"}]`,
			correct:   true,
		},
		{
			name:      "fill blank case folds",
			challenge: challenge(cd.ChallengeTypeFillBlank, `["1905"]`),
			answer:    `" 1905 "`,
			correct:   true,
		},
		{
			name:      "true false",
			challenge: challenge(cd.ChallengeTypeTrueFalse, `["false"]`),
			answer:    `"true"`,
			correct:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkAnswer(tt.challenge, json.RawMessage(tt.answer))
			if got.isCorrect != tt.correct {
				t.Fatalf("expected correct=%v, got %+v", tt.correct, got)
			}
		})
	}
}

func challenge(t cd.ChallengeType, answers string) cd.Challenge {
	return cd.Challenge{
		ID:      cd.ChallengeID(uuid.New()),
		SkillID: cd.SkillID(uuid.New()),
		Type:    t,
		Answers: json.RawMessage(answers),
		Status:  cd.ContentStatusPublished,
	}
}
