package domain

import "testing"

func TestChallengeTypeValidation(t *testing.T) {
	if _, err := NewChallengeType("theory"); err != nil {
		t.Fatalf("theory should be valid: %v", err)
	}
	if _, err := NewChallengeType("fill_in_blank"); err != nil {
		t.Fatalf("fill_in_blank should be valid: %v", err)
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
