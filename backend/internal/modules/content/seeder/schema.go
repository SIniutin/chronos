package seeder

import (
	"encoding/json"
	"fmt"
	"os"
)

type SeedFile struct {
	Course   SeedCourse    `json:"course"`
	Sections []SeedSection `json:"sections"`
}

type SeedCourse struct {
	Title      string `json:"title"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type SeedSection struct {
	Theme       string     `json:"theme"`
	Description string     `json:"description"`
	Units       []SeedUnit `json:"units"`
}

type SeedUnit struct {
	Title  string      `json:"title"`
	Skills []SeedSkill `json:"skills"`
}

type SeedSkill struct {
	Title      string          `json:"title"`
	Icon       string          `json:"icon"`
	Challenges []SeedChallenge `json:"challenges"`
}

type SeedChallenge struct {
	Type        string          `json:"type"`
	Difficulty  string          `json:"difficulty"`
	Tags        json.RawMessage `json:"tags"`
	Level       int             `json:"level"`
	LessonCount int             `json:"lesson_count"`
	Prompt      string          `json:"prompt"`
	Body        string          `json:"body"`
	Payload     json.RawMessage `json:"payload"`
	Options     json.RawMessage `json:"options"`
	Answers     json.RawMessage `json:"answers"`
	Explanation string          `json:"explanation"`
	Position    int             `json:"position"`
	Status      string          `json:"status"`
}

func LoadFile(path string) (SeedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SeedFile{}, fmt.Errorf("read seed file: %w", err)
	}
	var seed SeedFile
	if err := json.Unmarshal(data, &seed); err != nil {
		return SeedFile{}, fmt.Errorf("decode seed file: %w", err)
	}
	return seed, nil
}
