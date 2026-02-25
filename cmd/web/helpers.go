package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/migrate"

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
	LoggedInUser    *ent.User
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
	return templateData{
		CurrentYear:     time.Now().Year(),
		CSRFToken:       nosurf.Token(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		FlashError:      app.sessionManager.PopString(r.Context(), "flash_error"),
		IsAuthenticated: app.isAuthenticated(r),
		LoggedInUser:    app.getAuthenticatedUser(r),
		Form:            make(map[string]interface{}),
		Data:            make(map[string]interface{}),
	}
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
