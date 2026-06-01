package domain

import "testing"

func TestChallengeTypeValidation(t *testing.T) {
	validTypes := []string{
		"theory",
		"single_choice",
		"multiple_choice",
		"timeline",
		"match_pairs",
		"image_question",
		"match_image",
		"match_photos",
		"quote_question",
		"true_false",
		"fill_in_blank",
		"map_point",
		"map_area",
	}
	for _, typ := range validTypes {
		if _, err := NewChallengeType(typ); err != nil {
			t.Fatalf("%s should be valid: %v", typ, err)
		}
	}
	if _, err := NewChallengeType("single_quiz"); err == nil {
		t.Fatalf("single_quiz should be invalid")
	}
}

func TestDifficultyValidation(t *testing.T) {
	if _, err := NewDifficulty("medium"); err != nil {
		t.Fatalf("medium should be valid: %v", err)
	}
	if _, err := NewDifficulty("legendary"); err == nil {
		t.Fatalf("legendary should be invalid")
	}
}

func TestTagsValidation(t *testing.T) {
	if err := ValidateTags([]byte(`["witte","1905"]`)); err != nil {
		t.Fatalf("valid tags rejected: %v", err)
	}
	if err := ValidateTags([]byte(`{"bad":true}`)); err == nil {
		t.Fatalf("object tags should be invalid")
	}
}
