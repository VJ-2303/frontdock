package projects

import (
	"errors"
	"net/http"
	"strings"

	"github.com/VJ-2303/frontdock/internal/httpx"
	"github.com/google/uuid"
)

type Handler struct {
	projects *Store
}

func NewHandler(projects *Store) *Handler {
	return &Handler{
		projects: projects,
	}
}

func (h *Handler) CreateProjects(w http.ResponseWriter, r *http.Request) {
	u, _ := httpx.UserFrom(r.Context())

	var CreateProjectReq struct {
		Name      string `json:"name"`
		Subdomain string `json:"subdomain"`
	}

	if err := httpx.Decode(w, r, &CreateProjectReq); err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "Please check the request body")
		return
	}

	CreateProjectReq.Name = strings.TrimSpace(CreateProjectReq.Name)
	if CreateProjectReq.Name == "" || len(CreateProjectReq.Name) > 100 {
		httpx.Error(w, http.StatusUnprocessableEntity, "bad_request", "name must be 1-100 characters.")
		return
	}

	if CreateProjectReq.Subdomain != "" {
		sub := strings.ToLower(strings.TrimSpace(CreateProjectReq.Subdomain))
		if err := ValidateSubdomains(sub); err != nil {
			httpx.Error(w, http.StatusUnprocessableEntity, "bad_request", err.Error())
			return
		}
		project, err := h.projects.Create(r.Context(), u.ID, CreateProjectReq.Name, sub)
		if err != nil {
			if errors.Is(err, ErrSubdomainTaken) {
				httpx.Error(w, http.StatusConflict, "subdomain_taken", err.Error())
				return
			}
			if errors.Is(err, ErrNameTaken) {
				httpx.Error(w, http.StatusConflict, "name_taken", err.Error())
				return
			}
			httpx.Error(w, http.StatusInternalServerError, "internal_error", "internal Server error")
			return
		}
		httpx.JSON(w, http.StatusCreated, project)
		return
	}
	for range 5 {
		sub := GenerateSubdomain()
		project, err := h.projects.Create(r.Context(), u.ID, CreateProjectReq.Name, sub)
		if err != nil {
			if errors.Is(err, ErrSubdomainTaken) {
				continue
			}
			if errors.Is(err, ErrNameTaken) {
				httpx.Error(w, http.StatusConflict, "name_taken", err.Error())
				return
			}
			httpx.Error(w, http.StatusInternalServerError, "internal_error", "internal Server error")
			return
		}
		httpx.JSON(w, http.StatusCreated, project)
		return
	}
	httpx.Error(w, http.StatusServiceUnavailable, "subdomain_generation_failed", "could not allocate a subdomain, try again")
}

func (h *Handler) GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project_id, err := uuid.Parse(id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	u, _ := httpx.UserFrom(r.Context())

	p, err := h.projects.GetOwnedByID(r.Context(), project_id, u.ID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			httpx.Error(w, http.StatusNotFound, "not_found", "project not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "internal_server", "the server encountered an error")
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}
