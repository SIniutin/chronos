package usecase

import "github.com/SIniutin/history-app-backend/internal/modules/content/domain"

type Service struct {
	coursesRepo   domain.CoursesRepository
	sectionsRepo  domain.SectionsRepository
	unitsRepo     domain.UnitsRepository
	skillsRepo    domain.SkillsRepository
	challengeRepo domain.ChallengeRepository
	statusRepo    domain.StatusRepository
}

type Dependencies struct {
	Courses    domain.CoursesRepository
	Sections   domain.SectionsRepository
	Units      domain.UnitsRepository
	Skills     domain.SkillsRepository
	Challenges domain.ChallengeRepository
	Status     domain.StatusRepository
}

func NewService(deps Dependencies) *Service {
	return &Service{
		coursesRepo:   deps.Courses,
		sectionsRepo:  deps.Sections,
		unitsRepo:     deps.Units,
		skillsRepo:    deps.Skills,
		challengeRepo: deps.Challenges,
		statusRepo:    deps.Status,
	}
}

func NewServiceFromRepository(repo interface {
	domain.CoursesRepository
	domain.SectionsRepository
	domain.UnitsRepository
	domain.SkillsRepository
	domain.ChallengeRepository
	domain.StatusRepository
}) *Service {
	return NewService(Dependencies{
		Courses:    repo,
		Sections:   repo,
		Units:      repo,
		Skills:     repo,
		Challenges: repo,
		Status:     repo,
	})
}
