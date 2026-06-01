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
			name:      "match photos ignores pair order",
			challenge: challenge(cd.ChallengeTypeMatchPhotos, `[{"photo_id":"p1","label_id":"l1"},{"photo_id":"p2","label_id":"l2"}]`),
			answer:    `[{"photo_id":"p2","label_id":"l2"},{"photo_id":"p1","label_id":"l1"}]`,
			correct:   true,
		},
		{
			name:      "match photos incorrect pair",
			challenge: challenge(cd.ChallengeTypeMatchPhotos, `[{"photo_id":"p1","label_id":"l1"},{"photo_id":"p2","label_id":"l2"}]`),
			answer:    `[{"photo_id":"p1","label_id":"l2"},{"photo_id":"p2","label_id":"l1"}]`,
			correct:   false,
		},
		{
			name:      "match photos missing pair",
			challenge: challenge(cd.ChallengeTypeMatchPhotos, `[{"photo_id":"p1","label_id":"l1"},{"photo_id":"p2","label_id":"l2"}]`),
			answer:    `[{"photo_id":"p1","label_id":"l1"}]`,
			correct:   false,
		},
		{
			name:      "match photos extra pair",
			challenge: challenge(cd.ChallengeTypeMatchPhotos, `[{"photo_id":"p1","label_id":"l1"}]`),
			answer:    `[{"photo_id":"p1","label_id":"l1"},{"photo_id":"p2","label_id":"l2"}]`,
			correct:   false,
		},
		{
			name:      "match photos malformed answer",
			challenge: challenge(cd.ChallengeTypeMatchPhotos, `[{"photo_id":"p1","label_id":"l1"}]`),
			answer:    `{"photo_id":"p1","label_id":"l1"}`,
			correct:   false,
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
		{
			name:      "map point inside radius",
			challenge: challenge(cd.ChallengeTypeMapPoint, `{"lat":55.7558,"lng":37.6173,"radius_m":2000}`),
			answer:    `{"lat":55.7600,"lng":37.6200}`,
			correct:   true,
		},
		{
			name:      "map point outside radius",
			challenge: challenge(cd.ChallengeTypeMapPoint, `{"lat":55.7558,"lng":37.6173,"radius_m":2000}`),
			answer:    `{"lat":55.8000,"lng":37.7000}`,
			correct:   false,
		},
		{
			name:      "map point malformed answer",
			challenge: challenge(cd.ChallengeTypeMapPoint, `{"lat":55.7558,"lng":37.6173,"radius_m":2000}`),
			answer:    `{"x":55.7600}`,
			correct:   false,
		},
		{
			name:      "map point malformed challenge answer",
			challenge: challenge(cd.ChallengeTypeMapPoint, `{"lat":55.7558,"lng":37.6173}`),
			answer:    `{"lat":55.7600,"lng":37.6200}`,
			correct:   false,
		},
		{
			name:      "map area center and size within tolerance",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000,"area_tolerance":0.3}`),
			answer:    `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431}`,
			correct:   true,
		},
		{
			name:      "map area wrong center",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000,"area_tolerance":0.3}`),
			answer:    `{"center":{"lat":1,"lng":1},"area_m2":1236431}`,
			correct:   false,
		},
		{
			name:      "map area wrong size",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000,"area_tolerance":0.2}`),
			answer:    `{"center":{"lat":0.005,"lng":0.005},"area_m2":1000}`,
			correct:   false,
		},
		{
			name:      "map area malformed answer",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000,"area_tolerance":0.3}`),
			answer:    `{"center":{"lat":0.005},"area_m2":1236431}`,
			correct:   false,
		},
		{
			name:      "map area malformed challenge answer",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000}`),
			answer:    `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431}`,
			correct:   false,
		},
		{
			name:      "map area legacy polygon answer",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000,"area_tolerance":0.3}`),
			answer:    `{"points":[{"lat":0,"lng":0},{"lat":0,"lng":0.01},{"lat":0.01,"lng":0.01},{"lat":0.01,"lng":0}]}`,
			correct:   true,
		},
		{
			name:      "map area legacy polygon needs three points",
			challenge: challenge(cd.ChallengeTypeMapArea, `{"center":{"lat":0.005,"lng":0.005},"area_m2":1236431,"center_radius_m":1000,"area_tolerance":0.3}`),
			answer:    `{"points":[{"lat":0,"lng":0},{"lat":0,"lng":0.01}]}`,
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
