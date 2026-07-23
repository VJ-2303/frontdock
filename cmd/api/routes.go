package main

import (
	"net/http"

	"github.com/VJ-2303/frontdock/internal/config"
	"github.com/VJ-2303/frontdock/internal/deployments"
	"github.com/VJ-2303/frontdock/internal/httpx"
	"github.com/VJ-2303/frontdock/internal/projects"
	"github.com/VJ-2303/frontdock/internal/users"
)

func Routes(cfg *config.Config, userHandler *users.Handler, projectHandler *projects.Handler, deploymentHandler *deployments.Handler) http.Handler {
	mux := http.NewServeMux()

	verified := func(h http.HandlerFunc) http.Handler {
		return httpx.RequireAuth(cfg.JWTSecret)(
			httpx.RequireVerified(h),
		)
	}
	mux.HandleFunc("POST /auth/register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", userHandler.Login)
	mux.HandleFunc("GET /auth/verify", userHandler.VerifyEmail)

	mux.Handle("POST /projects", verified(projectHandler.CreateProjects))
	mux.Handle("GET /projects/{id}", verified(projectHandler.GetProjectHandler))

	mux.Handle("POST /projects/{id}/deployments", verified(deploymentHandler.CreateDeplyment))

	return mux
}
