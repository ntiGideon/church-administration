package main

import (
	"net/http"

	"github.com/ntiGideon/ent/church"
	"github.com/ntiGideon/ent/user"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /login
func (app *application) loginGet(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.LoginDto{}
	app.render(w, r, http.StatusOK, "login.gohtml", data)
}

// POST /login
func (app *application) loginPost(w http.ResponseWriter, r *http.Request) {
	var dto models.LoginDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Email), "email", "Email is required")
	dto.CheckField(validator.Matches(dto.Email, validator.EmailRX), "email", "Must be a valid email")
	dto.CheckField(validator.NotBlank(dto.Password), "password", "Password is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		return
	}

	u, err := app.userModel.Authenticate(r.Context(), dto.Email, dto.Password)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			dto.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = dto
			app.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
			return
		}
		app.serverError(w, r, err)
		return
	}

	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", u.ID)

	// Store church ID in session for easy access
	if edges := u.Edges.Church; edges != nil {
		app.sessionManager.Put(r.Context(), "authenticatedChurchID", edges.ID)
	}

	redirectURL := app.sessionManager.PopString(r.Context(), "requestedPathAfterLogin")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// POST /logout
func (app *application) logoutPost(w http.ResponseWriter, r *http.Request) {
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Remove(r.Context(), "authenticatedChurchID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// GET /register?token=<jwt>
func (app *application) registerGet(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.sessionManager.Put(r.Context(), "flash", "Invalid or missing registration token.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	email := app.churchModel.ExtractEmailFromToken(token)
	data := app.newTemplateData(r)
	data.Form = models.RegisterDto{RegistrationToken: token, Email: email}
	app.render(w, r, http.StatusOK, "register.gohtml", data)
}

// POST /register
func (app *application) registerPost(w http.ResponseWriter, r *http.Request) {
	var dto models.RegisterDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")
	dto.CheckField(validator.NotBlank(dto.Email), "email", "Email is required")
	dto.CheckField(validator.Matches(dto.Email, validator.EmailRX), "email", "Must be a valid email")
	dto.CheckField(validator.NotBlank(dto.Password), "password", "Password is required")
	dto.CheckField(validator.MinChars(dto.Password, 8), "password", "Password must be at least 8 characters")
	dto.CheckField(dto.Password == dto.ConfirmPassword, "confirm_password", "Passwords do not match")
	dto.CheckField(validator.NotBlank(dto.RegistrationToken), "registration_token", "A valid registration token is required")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "register.gohtml", data)
		return
	}

	// Find the church by registration token
	c, err := app.db.Church.Query().
		Where(church.RegistrationTokenEQ(dto.RegistrationToken)).
		Only(r.Context())
	if err != nil {
		dto.AddNonFieldError("Invalid or expired registration token. Please contact the administrator.")
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "register.gohtml", data)
		return
	}

	// Verify the JWT token
	valid, err := app.churchModel.VerifyToken(dto.RegistrationToken, c.Email)
	if !valid || err != nil {
		dto.AddNonFieldError("This invitation link has expired. Please request a new one.")
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "register.gohtml", data)
		return
	}

	// Always use the church's registered email — not whatever the form submitted
	dto.Email = c.Email

	// Create the branch admin user
	userID, err := app.userModel.Create(r.Context(), &dto, c.ID, user.RoleBranchAdmin)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	_ = userID

	// Clear the registration token so it can't be reused
	_, _ = app.db.Church.UpdateOne(c).ClearRegistrationToken().Save(r.Context())

	app.sessionManager.Put(r.Context(), "flash", "Registration successful! Please sign in.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// GET /invite/accept?token=<jwt>  — accept a member/staff invitation
func (app *application) inviteAcceptGet(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	data := app.newTemplateData(r)
	data.Form = models.InviteAcceptDto{Token: token}
	app.render(w, r, http.StatusOK, "invite_accept.gohtml", data)
}

// POST /invite/accept
func (app *application) inviteAcceptPost(w http.ResponseWriter, r *http.Request) {
	var dto models.InviteAcceptDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.FirstName), "first_name", "First name is required")
	dto.CheckField(validator.NotBlank(dto.LastName), "last_name", "Last name is required")
	dto.CheckField(validator.NotBlank(dto.Password), "password", "Password is required")
	dto.CheckField(validator.MinChars(dto.Password, 8), "password", "At least 8 characters required")
	dto.CheckField(dto.Password == dto.ConfirmPassword, "confirm_password", "Passwords do not match")
	dto.CheckField(validator.NotBlank(dto.Token), "token", "Invalid invitation token")

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "invite_accept.gohtml", data)
		return
	}

	// Find the invitation via the invitation model (validates status + expiry)
	inv, err := app.invitationModel.GetByToken(r.Context(), dto.Token)
	if err != nil {
		msg := "Invalid or expired invitation link."
		if err == models.ErrInvitationExpired {
			msg = "This invitation has expired. Please request a new one."
		}
		dto.AddNonFieldError(msg)
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "invite_accept.gohtml", data)
		return
	}

	regDto := &models.RegisterDto{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     inv.InviteeEmail,
		Phone:     dto.Phone,
		Password:  dto.Password,
	}

	churchID := inv.Edges.Church.ID
	userID, err := app.userModel.Create(r.Context(), regDto, churchID, user.Role(inv.Role))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Mark the invitation as accepted
	_ = app.invitationModel.MarkAccepted(r.Context(), inv.ID, userID)

	app.sessionManager.Put(r.Context(), "flash", "Account created! You can now sign in.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
