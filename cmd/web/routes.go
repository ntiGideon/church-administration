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

	// Custom Roles
	mux.Handle("GET /church/settings/roles",              adminOnly.ThenFunc(app.customRolesList))
	mux.Handle("GET /church/settings/roles/new",          adminOnly.ThenFunc(app.customRoleNewGet))
	mux.Handle("POST /church/settings/roles/new",         adminOnly.ThenFunc(app.customRoleNewPost))
	mux.Handle("GET /church/settings/roles/{id}",         adminOnly.ThenFunc(app.customRoleEditGet))
	mux.Handle("POST /church/settings/roles/{id}",        adminOnly.ThenFunc(app.customRoleEditPost))
	mux.Handle("POST /church/settings/roles/{id}/delete", adminOnly.ThenFunc(app.customRoleDelete))

	// CSV Export & Import  (registered before wildcard /members/{id})
	mux.Handle("GET /members/export",          protected.ThenFunc(app.memberExport))
	mux.Handle("GET /giving/export",           protected.ThenFunc(app.financeExport))
	mux.Handle("GET /members/import",          adminOnly.ThenFunc(app.memberImportGet))
	mux.Handle("POST /members/import",         adminOnly.ThenFunc(app.memberImportPost))
	mux.Handle("GET /members/import/template", adminOnly.ThenFunc(app.memberImportTemplate))

	// Congregation members (Contact records — no system account required)
	mux.Handle("GET /members", protected.ThenFunc(app.membersList))
	mux.Handle("GET /members/new", adminOnly.ThenFunc(app.memberNewGet))
	mux.Handle("POST /members/new", adminOnly.ThenFunc(app.memberNewPost))
	mux.Handle("GET /members/{id}", protected.ThenFunc(app.memberDetail))
	mux.Handle("GET /members/{id}/edit", adminOnly.ThenFunc(app.memberEditGet))
	mux.Handle("POST /members/{id}/edit", adminOnly.ThenFunc(app.memberEditPost))
	mux.Handle("POST /members/{id}/avatar", adminOnly.ThenFunc(app.memberAvatarPost))
	mux.Handle("POST /members/{id}/delete", adminOnly.ThenFunc(app.memberDelete))
	mux.Handle("GET /members/{id}/giving", protected.ThenFunc(app.memberGiving))
	mux.Handle("GET /members/{id}/attendance", protected.ThenFunc(app.memberAttendanceGet))
	mux.Handle("POST /members/{id}/relationships/new", adminOnly.ThenFunc(app.memberRelationshipNewPost))
	mux.Handle("POST /members/{id}/relationships/{rid}/delete", adminOnly.ThenFunc(app.memberRelationshipDelete))
	mux.Handle("POST /members/{id}/pledges/new", adminOnly.ThenFunc(app.memberPledgeNewPost))
	mux.Handle("POST /members/{id}/pledges/{pid}/delete", adminOnly.ThenFunc(app.memberPledgeDelete))

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
	mux.Handle("POST /events/{id}/checkin", adminOnly.ThenFunc(app.eventCheckIn))
	mux.Handle("POST /events/{id}/attendance/{aid}/remove", adminOnly.ThenFunc(app.eventRemoveAttendee))
	mux.Handle("POST /events/{id}/publish", adminOnly.ThenFunc(app.eventTogglePublish))

	// Groups / Cell Groups
	mux.Handle("GET /groups", protected.ThenFunc(app.groupsList))
	mux.Handle("GET /groups/new", adminOnly.ThenFunc(app.groupNewGet))
	mux.Handle("POST /groups/new", adminOnly.ThenFunc(app.groupNewPost))
	mux.Handle("GET /groups/{id}", protected.ThenFunc(app.groupDetail))
	mux.Handle("GET /groups/{id}/edit", adminOnly.ThenFunc(app.groupEditGet))
	mux.Handle("POST /groups/{id}/edit", adminOnly.ThenFunc(app.groupEditPost))
	mux.Handle("POST /groups/{id}/members/add", adminOnly.ThenFunc(app.groupAddMember))
	mux.Handle("POST /groups/{id}/members/{cid}/remove", adminOnly.ThenFunc(app.groupRemoveMember))
	mux.Handle("POST /groups/{id}/delete", adminOnly.ThenFunc(app.groupDelete))

	// Volunteer Rosters
	mux.Handle("GET /rosters", protected.ThenFunc(app.rostersList))
	mux.Handle("GET /rosters/new", adminOnly.ThenFunc(app.rosterNewGet))
	mux.Handle("POST /rosters/new", adminOnly.ThenFunc(app.rosterNewPost))
	mux.Handle("GET /rosters/{id}", protected.ThenFunc(app.rosterDetail))
	mux.Handle("GET /rosters/{id}/edit", adminOnly.ThenFunc(app.rosterEditGet))
	mux.Handle("POST /rosters/{id}/edit", adminOnly.ThenFunc(app.rosterEditPost))
	mux.Handle("POST /rosters/{id}/entries/add", adminOnly.ThenFunc(app.rosterAddEntry))
	mux.Handle("POST /rosters/{id}/entries/{eid}/remove", adminOnly.ThenFunc(app.rosterRemoveEntry))
	mux.Handle("POST /rosters/{id}/delete", adminOnly.ThenFunc(app.rosterDelete))

	// Pastoral Care Notes
	mux.Handle("GET /pastoral", protected.ThenFunc(app.pastoralList))
	mux.Handle("GET /pastoral/new", adminOnly.ThenFunc(app.pastoralNewGet))
	mux.Handle("POST /pastoral/new", adminOnly.ThenFunc(app.pastoralNewPost))
	mux.Handle("GET /pastoral/{id}", protected.ThenFunc(app.pastoralDetail))
	mux.Handle("GET /pastoral/{id}/edit", adminOnly.ThenFunc(app.pastoralEditGet))
	mux.Handle("POST /pastoral/{id}/edit", adminOnly.ThenFunc(app.pastoralEditPost))
	mux.Handle("POST /pastoral/{id}/followup", adminOnly.ThenFunc(app.pastoralMarkFollowUpDone))
	mux.Handle("POST /pastoral/{id}/delete", adminOnly.ThenFunc(app.pastoralDelete))

	// Departments
	mux.Handle("GET /departments", protected.ThenFunc(app.departmentsList))
	mux.Handle("GET /departments/new", adminOnly.ThenFunc(app.departmentNewGet))
	mux.Handle("POST /departments/new", adminOnly.ThenFunc(app.departmentNewPost))
	mux.Handle("GET /departments/{id}", protected.ThenFunc(app.departmentDetail))
	mux.Handle("GET /departments/{id}/edit", adminOnly.ThenFunc(app.departmentEditGet))
	mux.Handle("POST /departments/{id}/edit", adminOnly.ThenFunc(app.departmentEditPost))
	mux.Handle("POST /departments/{id}/members/add", adminOnly.ThenFunc(app.departmentAddMember))
	mux.Handle("POST /departments/{id}/members/{cid}/remove", adminOnly.ThenFunc(app.departmentRemoveMember))
	mux.Handle("POST /departments/{id}/delete", adminOnly.ThenFunc(app.departmentDelete))

	// Member Milestones
	mux.Handle("GET /milestones", protected.ThenFunc(app.milestonesListGet))
	mux.Handle("POST /members/{id}/milestones/new", adminOnly.ThenFunc(app.milestoneNewPost))
	mux.Handle("POST /members/{id}/milestones/{mid}/delete", adminOnly.ThenFunc(app.milestoneDelete))

	// Document Library
	mux.Handle("GET /documents", protected.ThenFunc(app.documentsList))
	mux.Handle("GET /documents/upload", adminOnly.ThenFunc(app.documentUploadGet))
	mux.Handle("POST /documents/upload", adminOnly.ThenFunc(app.documentUploadPost))
	mux.Handle("POST /documents/{id}/delete", adminOnly.ThenFunc(app.documentDelete))

	// Prayer Request Board
	mux.Handle("GET /prayer", protected.ThenFunc(app.prayerList))
	mux.Handle("GET /prayer/new", protected.ThenFunc(app.prayerNewGet))
	mux.Handle("POST /prayer/new", protected.ThenFunc(app.prayerNewPost))
	mux.Handle("GET /prayer/{id}", protected.ThenFunc(app.prayerDetail))
	mux.Handle("GET /prayer/{id}/edit", protected.ThenFunc(app.prayerEditGet))
	mux.Handle("POST /prayer/{id}/edit", protected.ThenFunc(app.prayerEditPost))
	mux.Handle("POST /prayer/{id}/status", protected.ThenFunc(app.prayerUpdateStatus))
	mux.Handle("POST /prayer/{id}/delete", adminOnly.ThenFunc(app.prayerDelete))

	// Visitor Management
	mux.Handle("GET /visitors", protected.ThenFunc(app.visitorsList))
	mux.Handle("GET /visitors/new", protected.ThenFunc(app.visitorNewGet))
	mux.Handle("POST /visitors/new", protected.ThenFunc(app.visitorNewPost))
	mux.Handle("GET /visitors/{id}", protected.ThenFunc(app.visitorDetail))
	mux.Handle("GET /visitors/{id}/edit", protected.ThenFunc(app.visitorEditGet))
	mux.Handle("POST /visitors/{id}/edit", protected.ThenFunc(app.visitorEditPost))
	mux.Handle("POST /visitors/{id}/status", protected.ThenFunc(app.visitorUpdateStatus))
	mux.Handle("POST /visitors/{id}/delete", adminOnly.ThenFunc(app.visitorDelete))

	// Sermon Library
	mux.Handle("GET /sermons", protected.ThenFunc(app.sermonsList))
	mux.Handle("GET /sermons/new", adminOnly.ThenFunc(app.sermonNewGet))
	mux.Handle("POST /sermons/new", adminOnly.ThenFunc(app.sermonNewPost))
	mux.Handle("GET /sermons/{id}", protected.ThenFunc(app.sermonDetail))
	mux.Handle("GET /sermons/{id}/edit", adminOnly.ThenFunc(app.sermonEditGet))
	mux.Handle("POST /sermons/{id}/edit", adminOnly.ThenFunc(app.sermonEditPost))
	mux.Handle("POST /sermons/{id}/publish", adminOnly.ThenFunc(app.sermonTogglePublish))
	mux.Handle("POST /sermons/{id}/delete", adminOnly.ThenFunc(app.sermonDelete))

	// Finance / Giving
	mux.Handle("GET /giving/donations", protected.ThenFunc(app.financeList))
	mux.Handle("GET /giving/tithes", protected.ThenFunc(app.financeList))
	mux.Handle("GET /giving/new", adminOnly.ThenFunc(app.financeNewGet))
	mux.Handle("POST /giving/new", adminOnly.ThenFunc(app.financeNewPost))

	// Announcements
	mux.Handle("GET /announcements", protected.ThenFunc(app.announcementsList))
	mux.Handle("GET /announcements/new", adminOnly.ThenFunc(app.announcementNewGet))
	mux.Handle("POST /announcements/new", adminOnly.ThenFunc(app.announcementNewPost))
	mux.Handle("GET /announcements/{id}", protected.ThenFunc(app.announcementDetail))
	mux.Handle("GET /announcements/{id}/edit", adminOnly.ThenFunc(app.announcementEditGet))
	mux.Handle("POST /announcements/{id}/edit", adminOnly.ThenFunc(app.announcementEditPost))
	mux.Handle("POST /announcements/{id}/publish", adminOnly.ThenFunc(app.announcementTogglePublish))
	mux.Handle("POST /announcements/{id}/delete", adminOnly.ThenFunc(app.announcementDelete))

	// Birthdays
	mux.Handle("GET /birthdays", protected.ThenFunc(app.birthdaysPage))

	// Church Directory
	mux.Handle("GET /directory", protected.ThenFunc(app.directoryGet))

	// Communications (mass email)
	mux.Handle("GET /communications",          adminOnly.ThenFunc(app.communicationsList))
	mux.Handle("GET /communications/compose",  adminOnly.ThenFunc(app.communicationComposeGet))
	mux.Handle("POST /communications/compose", adminOnly.ThenFunc(app.communicationComposePost))
	mux.Handle("GET /communications/{id}",     adminOnly.ThenFunc(app.communicationDetail))

	// Budget planning
	mux.Handle("GET /budgets",                          protected.ThenFunc(app.budgetsList))
	mux.Handle("GET /budgets/new",                      adminOnly.ThenFunc(app.budgetNewGet))
	mux.Handle("POST /budgets/new",                     adminOnly.ThenFunc(app.budgetNewPost))
	mux.Handle("GET /budgets/{id}",                     protected.ThenFunc(app.budgetDetail))
	mux.Handle("GET /budgets/{id}/edit",                adminOnly.ThenFunc(app.budgetEditGet))
	mux.Handle("POST /budgets/{id}/edit",               adminOnly.ThenFunc(app.budgetEditPost))
	mux.Handle("POST /budgets/{id}/lines/add",          adminOnly.ThenFunc(app.budgetLineAdd))
	mux.Handle("POST /budgets/{id}/lines/{lid}/delete", adminOnly.ThenFunc(app.budgetLineDelete))
	mux.Handle("POST /budgets/{id}/delete",             adminOnly.ThenFunc(app.budgetDelete))

	// Reports
	mux.Handle("GET /reports", protected.ThenFunc(app.reports))
	mux.Handle("GET /reports/finance", protected.ThenFunc(app.financeReports))
	mux.Handle("GET /reports/attendance", protected.ThenFunc(app.attendanceReports))

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

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders, app.errorPages)
	return standard.Then(mux)
}
