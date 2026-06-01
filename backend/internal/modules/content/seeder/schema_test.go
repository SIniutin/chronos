package seeder

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestLoadFileParsesStructuredHistorySeed(t *testing.T) {
	seed, err := LoadFile(filepath.Join("..", "..", "..", "..", "seeds", "history_course_structured.json"))
	if err != nil {
		t.Fatalf("load seed failed: %v", err)
	}
	if seed.Course.Title != "История России начала XX века" {
		t.Fatalf("unexpected course title: %q", seed.Course.Title)
	}
	sections, units, skills, challenges := countSeed(seed)
	if sections != 39 || units != 50 || skills != 57 || challenges != 289 {
		t.Fatalf("unexpected counts: sections=%d units=%d skills=%d challenges=%d", sections, units, skills, challenges)
	}
	for _, section := range seed.Sections {
		for _, unit := range section.Units {
			for _, skill := range unit.Skills {
				assertSkillSeedValid(t, skill)
			}
		}
	}
}

func assertSkillSeedValid(t *testing.T, skill SeedSkill) {
	t.Helper()
	if len(skill.Challenges) == 0 {
		t.Fatalf("skill %q has no challenges", skill.Title)
	}
	hasTheory := false
	for index, challenge := range skill.Challenges {
		if challenge.Type == "theory" {
			hasTheory = true
		}
		if challenge.Position != index+1 {
			t.Fatalf("skill %q challenge %q position=%d, want %d", skill.Title, challenge.Prompt, challenge.Position, index+1)
		}
		if challenge.Type == "" || challenge.Difficulty == "" || challenge.Prompt == "" || challenge.Status == "" {
			t.Fatalf("skill %q has incomplete challenge: %+v", skill.Title, challenge)
		}
		if len(challenge.Tags) == 0 || len(challenge.Payload) == 0 || len(challenge.Options) == 0 || len(challenge.Answers) == 0 {
			t.Fatalf("skill %q challenge %q has empty raw json fields", skill.Title, challenge.Prompt)
		}
		assertAnswersReferenceOptions(t, skill.Title, challenge)
	}
	if !hasTheory {
		t.Fatalf("skill %q has no theory challenge", skill.Title)
	}
}

