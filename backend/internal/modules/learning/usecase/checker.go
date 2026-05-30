package usecase

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
)

type checkResult struct {
	isCorrect bool
	mistakes  []string
}

func checkAnswer(challenge content_domain.Challenge, userAnswer json.RawMessage) checkResult {
	if challenge.Type == content_domain.ChallengeTypeTheory {
		return checkResult{isCorrect: true}
	}
	if len(userAnswer) == 0 {
		return incorrect("answer is empty")
	}

	expected, err := decodeAny(challenge.Answers)
	if err != nil {
		return incorrect("challenge answers are malformed")
	}
	submitted, err := decodeAny(userAnswer)
	if err != nil {
		return incorrect("answer is malformed")
	}

	switch challenge.Type {
	case content_domain.ChallengeTypeSingleChoice, content_domain.ChallengeTypeImage, content_domain.ChallengeTypeQuote, content_domain.ChallengeTypeTrueFalse:
		return checkSingle(expected, submitted)
	case content_domain.ChallengeTypeMultiple:
		return checkStringSet(expected, submitted)
	case content_domain.ChallengeTypeTimeline:
		return checkStringOrder(expected, submitted)
	case content_domain.ChallengeTypeMatchPairs, content_domain.ChallengeTypeMatchImage:
		return checkPairs(expected, submitted)
	case content_domain.ChallengeTypeFillBlank:
		return checkFillBlank(expected, submitted)
	default:
		return incorrect("challenge type is unsupported")
	}
}

func checkSingle(expected any, submitted any) checkResult {
	expectedItems := scalarList(expected)
	if len(expectedItems) != 1 {
		return incorrect("challenge expects exactly one answer")
	}
	if normalizeScalar(submitted) != expectedItems[0] {
		return incorrect("answer does not match")
	}
	return checkResult{isCorrect: true}
}

func checkStringSet(expected any, submitted any) checkResult {
	expectedItems := scalarList(expected)
	submittedItems := scalarList(submitted)
	sort.Strings(expectedItems)
	sort.Strings(submittedItems)
	if !sameStrings(expectedItems, submittedItems) {
		return incorrect("selected options do not match")
	}
	return checkResult{isCorrect: true}
}

func checkStringOrder(expected any, submitted any) checkResult {
	if !sameStrings(scalarList(expected), scalarList(submitted)) {
		return incorrect("order does not match")
	}
	return checkResult{isCorrect: true}
}

func checkPairs(expected any, submitted any) checkResult {
	expectedPairs := pairList(expected)
	submittedPairs := pairList(submitted)
	sort.Strings(expectedPairs)
	sort.Strings(submittedPairs)
	if len(expectedPairs) == 0 || !sameStrings(expectedPairs, submittedPairs) {
		return incorrect("pairs do not match")
	}
	return checkResult{isCorrect: true}
}

func checkFillBlank(expected any, submitted any) checkResult {
	answer := normalizeText(normalizeScalar(submitted))
	if answer == "" {
		return incorrect("answer is empty")
	}
	for _, item := range scalarList(expected) {
		if normalizeText(item) == answer {
			return checkResult{isCorrect: true}
		}
	}
	return incorrect("answer does not match")
}

func decodeAny(raw json.RawMessage) (any, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func scalarList(value any) []string {
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if normalized := normalizeScalar(item); normalized != "" {
				out = append(out, normalized)
			}
		}
		return out
	default:
		if normalized := normalizeScalar(v); normalized != "" {
			return []string{normalized}
		}
		return nil
	}
}

func normalizeScalar(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func pairList(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		left := normalizeScalar(m["left_id"])
		right := normalizeScalar(m["right_id"])
		if left != "" && right != "" {
			out = append(out, left+"="+right)
		}
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func normalizeText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func incorrect(reason string) checkResult {
	return checkResult{mistakes: []string{reason}}
}
