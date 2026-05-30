package content_infra

import (
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	chi "github.com/go-chi/chi/v5"
)

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/courses", h.ListCourses)
	r.Get("/courses/{course_id}/sections", h.ListSections)
	r.Get("/sections/{section_id}/units", h.ListUnits)
	r.Get("/units/{unit_id}/skills", h.ListSkills)
	r.Get("/skills/{skill_id}/challenges", h.ListChallenges)

	r.Route("/editor/content", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
		}

		r.Group(func(r chi.Router) {
			r.Use(users_infra.RequireAnyRole("content_editor", "content_reviewer", "admin"))
			r.Get("/courses", h.AuthoringListCourses)
			r.Get("/courses/{course_id}/sections", h.AuthoringListSections)
			r.Get("/sections/{section_id}/units", h.AuthoringListUnits)
			r.Get("/units/{unit_id}/skills", h.AuthoringListSkills)
			r.Get("/skills/{skill_id}/challenges", h.AuthoringListChallenges)
		})

		r.Group(func(r chi.Router) {
			r.Use(users_infra.RequireAnyRole("content_editor", "admin"))
			r.Post("/courses", h.CreateCourse)
			r.Patch("/courses/{course_id}", h.UpdateCourse)
			r.Post("/sections", h.CreateSection)
			r.Patch("/sections/{section_id}", h.UpdateSection)
			r.Post("/units", h.CreateUnit)
			r.Patch("/units/{unit_id}", h.UpdateUnit)
			r.Post("/skills", h.CreateSkill)
			r.Patch("/skills/{skill_id}", h.UpdateSkill)
			r.Post("/challenges", h.CreateChallenge)
			r.Patch("/challenges/{challenge_id}", h.UpdateChallenge)
		})

		r.Group(func(r chi.Router) {
			r.Use(users_infra.RequireAnyRole("content_reviewer", "admin"))
			r.Post("/{entity}/{id}/publish", h.Publish)
		})

		r.Group(func(r chi.Router) {
			r.Use(users_infra.RequireRole("admin"))
			r.Post("/{entity}/{id}/archive", h.Archive)
		})
	})
}
