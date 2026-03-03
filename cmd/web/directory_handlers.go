package main

import (
	"net/http"
	"sort"
	"strings"

	"github.com/ntiGideon/ent"
)

// DirectoryEntry pairs a Contact with the group/department names they belong to.
type DirectoryEntry struct {
	Contact *ent.Contact
	Groups  []string
	Depts   []string
}

// AlphaSection groups directory entries under a single letter heading.
type AlphaSection struct {
	Letter  string
	Entries []DirectoryEntry
}

// GET /directory
func (app *application) directoryGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cid := app.churchID(r)

	contacts, _ := app.memberModel.ListContactsByChurch(ctx, cid)
	groups, _ := app.groupModel.ListByChurch(ctx, cid)
	depts, _ := app.departmentModel.ListByChurch(ctx, cid)

	// Build contactID → group name slices from loaded member edges
	contactGroups := make(map[int][]string)
	for _, g := range groups {
		for _, m := range g.Edges.Members {
			contactGroups[m.ID] = append(contactGroups[m.ID], g.Name)
		}
	}

	// Build contactID → dept name slices from loaded member edges
	contactDepts := make(map[int][]string)
	for _, d := range depts {
		for _, m := range d.Edges.Members {
			contactDepts[m.ID] = append(contactDepts[m.ID], d.Name)
		}
	}

	// Compute summary stats
	withPhoto := 0
	withContact := 0
	for _, c := range contacts {
		if c.ProfilePictureURL != "" {
			withPhoto++
		}
		if c.Phone != "" || c.Email != "" {
			withContact++
		}
	}

	// Build entries and bucket them alphabetically by first name
	type slotKey = string
	sectMap := map[slotKey][]DirectoryEntry{}
	var letters []string

	for _, c := range contacts {
		entry := DirectoryEntry{
			Contact: c,
			Groups:  contactGroups[c.ID],
			Depts:   contactDepts[c.ID],
		}
		letter := "#"
		if len(c.FirstName) > 0 {
			first := []rune(strings.TrimSpace(c.FirstName))
			if len(first) > 0 {
				letter = strings.ToUpper(string(first[0]))
			}
		}
		if _, exists := sectMap[letter]; !exists {
			letters = append(letters, letter)
		}
		sectMap[letter] = append(sectMap[letter], entry)
	}

	sort.Strings(letters)

	sections := make([]AlphaSection, len(letters))
	for i, l := range letters {
		sections[i] = AlphaSection{Letter: l, Entries: sectMap[l]}
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"sections":    sections,
		"groups":      groups,
		"depts":       depts,
		"total":       len(contacts),
		"withPhoto":   withPhoto,
		"withContact": withContact,
		"letters":     letters,
	}
	app.render(w, r, http.StatusOK, "directory.gohtml", data)
}
