package handlers

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/service"
)

type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(svc *service.TeamService) *TeamHandler {
	return &TeamHandler{service: svc}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team models.Team

	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeError(w, http.StatusBadRequest, models.ErrorNotFound, "Invalid request body")
		return
	}

	result, err := h.service.CreateTeam(r.Context(), &team)
	if err != nil {
		if err.Error() == "team already exists" {
			writeError(w, http.StatusBadRequest, models.ErrorTeamExists, "team_name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, models.ErrorNotFound, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"team": result})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, http.StatusBadRequest, models.ErrorNotFound, "team_name parameter is required")
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		writeError(w, http.StatusNotFound, models.ErrorNotFound, "resource not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(team)
}