package main

import (
	"net/http"

	"github.com/ntiGideon/ent/attendance"
	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/event"
)

// GET /reports/attendance
func (app *application) attendanceReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid := app.churchID(r)

	// Monthly attendance trend (12 months)
	atTrend, _ := app.attendanceModel.MonthlyAttendanceTrend(ctx, cid, 12)

	// Event type breakdown by attendance_count
	evBreakdown, _ := app.eventModel.EventTypeBreakdown(ctx, cid)

	// Member growth trend (12 months)
	growthTrend, _ := app.memberModel.MemberGrowthTrend(ctx, cid, 12)

	// Visitor conversion funnel
	visitorStatuses, _ := app.visitorModel.CountByStatus(ctx, cid)

	// Total check-ins for this church
	attQ := app.db.Attendance.Query().
		Where(attendance.HasEventWith(event.HasChurchWith(church.IDEQ(cid))))
	totalAttendances, _ := attQ.Count(ctx)

	// Total events for this church
	evQ := app.db.Event.Query()
	if cid > 0 {
		evQ = evQ.Where(event.HasChurchWith(church.IDEQ(cid)))
	}
	totalEvents, _ := evQ.Count(ctx)

	// Total members (contacts)
	totalMembers, _ := app.memberModel.CountContactsByChurch(ctx, cid)

	// This month check-ins (last slot of atTrend)
	thisMonth := 0
	if len(atTrend) > 0 {
		thisMonth = atTrend[len(atTrend)-1].Total
	}

	// Average check-ins per event
	avgPerEvent := 0
	if totalEvents > 0 {
		avgPerEvent = totalAttendances / totalEvents
	}

	// Build Chart.js-friendly slices from trend data
	months := make([]string, len(atTrend))
	presentData := make([]int, len(atTrend))
	lateData := make([]int, len(atTrend))
	for i, t := range atTrend {
		months[i] = t.Month
		presentData[i] = t.Present
		lateData[i] = t.Late
	}

	evLabels := make([]string, len(evBreakdown))
	evCounts := make([]int, len(evBreakdown))
	for i, e := range evBreakdown {
		evLabels[i] = e.Label
		evCounts[i] = e.Count
	}

	growthMonths := make([]string, len(growthTrend))
	growthCounts := make([]int, len(growthTrend))
	for i, g := range growthTrend {
		growthMonths[i] = g.Month
		growthCounts[i] = g.Count
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"months":          months,
		"presentData":     presentData,
		"lateData":        lateData,
		"evLabels":        evLabels,
		"evCounts":        evCounts,
		"growthMonths":    growthMonths,
		"growthCounts":    growthCounts,
		"visitorStatuses": visitorStatuses,
		"totalAttendances": totalAttendances,
		"totalEvents":     totalEvents,
		"totalMembers":    totalMembers,
		"thisMonth":       thisMonth,
		"avgPerEvent":     avgPerEvent,
	}
	app.render(w, r, http.StatusOK, "attendance_reports.gohtml", data)
}
