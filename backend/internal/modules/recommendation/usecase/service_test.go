package usecase

import (
	"context"
	"testing"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	pd "github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	rd "github.com/SIniutin/history-app-backend/internal/modules/recommendation/domain"
	"github.com/google/uuid"
)

func TestGetNextSkillRecommendationOrder(t *testing.T) {
	userID := rd.UserID(uuid.New())
	content := newRecContent()
	progress := newRecProgress()
	uc := NewService(content, progress)

	rec, err := uc.GetNextSkill(context.Background(), userID, content.course.ID)
	if err != nil {
		t.Fatalf("next failed: %v", err)
	}
	if rec.Type != rd.RecommendationTypeNewSkill || *rec.SkillID != content.skills[0].ID {
		t.Fatalf("expected first new skill, got %+v", rec)
	}

	progress.skills[key(userID, content.skills[1].ID)] = pd.SkillProgress{UserID: userID, SkillID: content.skills[1].ID, Status: pd.ProgressStatusInProgress}
	rec, err = uc.GetNextSkill(context.Background(), userID, content.course.ID)
	if err != nil {
		t.Fatalf("next failed: %v", err)
	}
	if rec.Type != rd.RecommendationTypeContinue || *rec.SkillID != content.skills[1].ID {
		t.Fatalf("expected continue, got %+v", rec)
	}

	progress.skills = map[string]pd.SkillProgress{}
	progress.skills[key(userID, content.skills[0].ID)] = pd.SkillProgress{UserID: userID, SkillID: content.skills[0].ID, Status: pd.ProgressStatusAvailable, Mastery: 0.4}
	rec, err = uc.GetNextSkill(context.Background(), userID, content.course.ID)
	if err != nil {
		t.Fatalf("next failed: %v", err)
	}
	if rec.Type != rd.RecommendationTypeReview {
		t.Fatalf("expected low mastery review, got %+v", rec)
	}
}

func TestPickChallengesForSession(t *testing.T) {
	content := newRecContent()
	uc := NewService(content, newRecProgress())
	ids, err := uc.PickChallengesForSession(context.Background(), rd.UserID(uuid.New()), content.skills[0].ID, 2)
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if len(ids) != 2 || ids[0] != content.challenges[0].ID || ids[1] != content.challenges[1].ID {
		t.Fatalf("unexpected picks: %+v", ids)
	}
}

type recContent struct {
	course     cd.Course
	section    cd.Section
	unit       cd.Unit
	skills     []cd.Skill
	challenges []cd.Challenge
}

func newRecContent() *recContent {
	course := cd.Course{ID: cd.CourseID(uuid.New())}
	section := cd.Section{ID: cd.SectionID(uuid.New()), CourseID: course.ID}
	unit := cd.Unit{ID: cd.UnitID(uuid.New()), SectionID: section.ID}
	skills := []cd.Skill{{ID: cd.SkillID(uuid.New()), UnitID: unit.ID, Position: 1}, {ID: cd.SkillID(uuid.New()), UnitID: unit.ID, Position: 2}}
	challenges := []cd.Challenge{{ID: cd.ChallengeID(uuid.New()), SkillID: skills[0].ID, Position: 1, Status: cd.ContentStatusPublished}, {ID: cd.ChallengeID(uuid.New()), SkillID: skills[0].ID, Position: 2, Status: cd.ContentStatusPublished}, {ID: cd.ChallengeID(uuid.New()), SkillID: skills[0].ID, Position: 3, Status: cd.ContentStatusDraft}}
	return &recContent{course: course, section: section, unit: unit, skills: skills, challenges: challenges}
}

func (c *recContent) ListPublishedSections(context.Context, cd.CourseID) ([]cd.Section, error) {
	return []cd.Section{c.section}, nil
}
func (c *recContent) ListPublishedUnits(context.Context, cd.SectionID) ([]cd.Unit, error) {
	return []cd.Unit{c.unit}, nil
}
func (c *recContent) ListPublishedSkills(context.Context, cd.UnitID) ([]cd.Skill, error) {
	return c.skills, nil
}
func (c *recContent) ListPublishedChallenges(_ context.Context, skillID cd.SkillID) ([]cd.Challenge, error) {
	var out []cd.Challenge
	for _, ch := range c.challenges {
		if ch.SkillID == skillID && ch.Status.IsPublished() {
			out = append(out, ch)
		}
	}
	return out, nil
}

type recProgress struct{ skills map[string]pd.SkillProgress }

func newRecProgress() *recProgress { return &recProgress{skills: map[string]pd.SkillProgress{}} }

func (p *recProgress) GetSkillProgress(_ context.Context, userID pd.UserID, skillID pd.SkillID) (*pd.SkillProgress, error) {
	progress, ok := p.skills[key(userID, skillID)]
	if !ok {
		return nil, pd.ErrNotFound
	}
	return &progress, nil
}

func (p *recProgress) ListSkillProgressByUser(_ context.Context, userID pd.UserID) ([]pd.SkillProgress, error) {
	var out []pd.SkillProgress
	for _, progress := range p.skills {
		if progress.UserID == userID {
			out = append(out, progress)
		}
	}
	return out, nil
}

func key(userID pd.UserID, skillID pd.SkillID) string {
	return uuid.UUID(userID).String() + ":" + uuid.UUID(skillID).String()
}
