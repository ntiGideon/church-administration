package main

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/ntiGideon/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// Middleware chains
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	protected := dynamic.Append(app.requireAuthentication)
	adminOnly := protected.Append(app.requireAdmin)
	superAdminOnly := protected.Append(app.requireSuperAdmin)

	// Homepage
	mux.Handle("GET /", dynamic.ThenFunc(app.home))

	// Auth
	mux.Handle("GET /login", dynamic.ThenFunc(app.loginGet))
	mux.Handle("POST /login", dynamic.ThenFunc(app.loginPost))
	mux.Handle("POST /logout", protected.ThenFunc(app.logoutPost))

	// Church registration via invitation token
	mux.Handle("GET /register", dynamic.ThenFunc(app.registerGet))
	mux.Handle("POST /register", dynamic.ThenFunc(app.registerPost))

	// Member/staff invitation accept
	mux.Handle("GET /invite/accept", dynamic.ThenFunc(app.inviteAcceptGet))
	mux.Handle("POST /invite/accept", dynamic.ThenFunc(app.inviteAcceptPost))

	// Dashboard
	mux.Handle("GET /dashboard", protected.ThenFunc(app.dashboard))

	// Profile
	mux.Handle("GET /profile", protected.ThenFunc(app.profileGet))
	mux.Handle("POST /profile", protected.ThenFunc(app.profilePost))
	mux.Handle("POST /profile/avatar", protected.ThenFunc(app.profileAvatarPost))

	// Church settings (branch admin manages their church)
	mux.Handle("GET /church/settings", protected.ThenFunc(app.churchSettings))
	mux.Handle("POST /church/settings", protected.ThenFunc(app.churchSettingsPost))
	mux.Handle("POST /church/settings/logo", adminOnly.ThenFunc(app.churchLogoPost))

	// Congregation members (Contact records — no system account required)
	mux.Handle("GET /members", protected.ThenFunc(app.membersList))
	mux.Handle("GET /members/new", protected.ThenFunc(app.memberNewGet))
	mux.Handle("POST /members/new", protected.ThenFunc(app.memberNewPost))
	mux.Handle("GET /members/{id}", protected.ThenFunc(app.memberDetail))
	mux.Handle("GET /members/{id}/edit", protected.ThenFunc(app.memberEditGet))
	mux.Handle("POST /members/{id}/edit", protected.ThenFunc(app.memberEditPost))
	mux.Handle("POST /members/{id}/avatar", protected.ThenFunc(app.memberAvatarPost))
	mux.Handle("POST /members/{id}/delete", protected.ThenFunc(app.memberDelete))

	// Workers (system users — invited by admin, can log in)
	mux.Handle("GET /workers", protected.ThenFunc(app.workersGet))
	mux.Handle("GET /workers/new", adminOnly.ThenFunc(app.workerInviteGet))
	mux.Handle("POST /workers/new", adminOnly.ThenFunc(app.workerInvitePost))

	// Events
	mux.Handle("GET /events", protected.ThenFunc(app.eventsList))
	mux.Handle("GET /events/new", adminOnly.ThenFunc(app.eventNewGet))
	mux.Handle("POST /events/new", adminOnly.ThenFunc(app.eventNewPost))
	mux.Handle("GET /events/{id}", protected.ThenFunc(app.eventDetail))
	mux.Handle("POST /events/{id}/attendance", adminOnly.ThenFunc(app.eventUpdateAttendance))
	mux.Handle("POST /events/{id}/publish", adminOnly.ThenFunc(app.eventTogglePublish))

	// Finance / Giving
	mux.Handle("GET /giving/donations", protected.ThenFunc(app.financeList))
	mux.Handle("GET /giving/tithes", protected.ThenFunc(app.financeList))
	mux.Handle("GET /giving/new", adminOnly.ThenFunc(app.financeNewGet))
	mux.Handle("POST /giving/new", adminOnly.ThenFunc(app.financeNewPost))

	// Announcements
	mux.Handle("GET /announcements", protected.ThenFunc(app.announcementsList))
	mux.Handle("GET /announcements/new", adminOnly.ThenFunc(app.announcementNewGet))
	mux.Handle("POST /announcements/new", adminOnly.ThenFunc(app.announcementNewPost))

	// Reports
	mux.Handle("GET /reports", protected.ThenFunc(app.reports))

	// Church Calendar
	mux.Handle("GET /calendar", protected.ThenFunc(app.calendarList))
	mux.Handle("GET /calendar/new", adminOnly.ThenFunc(app.calendarNewGet))
	mux.Handle("POST /calendar/new", adminOnly.ThenFunc(app.calendarNewPost))
	mux.Handle("GET /calendar/{id}", protected.ThenFunc(app.calendarDetail))
	mux.Handle("GET /calendar/{id}/edit", adminOnly.ThenFunc(app.calendarEditGet))
	mux.Handle("POST /calendar/{id}/edit", adminOnly.ThenFunc(app.calendarEditPost))
	mux.Handle("POST /calendar/{id}/delete", adminOnly.ThenFunc(app.calendarDelete))

	// Super-admin: church network
	mux.Handle("GET /invite-church", superAdminOnly.ThenFunc(app.inviteChurch))
	mux.Handle("POST /invite-church", superAdminOnly.ThenFunc(app.inviteChurchPost))
	mux.Handle("GET /admin/churches", superAdminOnly.ThenFunc(app.adminChurches))
	mux.Handle("GET /admin/churches/{id}", superAdminOnly.ThenFunc(app.adminChurchDetail))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}
