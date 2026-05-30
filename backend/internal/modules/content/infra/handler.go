package content_infra

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	content_usecase "github.com/SIniutin/history-app-backend/internal/modules/content/usecase"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	courses    content_usecase.CoursesUsecase
	sections   content_usecase.SectionsUsecase
	units      content_usecase.UnitsUsecase
	skills     content_usecase.SkillsUsecase
	challenges content_usecase.ChallengesUsecase
	auth       *users_infra.AuthMiddleware
}

type Dependencies struct {
	Courses    content_usecase.CoursesUsecase
	Sections   content_usecase.SectionsUsecase
	Units      content_usecase.UnitsUsecase
	Skills     content_usecase.SkillsUsecase
	Challenges content_usecase.ChallengesUsecase
	Auth       *users_infra.AuthMiddleware
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		courses:    deps.Courses,
		sections:   deps.Sections,
		units:      deps.Units,
		skills:     deps.Skills,
		challenges: deps.Challenges,
		auth:       deps.Auth,
	}
}

func (h *Handler) ListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courses.ListPublishedCourses(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(courses, content_usecase.ToAPICourse))
}

func (h *Handler) ListSections(w http.ResponseWriter, r *http.Request) {
	sections, err := h.sections.ListPublishedSections(r.Context(), content_api.ListSectionsInput{
		CourseID: chi.URLParam(r, "course_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(sections, content_usecase.ToAPISection))
}

func (h *Handler) ListUnits(w http.ResponseWriter, r *http.Request) {
	units, err := h.units.ListPublishedUnits(r.Context(), content_api.ListUnitsInput{
		SectionID: chi.URLParam(r, "section_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(units, content_usecase.ToAPIUnit))
}

func (h *Handler) ListSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.skills.ListPublishedSkills(r.Context(), content_api.ListSkillsInput{
		UnitID: chi.URLParam(r, "unit_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(skills, content_usecase.ToAPISkill))
}

func (h *Handler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	challenges, err := h.challenges.ListPublishedChallenges(r.Context(), content_api.ListChallengesInput{
		SkillID: chi.URLParam(r, "skill_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(challenges, content_usecase.ToAPIChallenge))
}

func (h *Handler) AuthoringListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courses.ListAllCourses(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(courses, content_usecase.ToAPICourse))
}

func (h *Handler) AuthoringListSections(w http.ResponseWriter, r *http.Request) {
	sections, err := h.sections.ListAllSections(r.Context(), content_api.ListSectionsInput{
		CourseID: chi.URLParam(r, "course_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(sections, content_usecase.ToAPISection))
}

func (h *Handler) AuthoringListUnits(w http.ResponseWriter, r *http.Request) {
	units, err := h.units.ListAllUnits(r.Context(), content_api.ListUnitsInput{
		SectionID: chi.URLParam(r, "section_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(units, content_usecase.ToAPIUnit))
}

func (h *Handler) AuthoringListSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.skills.ListAllSkills(r.Context(), content_api.ListSkillsInput{
		UnitID: chi.URLParam(r, "unit_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(skills, content_usecase.ToAPISkill))
}

func (h *Handler) AuthoringListChallenges(w http.ResponseWriter, r *http.Request) {
	challenges, err := h.challenges.ListAllChallenges(r.Context(), content_api.ListChallengesInput{
		SkillID: chi.URLParam(r, "skill_id"),
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapSlice(challenges, content_usecase.ToAPIAuthoringChallenge))
}

func (h *Handler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var req content_api.CourseWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	res, err := h.courses.CreateCourse(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPICourse(res), err, http.StatusCreated)
}

func (h *Handler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	var req content_api.CourseWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	req.ID = chi.URLParam(r, "course_id")
	res, err := h.courses.UpdateCourse(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPICourse(res), err, http.StatusOK)
}

func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	var req content_api.SectionWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	res, err := h.sections.CreateSection(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPISection(res), err, http.StatusCreated)
}

func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	var req content_api.SectionWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	req.ID = chi.URLParam(r, "section_id")
	res, err := h.sections.UpdateSection(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPISection(res), err, http.StatusOK)
}

func (h *Handler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	var req content_api.UnitWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	res, err := h.units.CreateUnit(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPIUnit(res), err, http.StatusCreated)
}

func (h *Handler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	var req content_api.UnitWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	req.ID = chi.URLParam(r, "unit_id")
	res, err := h.units.UpdateUnit(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPIUnit(res), err, http.StatusOK)
}

func (h *Handler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	var req content_api.SkillWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	res, err := h.skills.CreateSkill(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPISkill(res), err, http.StatusCreated)
}

func (h *Handler) UpdateSkill(w http.ResponseWriter, r *http.Request) {
	var req content_api.SkillWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	req.ID = chi.URLParam(r, "skill_id")
	res, err := h.skills.UpdateSkill(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPISkill(res), err, http.StatusOK)
}

func (h *Handler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	var req content_api.ChallengeWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	res, err := h.challenges.CreateChallenge(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPIAuthoringChallenge(res), err, http.StatusCreated)
}

func (h *Handler) UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	var req content_api.ChallengeWriteInput
	if !decodeAuthoring(w, r, &req) {
		return
	}
	req.ID = chi.URLParam(r, "challenge_id")
	res, err := h.challenges.UpdateChallenge(r.Context(), req)
	writeAuthoringResult(w, content_usecase.ToAPIAuthoringChallenge(res), err, http.StatusOK)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	input, ok := transitionInput(w, r)
	if !ok {
		return
	}
	if err := h.transitionUsecase(input.Entity).Publish(r.Context(), input); err != nil {
		writeUsecaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	input, ok := transitionInput(w, r)
	if !ok {
		return
	}
	if err := h.transitionUsecase(input.Entity).Archive(r.Context(), input); err != nil {
		writeUsecaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type transitionUsecase interface {
	Publish(ctx context.Context, input content_api.StatusTransitionInput) error
	Archive(ctx context.Context, input content_api.StatusTransitionInput) error
}

func (h *Handler) transitionUsecase(entity string) transitionUsecase {
	switch strings.TrimSpace(entity) {
	case "courses", "course":
		return h.courses
	case "sections", "section":
		return h.sections
	case "units", "unit":
		return h.units
	case "skills", "skill":
		return h.skills
	case "challenges", "challenge":
		return h.challenges
	default:
		return invalidTransitionUsecase{}
	}
}

type invalidTransitionUsecase struct{}

func (invalidTransitionUsecase) Publish(context.Context, content_api.StatusTransitionInput) error {
	return content_api.ErrInvalidInput
}

func (invalidTransitionUsecase) Archive(context.Context, content_api.StatusTransitionInput) error {
	return content_api.ErrInvalidInput
}

func decodeAuthoring(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return false
	}
	userID, ok := users_infra.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return false
	}
	switch input := v.(type) {
	case *content_api.CourseWriteInput:
		input.ActorID = userID
	case *content_api.SectionWriteInput:
		input.ActorID = userID
	case *content_api.UnitWriteInput:
		input.ActorID = userID
	case *content_api.SkillWriteInput:
		input.ActorID = userID
	case *content_api.ChallengeWriteInput:
		input.ActorID = userID
	default:
		writeError(w, http.StatusInternalServerError, "unsupported authoring input")
		return false
	}
	return true
}

func transitionInput(w http.ResponseWriter, r *http.Request) (content_api.StatusTransitionInput, bool) {
	userID, ok := users_infra.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return content_api.StatusTransitionInput{}, false
	}
	return content_api.StatusTransitionInput{
		Entity:  chi.URLParam(r, "entity"),
		ID:      chi.URLParam(r, "id"),
		ActorID: userID,
	}, true
}

func writeAuthoringResult(w http.ResponseWriter, v any, err error, status int) {
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, status, v)
}

func mapSlice[T any, U any](items []T, mapper func(T) U) []U {
	out := make([]U, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, content_api.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, content_api.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
