package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// ─── view models ─────────────────────────────────────────────────────────────

// RelativeView is a flattened representation of one relationship entry,
// resolved from the perspective of a given member (the "self" contact).
type RelativeView struct {
	RelationshipID int
	RelativeID     int
	FirstName      string
	LastName       string
	Initials       string
	PhotoURL       string
	RelationType   string // raw enum value, e.g. "wife"
	RelationLabel  string // display label, e.g. "Wife"
	Notes          string
}

// FamilyGroup groups relatives by relationship category (Spouse, Parents, …).
type FamilyGroup struct {
	Key     string
	Label   string
	Icon    string
	Color   string
	Bg      string
	Members []RelativeView
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var relationLabelMap = map[string]string{
	"wife": "Wife", "husband": "Husband",
	"mother": "Mother", "father": "Father",
	"daughter": "Daughter", "son": "Son",
	"sister": "Sister", "brother": "Brother",
	"grandmother": "Grandmother", "grandfather": "Grandfather",
	"granddaughter": "Granddaughter", "grandson": "Grandson",
	"aunt": "Aunt", "uncle": "Uncle",
	"niece": "Niece", "nephew": "Nephew",
	"cousin":      "Cousin",
	"friend":      "Friend",
	"godmother":   "Godmother", "godfather": "Godfather",
	"goddaughter": "Goddaughter", "godson": "Godson",
	"other": "Other",
}

var relationGroupMap = map[string]string{
	"wife": "spouse", "husband": "spouse",
	"mother": "parents", "father": "parents",
	"daughter": "children", "son": "children",
	"sister": "siblings", "brother": "siblings",
	"grandmother": "grandparents", "grandfather": "grandparents",
	"granddaughter": "grandchildren", "grandson": "grandchildren",
	"aunt": "extended", "uncle": "extended",
	"niece": "extended", "nephew": "extended", "cousin": "extended",
	"godmother": "spiritual", "godfather": "spiritual",
	"goddaughter": "spiritual", "godson": "spiritual",
	"friend": "friends",
	"other":  "other",
}

// buildFamilyTree groups a flat list of relationship records into FamilyGroups,
// always resolving "the other person" relative to memberID.
func buildFamilyTree(memberID int, rels []*ent.Relationship) []FamilyGroup {
	groups := []FamilyGroup{
		{Key: "spouse", Label: "Spouse", Icon: "fa-heart", Color: "#EC4899", Bg: "rgba(236,72,153,0.08)"},
		{Key: "parents", Label: "Parents", Icon: "fa-house-user", Color: "#3B82F6", Bg: "rgba(59,130,246,0.08)"},
		{Key: "children", Label: "Children", Icon: "fa-child", Color: "#10B981", Bg: "rgba(16,185,129,0.08)"},
		{Key: "siblings", Label: "Siblings", Icon: "fa-people-group", Color: "#8B5CF6", Bg: "rgba(139,92,246,0.08)"},
		{Key: "grandparents", Label: "Grandparents", Icon: "fa-person-cane", Color: "#F59E0B", Bg: "rgba(245,158,11,0.08)"},
		{Key: "grandchildren", Label: "Grandchildren", Icon: "fa-baby", Color: "#06B6D4", Bg: "rgba(6,182,212,0.08)"},
		{Key: "extended", Label: "Extended Family", Icon: "fa-people-roof", Color: "#F97316", Bg: "rgba(249,115,22,0.08)"},
		{Key: "spiritual", Label: "Spiritual Family", Icon: "fa-cross", Color: "#6366F1", Bg: "rgba(99,102,241,0.08)"},
		{Key: "friends", Label: "Friends", Icon: "fa-user-group", Color: "#0EA5E9", Bg: "rgba(14,165,233,0.08)"},
		{Key: "other", Label: "Other", Icon: "fa-link", Color: "#9CA3AF", Bg: "rgba(156,163,175,0.08)"},
	}

	groupIdx := make(map[string]int, len(groups))
	for i, g := range groups {
		groupIdx[g.Key] = i
	}

	for _, rel := range rels {
		relType := string(rel.RelationType)

		var relative *ent.Contact
		if rel.FromContactID == memberID {
			relative = rel.Edges.ToContact
		} else {
			relative = rel.Edges.FromContact
		}
		if relative == nil {
			continue
		}

		rv := RelativeView{
			RelationshipID: rel.ID,
			RelativeID:     relative.ID,
			FirstName:      relative.FirstName,
			LastName:       relative.LastName,
			PhotoURL:       relative.ProfilePictureURL,
			RelationType:   relType,
			RelationLabel:  relationLabelMap[relType],
			Notes:          rel.Notes,
		}
		if len(rv.FirstName) > 0 && len(rv.LastName) > 0 {
			rv.Initials = string([]rune(rv.FirstName)[0]) + string([]rune(rv.LastName)[0])
		}

		key := relationGroupMap[relType]
		if idx, ok := groupIdx[key]; ok {
			groups[idx].Members = append(groups[idx].Members, rv)
		}
	}

	return groups
}

// ─── handlers ─────────────────────────────────────────────────────────────────

// POST /members/{id}/relationships/new
func (app *application) memberRelationshipNewPost(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	var dto models.RelationshipDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(dto.RelativeID > 0, "relative_id", "Please select a relative")
	dto.CheckField(validator.NotBlank(dto.RelationType), "relation_type", "Please select a relationship type")
	dto.CheckField(dto.RelativeID != memberID, "relative_id", "A member cannot be their own relative")

	redirectURL := fmt.Sprintf("/members/%d", memberID)

	if !dto.Valid() {
		// Flash the first field error and redirect back
		for _, msg := range dto.FieldErrors {
			app.sessionManager.Put(r.Context(), "flash_error", msg)
			break
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	_, err = app.relationshipModel.Create(r.Context(), memberID, dto.RelativeID, dto.RelationType, dto.Notes)
	if err != nil {
		if err == models.ErrDuplicateRelationship {
			app.sessionManager.Put(r.Context(), "flash_error", "A relationship with that member already exists.")
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Relationship added successfully."))
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// POST /members/{id}/relationships/{rid}/delete
func (app *application) memberRelationshipDelete(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || memberID < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	rid, err := strconv.Atoi(r.PathValue("rid"))
	if err != nil || rid < 1 {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.relationshipModel.Delete(r.Context(), rid); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Relationship removed.")
	http.Redirect(w, r, fmt.Sprintf("/members/%d", memberID), http.StatusSeeOther)
}
