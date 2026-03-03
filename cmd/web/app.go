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
	programModel      *models.ProgramModel
	attendanceModel   *models.AttendanceModel
	groupModel        *models.GroupModel
	pledgeModel       *models.PledgeModel
	rosterModel       *models.RosterModel
	sermonModel       *models.SermonModel
	visitorModel      *models.VisitorModel
	prayerModel       *models.PrayerRequestModel
	documentModel     *models.DocumentModel
	pastoralModel     *models.PastoralNoteModel
	milestoneModel     *models.MilestoneModel
	departmentModel    *models.DepartmentModel
	relationshipModel    *models.RelationshipModel
	communicationModel   *models.CommunicationModel
	budgetModel          *models.BudgetModel
	customRoleModel      *models.CustomRoleModel
}