func assertAnswersReferenceOptions(t *testing.T, skillTitle string, challenge SeedChallenge) {
	t.Helper()
	switch challenge.Type {
	case "single_choice", "multiple_choice", "true_false", "image_question":
		var options []struct {
			ID string `json:"id"`
		}
		var answers []string
		if err := json.Unmarshal(challenge.Options, &options); err != nil {
			t.Fatalf("skill %q challenge %q options decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if err := json.Unmarshal(challenge.Answers, &answers); err != nil {
			t.Fatalf("skill %q challenge %q answers decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		ids := map[string]bool{}
		for _, option := range options {
			ids[option.ID] = true
		}
		for _, answer := range answers {
			if !ids[answer] {
				t.Fatalf("skill %q challenge %q answer %q has no option", skillTitle, challenge.Prompt, answer)
			}
		}
	case "match_pairs":
		var options struct {
			Left []struct {
				ID string `json:"id"`
			} `json:"left"`
			Right []struct {
				ID string `json:"id"`
			} `json:"right"`
		}
		var answers []struct {
			LeftID  string `json:"left_id"`
			RightID string `json:"right_id"`
		}
		if err := json.Unmarshal(challenge.Options, &options); err != nil {
			t.Fatalf("skill %q challenge %q options decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if err := json.Unmarshal(challenge.Answers, &answers); err != nil {
			t.Fatalf("skill %q challenge %q answers decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		leftIDs := map[string]bool{}
		rightIDs := map[string]bool{}
		for _, option := range options.Left {
			leftIDs[option.ID] = true
		}
		for _, option := range options.Right {
			rightIDs[option.ID] = true
		}
		for _, answer := range answers {
			if !leftIDs[answer.LeftID] || !rightIDs[answer.RightID] {
				t.Fatalf("skill %q challenge %q has invalid pair answer %+v", skillTitle, challenge.Prompt, answer)
			}
		}
	case "match_photos":
		var options struct {
			Photos []struct {
				ID       string `json:"id"`
				ImageURL string `json:"image_url"`
				Alt      string `json:"alt"`
			} `json:"photos"`
			Labels []struct {
				ID   string `json:"id"`
				Text string `json:"text"`
			} `json:"labels"`
		}
		var answers []struct {
			PhotoID string `json:"photo_id"`
			LabelID string `json:"label_id"`
		}
		if err := json.Unmarshal(challenge.Options, &options); err != nil {
			t.Fatalf("skill %q challenge %q options decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if err := json.Unmarshal(challenge.Answers, &answers); err != nil {
			t.Fatalf("skill %q challenge %q answers decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		photoIDs := map[string]bool{}
		labelIDs := map[string]bool{}
		for _, photo := range options.Photos {
			if photo.ID == "" || photo.ImageURL == "" || photo.Alt == "" {
				t.Fatalf("skill %q challenge %q has incomplete photo %+v", skillTitle, challenge.Prompt, photo)
			}
			photoIDs[photo.ID] = true
		}
		for _, label := range options.Labels {
			if label.ID == "" || label.Text == "" {
				t.Fatalf("skill %q challenge %q has incomplete label %+v", skillTitle, challenge.Prompt, label)
			}
			labelIDs[label.ID] = true
		}
		for _, answer := range answers {
			if !photoIDs[answer.PhotoID] || !labelIDs[answer.LabelID] {
				t.Fatalf("skill %q challenge %q has invalid photo answer %+v", skillTitle, challenge.Prompt, answer)
			}
		}
	case "map_point":
		var payload struct {
			Center *struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"center"`
			Zoom float64 `json:"zoom"`
		}
		var answers struct {
			Lat     float64 `json:"lat"`
			Lng     float64 `json:"lng"`
			RadiusM float64 `json:"radius_m"`
		}
		if err := json.Unmarshal(challenge.Payload, &payload); err != nil {
			t.Fatalf("skill %q challenge %q payload decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if err := json.Unmarshal(challenge.Answers, &answers); err != nil {
			t.Fatalf("skill %q challenge %q answers decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if payload.Center == nil || payload.Zoom <= 0 || answers.RadiusM <= 0 {
			t.Fatalf("skill %q challenge %q has invalid map_point", skillTitle, challenge.Prompt)
		}
	case "map_area":
		var payload struct {
			Center *struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"center"`
			Zoom float64 `json:"zoom"`
		}
		var answers struct {
			Center *struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"center"`
			AreaM2        float64 `json:"area_m2"`
			CenterRadiusM float64 `json:"center_radius_m"`
			AreaTolerance float64 `json:"area_tolerance"`
		}
		if err := json.Unmarshal(challenge.Payload, &payload); err != nil {
			t.Fatalf("skill %q challenge %q payload decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if err := json.Unmarshal(challenge.Answers, &answers); err != nil {
			t.Fatalf("skill %q challenge %q answers decode failed: %v", skillTitle, challenge.Prompt, err)
		}
		if payload.Center == nil || payload.Zoom <= 0 || answers.Center == nil || answers.AreaM2 <= 0 || answers.CenterRadiusM <= 0 || answers.AreaTolerance < 0 {
			t.Fatalf("skill %q challenge %q has invalid map_area", skillTitle, challenge.Prompt)
		}
	}
}

func countSeed(seed SeedFile) (sections int, units int, skills int, challenges int) {
	sections = len(seed.Sections)
	for _, section := range seed.Sections {
		units += len(section.Units)
		for _, unit := range section.Units {
			skills += len(unit.Skills)
			for _, skill := range unit.Skills {
				challenges += len(skill.Challenges)
			}
		}
	}
	return sections, units, skills, challenges
}
