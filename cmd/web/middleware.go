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

// statusCapture wraps http.ResponseWriter to intercept 404 and 405 responses
// (both from the ServeMux and from app.clientError calls inside handlers)
// so we can replace the default plain-text body with styled HTML error pages.
type statusCapture struct {
	http.ResponseWriter
	status int
}

func (sc *statusCapture) WriteHeader(status int) {
	sc.status = status
	// Intercept 404 and 405 — don't forward to the real writer yet.
	if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
		sc.ResponseWriter.WriteHeader(status)
	}
}

func (sc *statusCapture) Write(b []byte) (int, error) {
	// If an intercepted status was set, discard the default error body.
	if sc.status == http.StatusNotFound || sc.status == http.StatusMethodNotAllowed {
		return len(b), nil
	}
	// Implicit 200 — let the underlying writer handle it.
	return sc.ResponseWriter.Write(b)
}

// errorPages intercepts 404 and 405 responses and renders styled HTML error
// pages instead of the default plain-text ones. It is safe to add to the
// outer "standard" middleware chain; any protected route that goes through
// the "dynamic" chain will still have the correct auth / session context.
func (app *application) errorPages(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc := &statusCapture{ResponseWriter: w}
		next.ServeHTTP(sc, r)

		switch sc.status {
		case http.StatusNotFound:
			// Remove the text/plain content-type set by the mux before our HTML render.
			w.Header().Del("Content-Type")
			w.Header().Del("X-Content-Type-Options")
			data := app.newTemplateData(r)
			app.render(w, r, http.StatusNotFound, "404.gohtml", data)
		case http.StatusMethodNotAllowed:
			w.Header().Del("Content-Type")
			w.Header().Del("X-Content-Type-Options")
			data := app.newTemplateData(r)
			app.render(w, r, http.StatusMethodNotAllowed, "405.gohtml", data)
		}
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
