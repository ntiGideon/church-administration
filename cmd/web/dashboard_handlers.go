package main

import (
	"net/http"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/user"
)

// GET /dashboard
func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	stats := map[string]interface{}{
		"totalIncome":       float64(0),
		"totalExpenses":     float64(0),
		"netBalance":        float64(0),
		"thisMonthIncome":   float64(0),
		"thisMonthExpenses": float64(0),
	}

	isSuperAdmin := string(u.Role) == "super_admin"

	if isSuperAdmin {
		// Super admin: show network-wide overview
		churches, _ := app.db.Church.Query().
			WithUsers(func(uq *ent.UserQuery) {
				uq.Where(user.IsActiveEQ(true))
			}).
			All(r.Context())

		totalMembers := 0
		for _, c := range churches {
			totalMembers += len(c.Edges.Users)
		}

		stats["churches"] = churches
		stats["churchCount"] = len(churches)
		stats["memberCount"] = totalMembers

		// Global finance totals (all churches)
		summary, _ := app.financeModel.Summary(r.Context(), 0)
		if summary != nil {
			stats["totalIncome"] = summary.TotalIncome
			stats["totalExpenses"] = summary.TotalExpenses
			stats["netBalance"] = summary.NetBalance
			stats["thisMonthIncome"] = summary.ThisMonthIncome
			stats["thisMonthExpenses"] = summary.ThisMonthExpenses
		}

	} else {
		// Branch admin / other roles: show church-scoped data
		churchID := 0
		if u.Edges.Church != nil {
			churchID = u.Edges.Church.ID
		}

		if churchID > 0 {
			memberCount, _ := app.memberModel.CountByChurch(r.Context(), churchID)
			stats["memberCount"] = memberCount

			upcomingEvents, _ := app.eventModel.Upcoming(r.Context(), churchID, 5)
			stats["upcomingEvents"] = upcomingEvents

			summary, _ := app.financeModel.Summary(r.Context(), churchID)
			if summary != nil {
				stats["totalIncome"] = summary.TotalIncome
				stats["totalExpenses"] = summary.TotalExpenses
				stats["netBalance"] = summary.NetBalance
				stats["thisMonthIncome"] = summary.ThisMonthIncome
				stats["thisMonthExpenses"] = summary.ThisMonthExpenses
			}

			recentTx, _ := app.financeModel.RecentTransactions(r.Context(), churchID, 5)
			stats["recentTransactions"] = recentTx

			// Congregation size from church record
			if u.Edges.Church != nil {
				stats["congregationSize"] = u.Edges.Church.CongregationSize
			}
		}
	}

	data.Data = stats
	app.render(w, r, http.StatusOK, "dashboard.gohtml", data)
}
