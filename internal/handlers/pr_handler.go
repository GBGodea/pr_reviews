package handlers

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/service"
)

type PRHandler struct {
	service *service.PRService
}

func NewPRHandler(svc *service.PRService) *PRHandler {
	return &PRHandler{service: svc}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PRId     string `json:"pull_request_id"`
		PRName   string `json:"pull_request_name"`
		AuthorId string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, models.ErrorNotFound, "Invalid request body")
		return
	}

	pr, err := h.service.CreatePR(r.Context(), req.PRId, req.PRName, req.AuthorId)
	if err != nil {
		if err.Error() == "author not found" {
			writeError(w, http.StatusNotFound, models.ErrorNotFound, "resource not found")
			return
		}
		if err.Error() == "PR already exists" {
			writeError(w, http.StatusConflict, models.ErrorPRExists, "PR id already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, models.ErrorNotFound, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"pr": pr})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PRId string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, models.ErrorNotFound, "Invalid request body")
		return
	}

	pr, err := h.service.MergePR(r.Context(), req.PRId)
	if err != nil {
		writeError(w, http.StatusNotFound, models.ErrorNotFound, "resource not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"pr": pr})
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PRId      string `json:"pull_request_id"`
		OldUserId string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, models.ErrorNotFound, "Invalid request body")
		return
	}

	pr, newReviewerId, err := h.service.ReassignReviewer(r.Context(), req.PRId, req.OldUserId)
	if err != nil {
		switch err.Error() {
		case "PR is merged":
			writeError(w, http.StatusConflict, models.ErrorPRMerged, "cannot reassign on merged PR")
		case "reviewer is not assigned":
			writeError(w, http.StatusConflict, models.ErrorNotAssigned, "reviewer is not assigned to this PR")
		case "no active replacement candidate in team":
			writeError(w, http.StatusConflict, models.ErrorNoCandidate, "no active replacement candidate in team")
		default:
			writeError(w, http.StatusNotFound, models.ErrorNotFound, "resource not found")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerId,
	})
}