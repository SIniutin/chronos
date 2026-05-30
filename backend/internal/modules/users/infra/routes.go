package users_infra

import chi "github.com/go-chi/chi/v5"

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)

		r.Group(func(r chi.Router) {
			if h.auth != nil {
				r.Use(h.auth.Auth)
			}
			r.Post("/logout", h.Logout)
		})
	})

	r.Route("/users", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
		}
		r.Get("/me", h.GetMe)
		r.Patch("/me", h.UpdateMe)
		r.Delete("/me", h.DeleteMe)
	})

	r.Route("/admin/users", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
			r.Use(RequireRole("admin"))
		}
		r.Get("/lookup", h.FindUser)
		r.Patch("/{user_id}/role", h.ChangeUserRole)
	})
}
