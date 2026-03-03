package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/migrate"
	"github.com/ntiGideon/ent/user"
	"github.com/ntiGideon/internal/permissions"

	_ "github.com/lib/pq"
)

type templateData struct {
	Form            any
	Flash           string
	FlashError      string
	Toast           map[string]interface{}
	CurrentYear     int
	CSRFToken       string
	IsAuthenticated bool
	IsAdmin         bool // true for super_admin and branch_admin (matches adminOnly middleware)
	LoggedInUser    *ent.User
	Permissions     map[string]bool
	Data            any
}

func connectDB(connectionString string) (*ent.Client, error) {
	client, err := ent.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	err = client.Schema.Create(ctx, migrate.WithGlobalUniqueID(true))
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to the database")
	return client, nil
}

func OpenDriver(connectionString string) (*sql.DB, error) {
	drv, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return drv, nil
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		return err
	}
	return nil
}

func (app *application) newTemplateData(r *http.Request) templateData {
	u := app.getAuthenticatedUser(r)
	isAdmin := u != nil && (u.Role == user.RoleSuperAdmin || u.Role == user.RoleBranchAdmin)
	return templateData{
		CurrentYear:     time.Now().Year(),
		CSRFToken:       nosurf.Token(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		FlashError:      app.sessionManager.PopString(r.Context(), "flash_error"),
		IsAuthenticated: app.isAuthenticated(r),
		IsAdmin:         isAdmin,
		LoggedInUser:    u,
		Permissions:     resolvePermissions(u),
		Form:            make(map[string]interface{}),
		Data:            make(map[string]interface{}),
	}
}

// resolvePermissions builds the permission map for the given user.
// Admin roles receive all permissions. Custom roles parse their JSON permission list.
// All other roles fall back to DefaultPermissionsForRole.
func resolvePermissions(u *ent.User) map[string]bool {
	if u == nil {
		return map[string]bool{}
	}
	// Admins always have everything
	if u.Role == user.RoleSuperAdmin || u.Role == user.RoleBranchAdmin {
		return permissions.DefaultPermissionsForRole(u.Role.String())
	}
	// Custom role — parse JSON
	if u.Edges.CustomRole != nil {
		return parsePermissionJSON(u.Edges.CustomRole.Permissions)
	}
	// Fall back to built-in role defaults
	return permissions.DefaultPermissionsForRole(u.Role.String())
}

// parsePermissionJSON converts a JSON array of permission key strings into a bool map.
func parsePermissionJSON(raw string) map[string]bool {
	var keys []string
	_ = json.Unmarshal([]byte(raw), &keys)
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (app *application) getAuthenticatedUser(r *http.Request) *ent.User {
	u, _ := r.Context().Value(userContextKey).(*ent.User)
	return u
}

// churchIDFromSession returns the church ID stored in session (0 for superadmin).
func (app *application) churchIDFromSession(r *http.Request) int {
	return app.sessionManager.GetInt(r.Context(), "authenticatedChurchID")
}
