package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// CalendarDay represents a single cell in the calendar grid.
type CalendarDay struct {
	Day     int
	IsToday bool
	Entries []*ent.ProgramEntry
}

// buildCalendarDays returns a flat slice of CalendarDay cells (padded to fill 7-column rows).
func buildCalendarDays(year, month int, entries []*ent.ProgramEntry) []CalendarDay {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	firstWeekday := int(firstDay.Weekday()) // 0=Sunday

	now := time.Now()
	todayDay := 0
	if now.Year() == year && int(now.Month()) == month {
		todayDay = now.Day()
	}

	byDay := map[int][]*ent.ProgramEntry{}
	for _, e := range entries {
		d := e.Date.Day()
		byDay[d] = append(byDay[d], e)
	}

	days := make([]CalendarDay, 0, 42)
	for i := 0; i < firstWeekday; i++ {
		days = append(days, CalendarDay{})
	}
	for d := 1; d <= daysInMonth; d++ {
		days = append(days, CalendarDay{
			Day:     d,
			IsToday: d == todayDay,
			Entries: byDay[d],
		})
	}
	for len(days)%7 != 0 {
		days = append(days, CalendarDay{})
	}
	return days
}

// GET /calendar
func (app *application) calendarList(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	now := time.Now()
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = int(now.Month())
	}

	// Clamp month
	if month < 1 {
		month = 1
	}
	if month > 12 {
		month = 12
	}

	entries, err := app.programModel.ListByMonthYear(r.Context(), churchID, year, month)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Previous / next month navigation
	prev := time.Date(year, time.Month(month)-1, 1, 0, 0, 0, 0, time.UTC)
	next := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC)

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"entries":      entries,
		"calDays":      buildCalendarDays(year, month, entries),
		"year":         year,
		"month":        month,
		"monthName":    time.Month(month).String(),
		"prevYear":     prev.Year(),
		"prevMonth":    int(prev.Month()),
		"nextYear":     next.Year(),
		"nextMonth":    int(next.Month()),
		"isCurrentMonth": year == now.Year() && month == int(now.Month()),
	}
	app.render(w, r, http.StatusOK, "calendar.gohtml", data)
}

// GET /calendar/new
func (app *application) calendarNewGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	// Pre-fill date from query param if provided
	prefillDate := r.URL.Query().Get("date")
	data.Form = models.ProgramDto{Date: prefillDate}
	app.render(w, r, http.StatusOK, "calendar_new.gohtml", data)
}

// POST /calendar/new
func (app *application) calendarNewPost(w http.ResponseWriter, r *http.Request) {
	var dto models.ProgramDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.ProgramType), "program_type", "Program type is required")
	dto.CheckField(validator.NotBlank(dto.Date), "date", "Date is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "calendar_new.gohtml", data)
		return
	}

	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	entry, err := app.programModel.Create(r.Context(), &dto, churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Program entry added to calendar!")
	http.Redirect(w, r, "/calendar/"+strconv.Itoa(entry.ID), http.StatusSeeOther)
}

// GET /calendar/{id}
func (app *application) calendarDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	entry, err := app.programModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{"entry": entry}
	app.render(w, r, http.StatusOK, "calendar_detail.gohtml", data)
}

// GET /calendar/{id}/edit
func (app *application) calendarEditGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	entry, err := app.programModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	// Pre-fill form from existing entry
	dto := models.ProgramDto{
		Title:             entry.Title,
		ProgramType:       string(entry.ProgramType),
		Date:              entry.Date.Format("2006-01-02"),
		Theme:             entry.Theme,
		SermonTopic:       entry.SermonTopic,
		VisionGoals:       entry.VisionGoals,
		Preacher:          entry.Preacher,
		OpeningPrayerBy:   entry.OpeningPrayerBy,
		ClosingPrayerBy:   entry.ClosingPrayerBy,
		WorshipLeader:     entry.WorshipLeader,
		ResponsiblePerson: entry.ResponsiblePerson,
		Notes:             entry.Notes,
		IsPublished:       entry.IsPublished,
	}

	data := app.newTemplateData(r)
	data.Form = dto
	data.Data = map[string]interface{}{"entry": entry}
	app.render(w, r, http.StatusOK, "calendar_edit.gohtml", data)
}

// POST /calendar/{id}/edit
func (app *application) calendarEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.ProgramDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Title), "title", "Title is required")
	dto.CheckField(validator.NotBlank(dto.ProgramType), "program_type", "Program type is required")
	dto.CheckField(validator.NotBlank(dto.Date), "date", "Date is required")

	if !dto.Valid() {
		entry, _ := app.programModel.GetByID(r.Context(), id)
		data := app.newTemplateData(r)
		data.Form = dto
		data.Data = map[string]interface{}{"entry": entry}
		app.render(w, r, http.StatusUnprocessableEntity, "calendar_edit.gohtml", data)
		return
	}

	if _, err := app.programModel.Update(r.Context(), id, &dto); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Program entry updated!")
	http.Redirect(w, r, "/calendar/"+strconv.Itoa(id), http.StatusSeeOther)
}

// POST /calendar/{id}/delete
func (app *application) calendarDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	entry, err := app.programModel.GetByID(r.Context(), id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	// Remember year/month for redirect back to that calendar page
	year := entry.Date.Year()
	month := int(entry.Date.Month())

	if err := app.programModel.Delete(r.Context(), id); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Program entry removed.")
	http.Redirect(w, r, "/calendar?year="+strconv.Itoa(year)+"&month="+strconv.Itoa(month), http.StatusSeeOther)
}

// derefStrCal dereferences an optional *string field, returning "" for nil.
func derefStrCal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
