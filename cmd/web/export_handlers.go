package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"
)

// csvWriter sets the response headers for a CSV file download and returns an
// *encoding/csv.Writer ready to receive rows.
func csvWriter(w http.ResponseWriter, filename string) *csv.Writer {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-cache")
	cw := csv.NewWriter(w)
	// BOM so Excel opens UTF-8 correctly on Windows
	fmt.Fprint(w, "\xEF\xBB\xBF")
	return cw
}

// GET /members/export
func (app *application) memberExport(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	members, err := app.memberModel.ListContactsByChurch(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	filename := fmt.Sprintf("members-%s.csv", time.Now().Format("2006-01-02"))
	cw := csvWriter(w, filename)

	// Header row
	_ = cw.Write([]string{
		"ID", "First Name", "Last Name", "Middle Name",
		"Email", "Phone", "Gender", "Date of Birth", "Age",
		"Marital Status", "Day Born",
		"Occupation", "Address", "City", "Country", "Region", "Hometown",
		"ID Number", "Sunday School Class",
		"Has Spouse", "Is Baptized", "Baptized By", "Baptism Church",
		"Baptism Cert No.", "Baptism Date", "Membership Year",
	})

	today := time.Now()
	for _, m := range members {
		dob := ""
		age := ""
		if !m.DateOfBirth.IsZero() {
			dob = m.DateOfBirth.Format("2006-01-02")
			age = fmt.Sprintf("%d", today.Year()-m.DateOfBirth.Year())
		}
		baptismDate := ""
		if !m.BaptismDate.IsZero() {
			baptismDate = m.BaptismDate.Format("2006-01-02")
		}
		membershipYear := ""
		if m.MembershipYear > 0 {
			membershipYear = fmt.Sprintf("%d", m.MembershipYear)
		}
		_ = cw.Write([]string{
			fmt.Sprintf("%d", m.ID),
			m.FirstName,
			m.LastName,
			m.MiddleName,
			m.Email,
			m.Phone,
			string(m.Gender),
			dob,
			age,
			string(m.MaritalStatus),
			string(m.DayBorn),
			m.Occupation,
			m.AddressLine1,
			m.City,
			m.Country,
			m.Region,
			m.Hometown,
			m.IDNumber,
			m.SundaySchoolClass,
			boolStr(m.HasSpouse),
			boolStr(m.IsBaptized),
			m.BaptizedBy,
			m.BaptismChurch,
			m.BaptismCertNumber,
			baptismDate,
			membershipYear,
		})
	}

	cw.Flush()
}

// GET /giving/export
func (app *application) financeExport(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	churchID := 0
	if u != nil && u.Edges.Church != nil {
		churchID = u.Edges.Church.ID
	}

	records, err := app.financeModel.List(r.Context(), churchID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	filename := fmt.Sprintf("finance-%s.csv", time.Now().Format("2006-01-02"))
	cw := csvWriter(w, filename)

	_ = cw.Write([]string{
		"Date", "Description", "Type", "Category",
		"Amount", "Currency", "Donor", "Payment Method", "Notes", "Recorded By",
	})

	for _, rec := range records {
		donor := ""
		if rec.Edges.Donor != nil {
			donor = rec.Edges.Donor.FirstName + " " + rec.Edges.Donor.LastName
		}
		recordedBy := ""
		if rec.Edges.RecordedBy != nil && rec.Edges.RecordedBy.Edges.Contact != nil {
			c := rec.Edges.RecordedBy.Edges.Contact
			recordedBy = c.FirstName + " " + c.LastName
		}
		_ = cw.Write([]string{
			rec.TransactionDate.Format("2006-01-02"),
			rec.Description,
			string(rec.TransactionType),
			rec.Category,
			fmt.Sprintf("%.2f", rec.Amount),
			rec.Currency,
			donor,
			rec.PaymentMethod,
			rec.Notes,
			recordedBy,
		})
	}

	cw.Flush()
}

// boolStr converts a bool to "Yes" or "No".
func boolStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
