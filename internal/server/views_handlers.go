package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/agentplexus/multi-agent-designer/internal/ent"
	"github.com/agentplexus/multi-agent-designer/internal/ent/schema"
	"github.com/agentplexus/multi-agent-designer/internal/ent/view"
)

// ViewRequest is the API request body for creating/updating views.
type ViewRequest struct {
	Name          string                                `json:"name"`
	Teams         []string                              `json:"teams"`
	NodePositions map[string]map[string]schema.Position `json:"node_positions"`
	LeftWidth     int                                   `json:"left_width"`
	RightWidth    int                                   `json:"right_width"`
}

// ViewResponse is the API response for a view.
type ViewResponse struct {
	ID            string                                `json:"id"`
	Name          string                                `json:"name"`
	Teams         []string                              `json:"teams"`
	NodePositions map[string]map[string]schema.Position `json:"node_positions"`
	LeftWidth     int                                   `json:"left_width"`
	RightWidth    int                                   `json:"right_width"`
	CreatedAt     string                                `json:"created_at"`
	UpdatedAt     string                                `json:"updated_at"`
}

func viewToResponse(v *ent.View) ViewResponse {
	return ViewResponse{
		ID:            v.ID.String(),
		Name:          v.Name,
		Teams:         v.Teams,
		NodePositions: v.NodePositions,
		LeftWidth:     v.LeftWidth,
		RightWidth:    v.RightWidth,
		CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *Server) handleListViews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	views, err := s.db.View.Query().
		Order(ent.Desc(view.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		http.Error(w, "failed to list views", http.StatusInternalServerError)
		return
	}

	result := make([]ViewResponse, len(views))
	for i, v := range views {
		result[i] = viewToResponse(v)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	v, err := s.db.View.Create().
		SetName(req.Name).
		SetTeams(req.Teams).
		SetNodePositions(req.NodePositions).
		SetLeftWidth(req.LeftWidth).
		SetRightWidth(req.RightWidth).
		Save(ctx)
	if err != nil {
		http.Error(w, "failed to create view", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, viewToResponse(v))
}

func (s *Server) handleGetView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid view id", http.StatusBadRequest)
		return
	}

	v, err := s.db.View.Get(ctx, id)
	if ent.IsNotFound(err) {
		http.Error(w, "view not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get view", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, viewToResponse(v))
}

func (s *Server) handleUpdateView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid view id", http.StatusBadRequest)
		return
	}

	var req ViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	v, err := s.db.View.UpdateOneID(id).
		SetName(req.Name).
		SetTeams(req.Teams).
		SetNodePositions(req.NodePositions).
		SetLeftWidth(req.LeftWidth).
		SetRightWidth(req.RightWidth).
		Save(ctx)
	if ent.IsNotFound(err) {
		http.Error(w, "view not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to update view", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, viewToResponse(v))
}

func (s *Server) handleDeleteView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid view id", http.StatusBadRequest)
		return
	}

	err = s.db.View.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		http.Error(w, "view not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to delete view", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
