package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/google/uuid"
)

func TestApplySessionResultCreatesProgressAndUpdatesMastery(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 1, 1)
	repo := newMemoryRepo()
	uc := NewService(Dependencies{Repository: repo, Content: graph})

	result, err := uc.ApplySessionResult(context.Background(), domain.SessionProgressInput{
		UserID:         userID,
		SkillID:        graph.skills[0].ID,
		CorrectAnswers: 1,
		TotalAnswers:   2,
		CompletedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if result.NewSkillLevel != 1 {
		t.Fatalf("expected first completed session level, got %d", result.NewSkillLevel)
	}
	if result.NewMastery != 0.15 {
		t.Fatalf("expected mastery 0.15, got %v", result.NewMastery)
	}
	if repo.courses[courseKey(userID, graph.course.ID)].Status != domain.ProgressStatusCompleted {
		t.Fatalf("expected course completed")
	}
	if repo.units[unitKey(userID, graph.units[0].ID)].Status != domain.ProgressStatusCompleted {
		t.Fatalf("expected unit completed")
	}
	if repo.skills[skillKey(userID, graph.skills[0].ID)].Status != domain.ProgressStatusCompleted {
		t.Fatalf("expected skill completed")
	}
}

func TestApplySessionResultLevelsCompletesAndUnlocksSequentially(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 2, 2)
	repo := newMemoryRepo()
	repo.skills[skillKey(userID, graph.skills[0].ID)] = domain.SkillProgress{
		UserID: userID, SkillID: graph.skills[0].ID, Status: domain.ProgressStatusInProgress, Level: 4, StartedAt: fixedTime(), UpdatedAt: fixedTime(),
	}
	uc := NewService(Dependencies{Repository: repo, Content: graph})

	result, err := uc.ApplySessionResult(context.Background(), domain.SessionProgressInput{
		UserID:         userID,
		SkillID:        graph.skills[0].ID,
		CorrectAnswers: 7,
		TotalAnswers:   10,
		CompletedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if !result.SkillCompleted || result.NewSkillLevel != 5 {
		t.Fatalf("expected completed level 5 skill, got %+v", result)
	}
	if len(result.UnlockedSkillIDs) != 1 || result.UnlockedSkillIDs[0] != graph.skills[1].ID {
		t.Fatalf("expected next skill unlock, got %+v", result.UnlockedSkillIDs)
	}
}

func TestApplySessionResultCompletesUnitAndUnlocksNextUnit(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 2, 1)
	repo := newMemoryRepo()
	repo.skills[skillKey(userID, graph.skills[0].ID)] = domain.SkillProgress{
		UserID: userID, SkillID: graph.skills[0].ID, Status: domain.ProgressStatusInProgress, Level: 4, StartedAt: fixedTime(), UpdatedAt: fixedTime(),
	}
	uc := NewService(Dependencies{Repository: repo, Content: graph})

	result, err := uc.ApplySessionResult(context.Background(), domain.SessionProgressInput{
		UserID:         userID,
		SkillID:        graph.skills[0].ID,
		CorrectAnswers: 10,
		TotalAnswers:   10,
		CompletedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if !result.UnitCompleted {
		t.Fatalf("expected unit completion, got %+v", result)
	}
	if len(result.UnlockedUnitIDs) != 1 || result.UnlockedUnitIDs[0] != graph.units[1].ID {
		t.Fatalf("expected next unit unlock, got %+v", result.UnlockedUnitIDs)
	}
	if len(result.UnlockedSkillIDs) != 1 || result.UnlockedSkillIDs[0] != graph.skills[1].ID {
		t.Fatalf("expected first skill in next unit unlock, got %+v", result.UnlockedSkillIDs)
	}
}

func TestApplySessionResultCompletesCourse(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 1, 1)
	repo := newMemoryRepo()
	repo.skills[skillKey(userID, graph.skills[0].ID)] = domain.SkillProgress{
		UserID: userID, SkillID: graph.skills[0].ID, Status: domain.ProgressStatusInProgress, Level: 4, StartedAt: fixedTime(), UpdatedAt: fixedTime(),
	}
	uc := NewService(Dependencies{Repository: repo, Content: graph})

	result, err := uc.ApplySessionResult(context.Background(), domain.SessionProgressInput{
		UserID:         userID,
		SkillID:        graph.skills[0].ID,
		CorrectAnswers: 1,
		TotalAnswers:   1,
		CompletedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if !result.CourseCompleted {
		t.Fatalf("expected course completion, got %+v", result)
	}
}

func TestCatalogProgressReturnsOnlyFirstSkillAvailableForNewUser(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 1, 3)
	uc := NewService(Dependencies{Repository: newMemoryRepo(), Content: graph})

	catalog, err := uc.GetCatalogProgress(context.Background(), uuid.UUID(userID).String(), uuid.UUID(graph.course.ID).String())
	if err != nil {
		t.Fatalf("catalog failed: %v", err)
	}
	if catalog.TotalLessons != 3 || catalog.AvailableLessons != 1 || catalog.CompletedLessons != 0 {
		t.Fatalf("unexpected catalog counts: %+v", catalog)
	}
	if catalog.Skills[0].Status != string(domain.ProgressStatusAvailable) {
		t.Fatalf("expected first skill available, got %+v", catalog.Skills[0])
	}
	if catalog.Skills[1].Status != string(domain.ProgressStatusLocked) || catalog.Skills[2].Status != string(domain.ProgressStatusLocked) {
		t.Fatalf("expected following skills locked, got %+v", catalog.Skills)
	}
}

func TestCatalogProgressUnlocksNextSkillAfterCompletedSkill(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 1, 3)
	repo := newMemoryRepo()
	repo.skills[skillKey(userID, graph.skills[0].ID)] = domain.SkillProgress{
		UserID: userID, SkillID: graph.skills[0].ID, Status: domain.ProgressStatusCompleted, Level: 1, Mastery: 0.3,
	}
	uc := NewService(Dependencies{Repository: repo, Content: graph})

	catalog, err := uc.GetCatalogProgress(context.Background(), uuid.UUID(userID).String(), uuid.UUID(graph.course.ID).String())
	if err != nil {
		t.Fatalf("catalog failed: %v", err)
	}
	if catalog.AvailableLessons != 2 || catalog.CompletedLessons != 1 {
		t.Fatalf("unexpected catalog counts: %+v", catalog)
	}
	if catalog.Skills[0].Status != string(domain.ProgressStatusCompleted) {
		t.Fatalf("expected first skill completed, got %+v", catalog.Skills[0])
	}
	if catalog.Skills[1].Status != string(domain.ProgressStatusAvailable) {
		t.Fatalf("expected second skill available, got %+v", catalog.Skills[1])
	}
	if catalog.Skills[2].Status != string(domain.ProgressStatusLocked) {
		t.Fatalf("expected third skill locked, got %+v", catalog.Skills[2])
	}
}

func TestCompleteAllForUserMarksPublishedCatalogCompleted(t *testing.T) {
	userID := domain.UserID(uuid.New())
	graph := newContentGraph(1, 2, 2)
	repo := newMemoryRepo()
	uc := NewService(Dependencies{Repository: repo, Content: graph})

	catalog, err := uc.CompleteAllForUser(context.Background(), uuid.UUID(userID).String())
	if err != nil {
		t.Fatalf("complete all failed: %v", err)
	}
	if catalog.TotalLessons != 4 || catalog.CompletedLessons != 4 || catalog.AvailableLessons != 4 {
		t.Fatalf("unexpected completed catalog: %+v", catalog)
	}
	if repo.courses[courseKey(userID, graph.course.ID)].Status != domain.ProgressStatusCompleted {
		t.Fatalf("expected course completed")
	}
	for _, unit := range graph.units {
		if repo.units[unitKey(userID, unit.ID)].Status != domain.ProgressStatusCompleted {
			t.Fatalf("expected unit completed: %s", uuid.UUID(unit.ID))
		}
	}
	for _, skill := range graph.skills {
		progress := repo.skills[skillKey(userID, skill.ID)]
		if progress.Status != domain.ProgressStatusCompleted || progress.Level != maxSkillLevel || progress.Mastery != 1 {
			t.Fatalf("expected skill completed with max mastery: %+v", progress)
		}
	}
}

type contentGraph struct {
	course    cd.Course
	section   cd.Section
	units     []cd.Unit
	skills    []cd.Skill
	byUnit    map[cd.UnitID][]cd.Skill
	bySection map[cd.SectionID][]cd.Unit
}

func newContentGraph(sectionCount, unitsPerSection, skillsPerUnit int) *contentGraph {
	course := cd.Course{ID: cd.CourseID(uuid.New()), Status: cd.ContentStatusPublished}
	graph := &contentGraph{
		course:    course,
		section:   cd.Section{ID: cd.SectionID(uuid.New()), CourseID: course.ID, Position: 1, Status: cd.ContentStatusPublished},
		byUnit:    map[cd.UnitID][]cd.Skill{},
		bySection: map[cd.SectionID][]cd.Unit{},
	}
	for u := 0; u < sectionCount*unitsPerSection; u++ {
		unit := cd.Unit{ID: cd.UnitID(uuid.New()), SectionID: graph.section.ID, Position: u + 1, Status: cd.ContentStatusPublished}
		graph.units = append(graph.units, unit)
		graph.bySection[graph.section.ID] = append(graph.bySection[graph.section.ID], unit)
		for s := 0; s < skillsPerUnit; s++ {
			skill := cd.Skill{ID: cd.SkillID(uuid.New()), UnitID: unit.ID, Position: s + 1, Status: cd.ContentStatusPublished}
			graph.skills = append(graph.skills, skill)
			graph.byUnit[unit.ID] = append(graph.byUnit[unit.ID], skill)
		}
	}
	return graph
}

func (g *contentGraph) GetSkill(_ context.Context, id cd.SkillID) (cd.Skill, error) {
	for _, skill := range g.skills {
		if skill.ID == id {
			return skill, nil
		}
	}
	return cd.Skill{}, cd.ErrNotFound
}

func (g *contentGraph) GetUnit(_ context.Context, id cd.UnitID) (cd.Unit, error) {
	for _, unit := range g.units {
		if unit.ID == id {
			return unit, nil
		}
	}
	return cd.Unit{}, cd.ErrNotFound
}

func (g *contentGraph) GetSection(_ context.Context, id cd.SectionID) (cd.Section, error) {
	if g.section.ID == id {
		return g.section, nil
	}
	return cd.Section{}, cd.ErrNotFound
}

func (g *contentGraph) ListPublishedCourses(_ context.Context) ([]cd.Course, error) {
	return []cd.Course{g.course}, nil
}

func (g *contentGraph) ListPublishedSkills(_ context.Context, unitID cd.UnitID) ([]cd.Skill, error) {
	return append([]cd.Skill(nil), g.byUnit[unitID]...), nil
}

func (g *contentGraph) ListPublishedUnits(_ context.Context, sectionID cd.SectionID) ([]cd.Unit, error) {
	return append([]cd.Unit(nil), g.bySection[sectionID]...), nil
}

func (g *contentGraph) ListPublishedSections(_ context.Context, courseID cd.CourseID) ([]cd.Section, error) {
	if g.course.ID == courseID {
		return []cd.Section{g.section}, nil
	}
	return nil, nil
}

type memoryRepo struct {
	courses map[string]domain.CourseProgress
	units   map[string]domain.UnitProgress
	skills  map[string]domain.SkillProgress
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		courses: map[string]domain.CourseProgress{},
		units:   map[string]domain.UnitProgress{},
		skills:  map[string]domain.SkillProgress{},
	}
}

func (r *memoryRepo) GetCourseProgress(_ context.Context, userID domain.UserID, courseID domain.CourseID) (*domain.CourseProgress, error) {
	progress, ok := r.courses[courseKey(userID, courseID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &progress, nil
}

func (r *memoryRepo) SaveCourseProgress(_ context.Context, progress domain.CourseProgress) error {
	r.courses[courseKey(progress.UserID, progress.CourseID)] = progress
	return nil
}

func (r *memoryRepo) GetUnitProgress(_ context.Context, userID domain.UserID, unitID domain.UnitID) (*domain.UnitProgress, error) {
	progress, ok := r.units[unitKey(userID, unitID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &progress, nil
}

func (r *memoryRepo) SaveUnitProgress(_ context.Context, progress domain.UnitProgress) error {
	r.units[unitKey(progress.UserID, progress.UnitID)] = progress
	return nil
}

func (r *memoryRepo) GetSkillProgress(_ context.Context, userID domain.UserID, skillID domain.SkillID) (*domain.SkillProgress, error) {
	progress, ok := r.skills[skillKey(userID, skillID)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &progress, nil
}

func (r *memoryRepo) SaveSkillProgress(_ context.Context, progress domain.SkillProgress) error {
	if progress.Mastery < 0 || progress.Mastery > 1 {
		return errors.New("invalid mastery")
	}
	r.skills[skillKey(progress.UserID, progress.SkillID)] = progress
	return nil
}

func (r *memoryRepo) ListSkillProgressByUser(_ context.Context, userID domain.UserID) ([]domain.SkillProgress, error) {
	var out []domain.SkillProgress
	prefix := uuid.UUID(userID).String() + ":"
	for key, progress := range r.skills {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			out = append(out, progress)
		}
	}
	return out, nil
}

func courseKey(userID domain.UserID, courseID domain.CourseID) string {
	return uuid.UUID(userID).String() + ":" + uuid.UUID(courseID).String()
}

func unitKey(userID domain.UserID, unitID domain.UnitID) string {
	return uuid.UUID(userID).String() + ":" + uuid.UUID(unitID).String()
}

func skillKey(userID domain.UserID, skillID domain.SkillID) string {
	return uuid.UUID(userID).String() + ":" + uuid.UUID(skillID).String()
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
}
