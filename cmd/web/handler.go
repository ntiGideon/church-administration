package main

import (
	"fmt"
	"net/http"

	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/ent/user"
	"github.com/ntiGideon/internal/models"
	"github.com/ntiGideon/internal/validator"
)

// GET /
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.clientError(w, http.StatusNotFound)
		return
	}
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "home.gohtml", data)
}

// GET /invite-church
func (app *application) inviteChurch(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.InviteDto{}
	app.render(w, r, http.StatusOK, "invite_church.gohtml", data)
}

// POST /invite-church
func (app *application) inviteChurchPost(w http.ResponseWriter, r *http.Request) {
	var dto models.InviteDto
	if err := app.decodePostForm(r, &dto); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dto.CheckField(validator.NotBlank(dto.Name), "name", models.CannotBeBlankField)
	dto.CheckField(validator.NotBlank(dto.Email), "email", models.CannotBeBlankField)
	dto.CheckField(validator.NotBlank(dto.Address), "address", models.CannotBeBlankField)
	dto.CheckField(validator.NotBlank(dto.Branch), "branch", models.CannotBeBlankField)
	dto.CheckField(validator.Matches(dto.Email, validator.EmailRX), "email", models.ValidEmail)

	if dto.ExpiresAt <= 0 {
		dto.ExpiresAt = 72
	}

	if !dto.Valid() {
		data := app.newTemplateData(r)
		data.Form = dto
		app.render(w, r, http.StatusUnprocessableEntity, "invite_church.gohtml", data)
		return
	}

	resp := app.churchModel.InviteChurch(r.Context(), &dto)
	if resp.Error != nil {
		if resp.Error == models.EmailAlreadyExist {
			dto.AddFieldError("email", "A church with this email already exists")
			data := app.newTemplateData(r)
			data.Form = dto
			app.render(w, r, http.StatusUnprocessableEntity, "invite_church.gohtml", data)
			return
		}
		app.serverError(w, r, resp.Error)
		return
	}

	result, ok := resp.Data.(struct {
		InviteToken string
		ChurchName  string
	})
	if ok {
		registrationURL := fmt.Sprintf("http://localhost:3000/register?token=%s", result.InviteToken)
		go func() {
			_ = sendHTMLEmail(
				dto.Email,
				"You're Invited to Join the FaithConnect Church Network",
				buildChurchInviteEmail(result.ChurchName, dto.Email, registrationURL),
			)
		}()
	}

	app.sessionManager.Put(r.Context(), "flash", "Invitation sent to "+dto.Email)
	http.Redirect(w, r, "/admin/churches", http.StatusSeeOther)
}

// GET /admin/churches  — list all churches with stats (superadmin only)
func (app *application) adminChurches(w http.ResponseWriter, r *http.Request) {
	churches, err := app.db.Church.Query().
		WithUsers(func(uq *ent.UserQuery) {
			uq.Where(user.IsActiveEQ(true))
		}).
		All(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	totalMembers := 0
	pendingCount := 0
	for _, c := range churches {
		totalMembers += len(c.Edges.Users)
		if c.RegistrationToken != nil {
			pendingCount++
		}
	}

	data := app.newTemplateData(r)
	data.Data = map[string]interface{}{
		"churches":     churches,
		"totalMembers": totalMembers,
		"pendingCount": pendingCount,
	}
	app.render(w, r, http.StatusOK, "admin_churches.gohtml", data)
}

// buildChurchInviteEmail returns a branded HTML email for church invitations.
func buildChurchInviteEmail(churchName, adminEmail, registrationURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#f5f5f5;font-family:'Helvetica Neue',Arial,sans-serif">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f5f5f5;padding:40px 0">
  <tr><td align="center">
    <table width="580" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 20px rgba(0,0,0,0.08)">
      <!-- Header -->
      <tr><td style="background:linear-gradient(135deg,#50222D,#6B2F3D);padding:36px 40px;text-align:center">
        <h1 style="color:#ffffff;font-size:24px;font-weight:800;margin:0">FaithConnect</h1>
        <p style="color:rgba(255,255,255,0.8);font-size:13px;margin:6px 0 0">Church Management Platform</p>
      </td></tr>
      <!-- Body -->
      <tr><td style="padding:40px">
        <h2 style="color:#50222D;font-size:22px;font-weight:700;margin:0 0 16px">You've Been Invited!</h2>
        <p style="color:#444;font-size:15px;line-height:1.6;margin:0 0 16px">
          Your church, <strong>%s</strong>, has been invited to join the FaithConnect Church Management Network.
        </p>
        <p style="color:#444;font-size:15px;line-height:1.6;margin:0 0 24px">
          As the church administrator, you'll be able to:
        </p>
        <ul style="color:#555;font-size:14px;line-height:1.8;margin:0 0 28px;padding-left:20px">
          <li>Manage your church's members and staff</li>
          <li>Record and track events and service attendance</li>
          <li>Monitor financial giving and expenses</li>
          <li>Publish announcements to your congregation</li>
          <li>View detailed reports and analytics</li>
        </ul>
        <p style="color:#666;font-size:13px;margin:0 0 28px">
          This invitation was sent to <strong>%s</strong>. Click the button below to set up your account.
        </p>
        <div style="text-align:center;margin:0 0 28px">
          <a href="%s" style="display:inline-block;background:linear-gradient(135deg,#50222D,#6B2F3D);color:#ffffff;font-size:15px;font-weight:700;padding:14px 36px;border-radius:10px;text-decoration:none;letter-spacing:0.3px">
            Create My Account &rarr;
          </a>
        </div>
        <div style="background:#fff8f9;border:1px solid #f0d5d9;border-radius:8px;padding:14px 18px;margin-bottom:24px">
          <p style="color:#666;font-size:12px;margin:0">
            <strong>Note:</strong> This link will expire in 72 hours. If you did not expect this invitation, please ignore this email.
          </p>
        </div>
        <p style="color:#888;font-size:12px">
          Or copy this link into your browser:<br>
          <span style="color:#50222D;word-break:break-all">%s</span>
        </p>
      </td></tr>
      <!-- Footer -->
      <tr><td style="background:#f9f9f9;padding:20px 40px;text-align:center;border-top:1px solid #eee">
        <p style="color:#aaa;font-size:12px;margin:0">FaithConnect &bull; Church Administration Platform</p>
      </td></tr>
    </table>
  </td></tr>
</table>
</body>
</html>`, churchName, adminEmail, registrationURL, registrationURL)
}
