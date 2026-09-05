package http

import (
	"errors"
	"faceit_stats_wrapper/internal/service"
	"net/http"
)

type Handler struct {
	svc service.StatsService
}

func NewHandler(svc service.StatsService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) InitRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /player/{nickname}", h.GetPlayer)
	mux.HandleFunc("GET /player/{nickname}/match", h.GetLastMatch)
	mux.HandleFunc("GET /player/{nickname}/matches", h.GetPlayerMatches)
}

func (h *Handler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	nickname := r.PathValue("nickname")
	if nickname == "" {
		writeError(w, http.StatusBadRequest, "nickname is required!")
		return
	}

	player, err := h.svc.GetPlayer(r.Context(), nickname)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, player)
}

func (h *Handler) GetLastMatch(w http.ResponseWriter, r *http.Request) {
	nickname := r.PathValue("nickname")
	if nickname == "" {
		writeError(w, http.StatusBadRequest, "nickname is required!")
		return
	}

	match, err := h.svc.GetLastMatch(r.Context(), nickname)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, match)
}

func (h *Handler) GetPlayerMatches(w http.ResponseWriter, r *http.Request) {
	nickname := r.PathValue("nickname")
	if nickname == "" {
		writeError(w, http.StatusBadRequest, "nickname is required!")
		return
	}

	matches, err := h.svc.GetPlayerMatches(r.Context(), nickname)
	if err != nil {
		if errors.Is(err, service.ErrPlayerNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, matches)
}
