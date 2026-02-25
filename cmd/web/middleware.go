package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/ntiGideon/ent/user"
	"github.com/ntiGideon/internal/models"
)

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info("request",
			"ip", r.RemoteAddr,
			"proto", r.Proto,
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)
		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // set to true behind HTTPS
	})
	// Use actual TLS state so origin checks compare http:// against http:// locally
	csrfHandler.SetIsTLSFunc(func(r *http.Request) bool { return r.TLS != nil })
	return csrfHandler
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// authenticate loads the logged-in user from the session into the request context.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		u, err := app.userModel.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				app.sessionManager.Remove(r.Context(), "authenticatedUserID")
				next.ServeHTTP(w, r)
				return
			}
			app.serverError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), isAuthenticatedKey, true)
		ctx = context.WithValue(ctx, userContextKey, u)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// requireAuthentication redirects unauthenticated users to the login page.
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			app.sessionManager.Put(r.Context(), "requestedPathAfterLogin", r.URL.Path)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// requireSuperAdmin allows only super_admin role.
func (app *application) requireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := app.getAuthenticatedUser(r)
		if u == nil || u.Role != user.RoleSuperAdmin {
			app.clientError(w, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin allows super_admin and branch_admin roles.
func (app *application) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := app.getAuthenticatedUser(r)
		if u == nil || (u.Role != user.RoleSuperAdmin && u.Role != user.RoleBranchAdmin) {
			app.clientError(w, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
