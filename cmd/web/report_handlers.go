package main

import (
	"net/http"

	"github.com/ntiGideon/ent/user"
)

// GET /reports
func (app *application) reports(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	churchID := 0
	if u.Role != user.RoleSuperAdmin && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	stats := map[string]interface{}{
		"totalIncome":       float64(0),
		"totalExpenses":     float64(0),
		"netBalance":        float64(0),
		"thisMonthIncome":   float64(0),
		"thisMonthExpenses": float64(0),
	}

	// Member stats
	memberCount, _ := app.memberModel.CountByChurch(r.Context(), churchID)
	stats["memberCount"] = memberCount

	members, _ := app.memberModel.ListByChurch(r.Context(), churchID)

	// Role breakdown
	roleCounts := map[string]int{}
	for _, m := range members {
		roleCounts[string(m.Role)]++
	}
	stats["roleCounts"] = roleCounts
	stats["members"] = members

	// Finance summary
	summary, _ := app.financeModel.Summary(r.Context(), churchID)
	if summary != nil {
		stats["totalIncome"] = summary.TotalIncome
		stats["totalExpenses"] = summary.TotalExpenses
		stats["netBalance"] = summary.NetBalance
		stats["thisMonthIncome"] = summary.ThisMonthIncome
		stats["thisMonthExpenses"] = summary.ThisMonthExpenses
	}

	// Recent transactions for table
	recentTx, _ := app.financeModel.RecentTransactions(r.Context(), churchID, 20)
	stats["recentTransactions"] = recentTx

	// Event stats
	events, _ := app.eventModel.List(r.Context(), churchID)
	stats["eventCount"] = len(events)

	totalAttendance := 0
	for _, e := range events {
		totalAttendance += e.AttendanceCount
	}
	stats["totalAttendance"] = totalAttendance

	// Church count for super admin
	if u.Role == user.RoleSuperAdmin {
		churchCount, _ := app.db.Church.Query().Count(r.Context())
		stats["churchCount"] = churchCount
	}

	data := app.newTemplateData(r)
	data.Data = stats
	app.render(w, r, http.StatusOK, "reports.gohtml", data)
}
