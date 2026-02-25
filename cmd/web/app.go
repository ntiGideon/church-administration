package main

import (
	"html/template"
	"log/slog"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/minio/minio-go/v7"
	"github.com/ntiGideon/ent"
	"github.com/ntiGideon/internal/models"
)

type application struct {
	logger         *slog.Logger
	db             *ent.Client
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	formDecoder    *form.Decoder
	minioClient       *minio.Client
	minioBucket       string
	churchModel       *models.ChurchModel
	userModel         *models.UserModel
	memberModel       *models.MemberModel
	eventModel        *models.EventModel
	financeModel      *models.FinanceModel
	invitationModel   *models.InvitationModel
	announcementModel *models.AnnouncementModel
}
