package main

import (
	"net/http"

	entcontact "github.com/ntiGideon/ent/contact"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GET /profile
func (app *application) profileGet(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	dto := models.ProfileDto{}
	if c := u.Edges.Contact; c != nil {
		dto.FirstName = c.FirstName
		dto.LastName = c.LastName
		dto.Phone = c.Phone
		dto.Gender = string(c.Gender)
		dto.Occupation = c.Occupation
		dto.City = c.City
		dto.Country = c.Country
		dto.IDNumber = c.IDNumber
		dto.Hometown = c.Hometown
		dto.Region = c.Region
		dto.SundaySchoolClass = c.SundaySchoolClass
		dto.DayBorn = string(c.DayBorn)
		dto.HasSpouse = c.HasSpouse
	}

	data := app.newTemplateData(r)
	data.Form = dto
	app.render(w, r, http.StatusOK, "profile.gohtml", data)
}



// POST /profile
func (app *application) profilePost(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var dto models.ProfileDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "profile.gohtml", data)
		return
	}

	if u.Edges.Contact == nil {
		app.serverError(w, r, models.ErrUserNotFound)
		return
	}

	upd := app.db.Contact.UpdateOneID(u.Edges.Contact.ID).
		SetFirstName(dto.FirstName).
		SetLastName(dto.LastName).
		SetNillablePhone(nullStr(dto.Phone)).
		SetNillableOccupation(nullStr(dto.Occupation)).
		SetNillableCity(nullStr(dto.City)).
		SetNillableCountry(nullStr(dto.Country)).
		SetNillableIDNumber(nullStr(dto.IDNumber)).
		SetNillableHometown(nullStr(dto.Hometown)).
		SetNillableRegion(nullStr(dto.Region)).
		SetNillableSundaySchoolClass(nullStr(dto.SundaySchoolClass)).
		SetHasSpouse(dto.HasSpouse)

	if dto.Gender != "" {
		upd = upd.SetGender(entcontact.Gender(dto.Gender))
	} else {
		upd = upd.ClearGender()
	}
	if dto.DayBorn != "" {
		upd = upd.SetDayBorn(entcontact.DayBorn(dto.DayBorn))
	} else {
		upd = upd.ClearDayBorn()
	}

	if _, err := upd.Save(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Profile updated successfully.")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// POST /profile/avatar
func (app *application) profileAvatarPost(w http.ResponseWriter, r *http.Request) {
	u := app.getAuthenticatedUser(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseMultipartForm(maxImageInputSize); err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "File too large — maximum accepted size is 20 MB.")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", "No file selected.")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	defer file.Close()

	url, err := app.uploadImage(file, header, "avatars")
	if err != nil {
		app.sessionManager.Put(r.Context(), "flash_error", err.Error())
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if u.Edges.Contact != nil {
		if _, err := app.db.Contact.UpdateOneID(u.Edges.Contact.ID).
			SetProfilePictureURL(url).
			Save(r.Context()); err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	app.sessionManager.Put(r.Context(), "flash", "Profile picture updated successfully.")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
