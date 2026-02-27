package main

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ntiGideon/ent"
)

// BirthdayEntry holds display data for one upcoming birthday.
type BirthdayEntry struct {
	ContactID   int
	FirstName   string
	LastName    string
	PhotoURL    string
	Initials    string
	DaysUntil   int
	AgeTurning  int
	BirthdayStr string // e.g. "Jan 2"
	IsToday     bool
}

// upcomingBirthdays returns contacts whose next birthday falls within [0, days]
// calendar days from today, sorted by days-until ascending.
func upcomingBirthdays(contacts []*ent.Contact, days int) []BirthdayEntry {
	today := time.Now()
	var entries []BirthdayEntry

	for _, c := range contacts {
		if c.DateOfBirth.IsZero() {
			continue
		}
		du := daysUntilBirthday(c.DateOfBirth, today)
		if du > days {
			continue
		}
		initials := ""
		if len(c.FirstName) > 0 {
			initials += string([]rune(c.FirstName)[:1])
		}
		if len(c.LastName) > 0 {
			initials += string([]rune(c.LastName)[:1])
		}
		entries = append(entries, BirthdayEntry{
			ContactID:   c.ID,
			FirstName:   c.FirstName,
			LastName:    c.LastName,
			PhotoURL:    c.ProfilePictureURL,
			Initials:    strings.ToUpper(initials),
			DaysUntil:   du,
			AgeTurning:  ageTurning(c.DateOfBirth, today),
			BirthdayStr: c.DateOfBirth.Format("Jan 2"),
			IsToday:     du == 0,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DaysUntil < entries[j].DaysUntil
	})
	return entries
}

// daysUntilBirthday returns 0 if today is the birthday, or the number of days
// until the next occurrence of that month/day.
func daysUntilBirthday(dob, today time.Time) int {
	loc := today.Location()
	todayMid := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	thisYear := time.Date(today.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, loc)
	diff := int(thisYear.Sub(todayMid).Hours() / 24)
	if diff < 0 {
		nextYear := time.Date(today.Year()+1, dob.Month(), dob.Day(), 0, 0, 0, 0, loc)
		diff = int(nextYear.Sub(todayMid).Hours() / 24)
	}
	return diff
}

// ageTurning returns the age the contact will turn on their next birthday.
func ageTurning(dob, today time.Time) int {
	loc := today.Location()
	todayMid := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	thisYear := time.Date(today.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, loc)
	age := today.Year() - dob.Year()
	if thisYear.Before(todayMid) {
		// Birthday already passed this year; they'll turn age+1 on the next occurrence
		age++
	}
	return age
}

// GET /birthdays
func (app *application) birthdaysPage(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	contacts, err := app.memberModel.ListContactsByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	window := 30
	switch r.URL.Query().Get("window") {
	case "7":
		window = 7
	case "60":
		window = 60
	case "365":
		window = 365
	}

	entries := upcomingBirthdays(contacts, window)

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"entries": entries,
		"window":  window,
	}
	app.render(w, r, http.StatusOK, "birthdays.gohtml", data)
}
