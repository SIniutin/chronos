package infra

import chi "github.com/go-chi/chi/v5"

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/learning", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
		}

		r.Post("/sessions", h.StartSession)
		r.Get("/sessions/{session_id}/current", h.GetCurrentChallenge)
		r.Post("/sessions/{session_id}/answer", h.SubmitAnswer)
		r.Post("/sessions/{session_id}/finish", h.FinishSession)
	})
}
