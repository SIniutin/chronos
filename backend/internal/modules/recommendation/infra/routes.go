package infra

import chi "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/recommendations", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
		}
		r.Get("/next", h.GetNext)
	})
}
