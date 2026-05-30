package seeder

import (
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
	if sections != 2 || units != 6 || skills != 13 || challenges != 52 {
		t.Fatalf("unexpected counts: sections=%d units=%d skills=%d challenges=%d", sections, units, skills, challenges)
	}
	for _, section := range seed.Sections {
		for _, unit := range section.Units {
			for _, skill := range unit.Skills {
				if len(skill.Challenges) == 0 {
					t.Fatalf("skill %q has no challenges", skill.Title)
				}
			}
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
