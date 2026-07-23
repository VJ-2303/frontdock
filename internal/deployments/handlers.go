package deployments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/VJ-2303/frontdock/internal/config"
	"github.com/VJ-2303/frontdock/internal/httpx"
	"github.com/VJ-2303/frontdock/internal/projects"
	"github.com/VJ-2303/frontdock/internal/queue"
	"github.com/VJ-2303/frontdock/internal/storage"
	"github.com/google/uuid"
)

type Handler struct {
	deployments *Store
	projects    *projects.Store
	cfg         *config.Config
	storage     *storage.Storage
	pub         *queue.Publisher
}

func NewHandler(deployment *Store, projects *projects.Store, cfg *config.Config, storage *storage.Storage, pub *queue.Publisher) *Handler {
	return &Handler{
		deployments: deployment,
		projects:    projects,
		cfg:         cfg,
		storage:     storage,
		pub:         pub,
	}
}

func (h *Handler) CreateDeplyment(w http.ResponseWriter, r *http.Request) {
	u, _ := httpx.UserFrom(r.Context())

	projectID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "invalid project id")
		return
	}
	proj, err := h.projects.GetOwnedByID(r.Context(), projectID, u.ID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "not_found", "project not found")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)

	mr, err := r.MultipartReader()
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "expected a multipart/for-data body")
		return
	}
	part, err := mr.NextPart()
	if err != nil || part.FormName() != "file" {
		httpx.Error(w, 400, "bad_request", `expected a form field named "file"`)
		return
	}
	defer part.Close()

	if !strings.HasSuffix(strings.ToLower(part.FileName()), "zip") {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_failed", "upload must be a zip file")
		return
	}

	deploymentID := uuid.New()
	uploadKey := fmt.Sprintf("uploads/%s.zip", deploymentID)

	err = h.storage.Put(r.Context(), h.storage.BucketUploads, uploadKey, part, -1, "application/zip")
	if err != nil {
		slog.Error("error uploading", "err", err.Error())
		_ = h.storage.Delete(context.WithoutCancel(r.Context()), h.storage.BucketUploads, uploadKey)

		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			httpx.Error(w, 413, "payload_too_large",
				fmt.Sprintf("zip must be under %d bytes",
					h.cfg.MaxUploadBytes))
			return
		}
		httpx.Error(w, 500, "storage_error", "could not store upload")
		return
	}
	d, err := h.deployments.CreateWithID(r.Context(), deploymentID, projectID, uploadKey)
	if err != nil {
		slog.Error("error in DB", "err", err.Error())
		_ = h.storage.Delete(r.Context(), h.storage.BucketUploads, uploadKey)
		httpx.Error(w, 500, "internal_error", "could not create deployment")
		return
	}

	err = h.pub.Publish(r.Context(), queue.RoutingDeployRequested, queue.DeployMessage{
		Type:         "deploy.requested",
		DeploymentID: d.ID.String(),
		ProjectID:    proj.ID.String(),
		UploadKey:    uploadKey,
	})

	if err != nil {
		slog.Error("publish failed; marking deployment failed", "deployment_id", d.ID, "err", err)
		_ = h.deployments.MarkFailed(r.Context(), d.ID, "could not queue deployment job")
		httpx.Error(w, http.StatusServiceUnavailable, "queue_unavailable", "could not queue the deployment, please retry later")
		return
	}
	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"id":         d.ID,
		"version":    d.Version,
		"status":     d.Status,
		"status_url": fmt.Sprintf("/projects/%s/deployments/%s", proj.ID, d.ID),
	})
}
