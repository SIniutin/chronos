package usecase

import (
	"context"
	"errors"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	progress_domain "github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/SIniutin/history-app-backend/internal/modules/recommendation/domain"
)

const defaultChallengeLimit = 10
const lowMasteryThreshold = 0.6

type Service struct {
	content  ContentPort
	progress ProgressPort
}

type ContentPort interface {
	ListPublishedSections(ctx context.Context, courseID content_domain.CourseID) ([]content_domain.Section, error)
	ListPublishedUnits(ctx context.Context, sectionID content_domain.SectionID) ([]content_domain.Unit, error)
	ListPublishedSkills(ctx context.Context, unitID content_domain.UnitID) ([]content_domain.Skill, error)
	ListPublishedChallenges(ctx context.Context, skillID content_domain.SkillID) ([]content_domain.Challenge, error)
}

type ProgressPort interface {
	GetSkillProgress(ctx context.Context, userID progress_domain.UserID, skillID progress_domain.SkillID) (*progress_domain.SkillProgress, error)
	ListSkillProgressByUser(ctx context.Context, userID progress_domain.UserID) ([]progress_domain.SkillProgress, error)
}

func NewService(content ContentPort, progress ProgressPort) *Service {
	return &Service{content: content, progress: progress}
}

func (s *Service) GetNextSkill(ctx context.Context, userID domain.UserID, courseID domain.CourseID) (*domain.Recommendation, error) {
	items, err := s.courseSkills(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, content_domain.ErrNotFound
	}
	var completed []skillItem
	var first *skillItem
	var firstUnlocked *skillItem
	previousCompleted := true
	for i := range items {
		item := items[i]
		if first == nil {
			first = &item
		}
		progress, err := s.progress.GetSkillProgress(ctx, userID, item.skill.ID)
		if err != nil {
			if errors.Is(err, progress_domain.ErrNotFound) {
				if previousCompleted && firstUnlocked == nil {
					firstUnlocked = &item
				}
				previousCompleted = false
				continue
			}
			return nil, err
		}
		switch progress.Status {
		case progress_domain.ProgressStatusInProgress:
			return rec(domain.RecommendationTypeContinue, courseID, item.unit.ID, item.skill.ID, "continue in-progress skill"), nil
		case progress_domain.ProgressStatusAvailable:
			if progress.Mastery > 0 && progress.Mastery < lowMasteryThreshold {
				return rec(domain.RecommendationTypeReview, courseID, item.unit.ID, item.skill.ID, "review low mastery skill"), nil
			}
			return rec(domain.RecommendationTypeNewSkill, courseID, item.unit.ID, item.skill.ID, "start available skill"), nil
		case progress_domain.ProgressStatusCompleted:
			completed = append(completed, item)
			previousCompleted = true
		default:
			previousCompleted = false
		}
	}
	if first != nil && len(completed) == 0 {
		return rec(domain.RecommendationTypeNewSkill, courseID, first.unit.ID, first.skill.ID, "start first skill"), nil
	}
	if firstUnlocked != nil {
		return rec(domain.RecommendationTypeNewSkill, courseID, firstUnlocked.unit.ID, firstUnlocked.skill.ID, "start unlocked skill"), nil
	}
	if len(completed) > 0 {
		best := completed[0]
		bestMastery := 2.0
		for _, item := range completed {
			p, err := s.progress.GetSkillProgress(ctx, userID, item.skill.ID)
			if err == nil && p.Mastery < bestMastery {
				best = item
				bestMastery = p.Mastery
			}
		}
		return rec(domain.RecommendationTypeReview, courseID, best.unit.ID, best.skill.ID, "review completed course"), nil
	}
	return rec(domain.RecommendationTypeNewSkill, courseID, first.unit.ID, first.skill.ID, "start next skill"), nil
}

func (s *Service) PickChallengesForSession(ctx context.Context, _ domain.UserID, skillID domain.SkillID, limit int) ([]domain.ChallengeID, error) {
	if limit <= 0 {
		limit = defaultChallengeLimit
	}
	challenges, err := s.content.ListPublishedChallenges(ctx, skillID)
	if err != nil {
		return nil, err
	}
	if len(challenges) > limit {
		challenges = challenges[:limit]
	}
	out := make([]domain.ChallengeID, 0, len(challenges))
	for _, c := range challenges {
		out = append(out, c.ID)
	}
	return out, nil
}

type skillItem struct {
	unit  content_domain.Unit
	skill content_domain.Skill
}

func (s *Service) courseSkills(ctx context.Context, courseID content_domain.CourseID) ([]skillItem, error) {
	sections, err := s.content.ListPublishedSections(ctx, courseID)
	if err != nil {
		return nil, err
	}
	var out []skillItem
	for _, section := range sections {
		units, err := s.content.ListPublishedUnits(ctx, section.ID)
		if err != nil {
			return nil, err
		}
		for _, unit := range units {
			skills, err := s.content.ListPublishedSkills(ctx, unit.ID)
			if err != nil {
				return nil, err
			}
			for _, skill := range skills {
				out = append(out, skillItem{unit: unit, skill: skill})
			}
		}
	}
	return out, nil
}

func rec(t domain.RecommendationType, courseID domain.CourseID, unitID domain.UnitID, skillID domain.SkillID, reason string) *domain.Recommendation {
	return &domain.Recommendation{Type: t, CourseID: courseID, UnitID: &unitID, SkillID: &skillID, Reason: reason}
}
