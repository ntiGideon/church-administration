package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ntiGideon/internal/models"
)

// ImportResult holds the outcome of a single CSV data row during bulk import.
type ImportResult struct {
	Row    int
	Name   string
	Status string // "imported" | "skipped" | "error"
	Reason string
}

// csvHeaderMap maps common column header variants (lowercased) to internal keys.
var csvHeaderMap = map[string]string{
	"first name":          "first_name",
	"firstname":           "first_name",
	"first_name":          "first_name",
	"last name":           "last_name",
	"lastname":            "last_name",
	"last_name":           "last_name",
	"middle name":         "middle_name",
	"middlename":          "middle_name",
	"middle_name":         "middle_name",
	"email":               "email",
	"phone":               "phone",
	"gender":              "gender",
	"date of birth":       "dob",
	"date_of_birth":       "dob",
	"dob":                 "dob",
	"birthday":            "dob",
	"birth date":          "dob",
	"marital status":      "marital_status",
	"marital_status":      "marital_status",
	"day born":            "day_born",
	"day_born":            "day_born",
	"occupation":          "occupation",
	"address":             "address",
	"address_line1":       "address",
	"address line1":       "address",
	"city":                "city",
	"country":             "country",
	"region":              "region",
	"hometown":            "hometown",
	"id number":           "id_number",
	"id_number":           "id_number",
	"sunday school class": "sunday_school_class",
	"sunday_school_class": "sunday_school_class",
	"has spouse":          "has_spouse",
	"has_spouse":          "has_spouse",
	"is baptized":         "is_baptized",
	"is_baptized":         "is_baptized",
	"baptized by":         "baptized_by",
	"baptized_by":         "baptized_by",
	"baptism church":      "baptism_church",
	"baptism_church":      "baptism_church",
	"baptism cert no.":    "baptism_cert",
	"baptism cert no":     "baptism_cert",
	"baptism_cert_number": "baptism_cert",
	"baptism cert number": "baptism_cert",
	"baptism date":        "baptism_date",
	"baptism_date":        "baptism_date",
	"membership year":     "membership_year",
	"membership_year":     "membership_year",
}

// GET /members/import
func (app *application) memberImportGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "member_import.gohtml", data)
}

// GET /members/import/template — download a blank CSV template
func (app *application) memberImportTemplate(w http.ResponseWriter, r *http.Request) {
	cw := csvWriter(w, "members-import-template.csv")
	_ = cw.Write([]string{
		"First Name", "Last Name", "Middle Name", "Email", "Phone",
		"Gender", "Date of Birth", "Marital Status", "Day Born",
		"Occupation", "Address", "City", "Country", "Region", "Hometown",
		"ID Number", "Sunday School Class",
		"Has Spouse", "Is Baptized", "Baptized By", "Baptism Church",
		"Baptism Cert No.", "Baptism Date", "Membership Year",
	})
	// One sample row so users can see the expected format
	_ = cw.Write([]string{
		"Kofi", "Mensah", "", "kofi@example.com", "0241234567",
		"male", "1985-06-15", "married", "Monday",
		"Teacher", "123 High Street", "Accra", "Ghana", "Greater Accra", "Kumasi",
		"GH-12345", "Class A",
		"Yes", "Yes", "Rev. Asante", "First Baptist", "BC-001", "2010-04-20", "2015",
	})
	cw.Flush()
}

// POST /members/import — parse the uploaded CSV and create contact records
func (app *application) memberImportPost(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}
	if churchID == 0 {
		app.clientError(w, http.StatusForbidden)
		return
	}

	const maxCSVSize = 5 << 20 // 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxCSVSize)
	if err := r.ParseMultipartForm(maxCSVSize); err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "File too large — maximum 5 MB.")
		http.Redirect(w, r, "/members/import", http.StatusSeeOther)
		return
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "No file selected.")
		http.Redirect(w, r, "/members/import", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Strip UTF-8 BOM if present (common when file was saved by Excel)
	buf := make([]byte, 3)
	n, _ := file.Read(buf)
	var reader *csv.Reader
	if n == 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		reader = csv.NewReader(file) // BOM consumed, rest of file follows
	} else {
		reader = csv.NewReader(io.MultiReader(strings.NewReader(string(buf[:n])), file))
	}
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "Could not parse the CSV — check the file format.")
		http.Redirect(w, r, "/members/import", http.StatusSeeOther)
		return
	}
	if len(rows) < 2 {
		app.sessionManager.Put(r.Context(), "flash_error", "The file has no data rows.")
		http.Redirect(w, r, "/members/import", http.StatusSeeOther)
		return
	}

	// Build column-index map from the header row
	colIdx := make(map[string]int)
	for i, h := range rows[0] {
		norm := strings.ToLower(strings.TrimSpace(h))
		if key, ok := csvHeaderMap[norm]; ok {
			colIdx[key] = i
		}
	}

	get := func(row []string, key string) string {
		if i, ok := colIdx[key]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	parseBool := func(s string) bool {
		s = strings.ToLower(s)
		return s == "yes" || s == "true" || s == "1"
	}

	var results []ImportResult
	imported, skipped, errored := 0, 0, 0

	for i, row := range rows[1:] {
		lineNum := i + 2 // 1-based, accounting for header

		firstName := get(row, "first_name")
		lastName  := get(row, "last_name")
		name      := strings.TrimSpace(firstName + " " + lastName)
		if name == "" {
			name = fmt.Sprintf("Row %d", lineNum)
		}

		if firstName == "" || lastName == "" {
			results = append(results, ImportResult{lineNum, name, "skipped", "First Name and Last Name are required"})
			skipped++
			continue
		}

		dto := models.MemberDto{
			FirstName:         firstName,
			LastName:          lastName,
			MiddleName:        get(row, "middle_name"),
			Email:             get(row, "email"),
			Phone:             get(row, "phone"),
			Gender:            get(row, "gender"),
			DateOfBirth:       get(row, "dob"),
			MaritalStatus:     get(row, "marital_status"),
			DayBorn:           get(row, "day_born"),
			Occupation:        get(row, "occupation"),
			AddressLine1:      get(row, "address"),
			City:              get(row, "city"),
			Country:           get(row, "country"),
			Region:            get(row, "region"),
			Hometown:          get(row, "hometown"),
			IDNumber:          get(row, "id_number"),
			SundaySchoolClass: get(row, "sunday_school_class"),
			HasSpouse:         parseBool(get(row, "has_spouse")),
			IsBaptized:        parseBool(get(row, "is_baptized")),
			BaptizedBy:        get(row, "baptized_by"),
			BaptismChurch:     get(row, "baptism_church"),
			BaptismCertNumber: get(row, "baptism_cert"),
			BaptismDate:       get(row, "baptism_date"),
		}
		if my := get(row, "membership_year"); my != "" {
			if yr, err2 := strconv.Atoi(my); err2 == nil {
				dto.MembershipYear = yr
			}
		}

		if _, err2 := app.memberModel.CreateContact(r.Context(), &dto, churchID); err2 != nil {
			results = append(results, ImportResult{lineNum, name, "error", "Could not save — check for duplicate or invalid data"})
			errored++
		} else {
			results = append(results, ImportResult{lineNum, name, "imported", ""})
			imported++
		}
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"results":  results,
		"imported": imported,
		"skipped":  skipped,
		"errored":  errored,
		"total":    len(rows) - 1,
	}
	app.render(w, r, http.StatusOK, "member_import.gohtml", data)
}
