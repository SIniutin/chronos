package infra

import (
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	chi "github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/progress", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
		}
		r.Get("/catalog", h.GetCatalog)
	})
	r.Route("/admin/progress", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
			r.Use(users_infra.RequireRole("admin"))
		}
		r.Post("/users/{user_id}/complete-all", h.CompleteAllForUser)
	})
}
