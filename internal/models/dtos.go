package models

import "github.com/ntiGideon/internal/validator"

type ModelResponse struct {
	Data  interface{}
	Error error
}

// LoginDto is used for user login form
type LoginDto struct {
	Email    string `form:"email"`
	Password string `form:"password"`

	validator.Validator `form:"-"`
}

// RegisterDto is used when a church admin accepts an invitation and creates their account
type RegisterDto struct {
	RegistrationToken string `form:"registration_token"`
	FirstName         string `form:"first_name"`
	LastName          string `form:"last_name"`
	Email             string `form:"email"`
	Phone             string `form:"phone"`
	Password          string `form:"password"`
	ConfirmPassword   string `form:"confirm_password"`

	validator.Validator `form:"-"`
}

// InviteDto is used by superadmin to invite a new church branch
type InviteDto struct {
	Email     string `form:"email"`
	Address   string `form:"address"`
	Name      string `form:"name"`
	Branch    string `form:"branch"`
	ExpiresAt int    `form:"expires_at"`

	validator.Validator `form:"-"`
}

// MemberInviteDto is used by branch admin to invite staff/members
type MemberInviteDto struct {
	Email string `form:"email"`
	Name  string `form:"name"`
	Role  string `form:"role"`

	validator.Validator `form:"-"`
}

// InviteAcceptDto is used by invited staff/members to complete their registration
type InviteAcceptDto struct {
	Token           string `form:"token"`
	FirstName       string `form:"first_name"`
	LastName        string `form:"last_name"`
	Phone           string `form:"phone"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`

	validator.Validator `form:"-"`
}

// EventDto is used to create a new event
type EventDto struct {
	Title       string `form:"title"`
	Description string `form:"description"`
	StartTime   string `form:"start_time"`
	EndTime     string `form:"end_time"`
	Location    string `form:"location"`
	EventType   string `form:"event_type"`

	validator.Validator `form:"-"`
}

// FinanceDto is used to record a financial transaction
type FinanceDto struct {
	Description     string  `form:"description"`
	TransactionType string  `form:"transaction_type"`
	Amount          float64 `form:"amount"`
	Currency        string  `form:"currency"`
	TransactionDate string  `form:"transaction_date"`
	Category        string  `form:"category"`
	PaymentMethod   string  `form:"payment_method"`
	Notes           string  `form:"notes"`
	ContactID       int     `form:"contact_id"`

	validator.Validator `form:"-"`
}

// PledgeDto is used to create a pledge for a member.
type PledgeDto struct {
	Title     string  `form:"title"`
	Category  string  `form:"category"`
	Amount    float64 `form:"amount"`
	Currency  string  `form:"currency"`
	StartDate string  `form:"start_date"`
	EndDate   string  `form:"end_date"`
	Frequency string  `form:"frequency"`
	Notes     string  `form:"notes"`

	validator.Validator `form:"-"`
}

// ProgramDto is used to create/update a church calendar program entry.
type ProgramDto struct {
	Title             string `form:"title"`
	ProgramType       string `form:"program_type"`
	Date              string `form:"date"`
	Theme             string `form:"theme"`
	SermonTopic       string `form:"sermon_topic"`
	VisionGoals       string `form:"vision_goals"`
	Preacher          string `form:"preacher"`
	OpeningPrayerBy   string `form:"opening_prayer_by"`
	ClosingPrayerBy   string `form:"closing_prayer_by"`
	WorshipLeader     string `form:"worship_leader"`
	ResponsiblePerson string `form:"responsible_person"`
	Notes             string `form:"notes"`
	IsPublished       bool   `form:"is_published"`

	validator.Validator `form:"-"`
}

// AnnouncementDto is used to create an announcement
type AnnouncementDto struct {
	Title       string `form:"title"`
	Content     string `form:"content"`
	Category    string `form:"category"`
	IsPublished bool   `form:"is_published"`

	validator.Validator `form:"-"`
}

// ProfileDto is used to update the logged-in user's contact info
type ProfileDto struct {
	FirstName         string `form:"first_name"`
	LastName          string `form:"last_name"`
	Phone             string `form:"phone"`
	Gender            string `form:"gender"`
	Occupation        string `form:"occupation"`
	City              string `form:"city"`
	Country           string `form:"country"`
	IDNumber          string `form:"id_number"`
	Hometown          string `form:"hometown"`
	Region            string `form:"region"`
	SundaySchoolClass string `form:"sunday_school_class"`
	DayBorn           string `form:"day_born"`
	HasSpouse         bool   `form:"has_spouse"`

	validator.Validator `form:"-"`
}

// ChurchSettingsDto is used by branch admin to update their church profile
type ChurchSettingsDto struct {
	Name             string `form:"name"`
	Address          string `form:"address"`
	City             string `form:"city"`
	Country          string `form:"country"`
	Phone            string `form:"phone"`
	Website          string `form:"website"`
	CongregationSize int    `form:"congregation_size"`

	validator.Validator `form:"-"`
}

// RosterDto is used to create or update a duty roster.
type RosterDto struct {
	Title       string `form:"title"`
	ServiceDate string `form:"service_date"`
	Department  string `form:"department"`
	Notes       string `form:"notes"`

	validator.Validator `form:"-"`
}

// RosterEntryDto is used to assign a volunteer to a roster.
type RosterEntryDto struct {
	ContactID int    `form:"contact_id"`
	Role      string `form:"role"`
	Notes     string `form:"notes"`

	validator.Validator `form:"-"`
}

// DocumentDto is used when uploading a document.
type DocumentDto struct {
	Title       string `form:"title"`
	Description string `form:"description"`
	Category    string `form:"category"`
	IsPublic    bool   `form:"is_public"`

	validator.Validator `form:"-"`
}

// PrayerRequestDto is used to create or update a prayer request.
type PrayerRequestDto struct {
	Title         string `form:"title"`
	Body          string `form:"body"`
	RequesterName string `form:"requester_name"`
	IsAnonymous   bool   `form:"is_anonymous"`
	IsPrivate     bool   `form:"is_private"`
	ContactID     int    `form:"contact_id"`

	validator.Validator `form:"-"`
}

// VisitorDto is used to create or update a visitor record.
type VisitorDto struct {
	FirstName      string `form:"first_name"`
	LastName       string `form:"last_name"`
	Email          string `form:"email"`
	Phone          string `form:"phone"`
	Address        string `form:"address"`
	VisitDate      string `form:"visit_date"`
	HowHeard       string `form:"how_heard"`
	InvitedBy      string `form:"invited_by"`
	Notes          string `form:"notes"`
	FollowUpStatus string `form:"follow_up_status"`

	validator.Validator `form:"-"`
}

// SermonDto is used to create or update a sermon record.
type SermonDto struct {
	Title       string `form:"title"`
	Speaker     string `form:"speaker"`
	Series      string `form:"series"`
	Scripture   string `form:"scripture"`
	Description string `form:"description"`
	MediaURL    string `form:"media_url"`
	ServiceDate string `form:"service_date"`
	IsPublished bool   `form:"is_published"`

	validator.Validator `form:"-"`
}

// GroupDto is used to create or update a small group
type GroupDto struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	GroupType   string `form:"group_type"`
	MeetingDay  string `form:"meeting_day"`
	MeetingTime string `form:"meeting_time"`
	Location    string `form:"location"`
	LeaderID    int    `form:"leader_id"`
	IsActive    bool   `form:"is_active"`

	validator.Validator `form:"-"`
}

// PastoralNoteDto is used to create or update a pastoral care note.
type PastoralNoteDto struct {
	VisitDate     string `form:"visit_date"`
	CareType      string `form:"care_type"`
	Notes         string `form:"notes"`
	NeedsFollowUp bool   `form:"needs_follow_up"`
	FollowUpDate  string `form:"follow_up_date"`
	ContactID     int    `form:"contact_id"`

	validator.Validator `form:"-"`
}

// MemberDto is used for creating/editing a member profile
type MemberDto struct {
	FirstName    string `form:"first_name"`
	LastName     string `form:"last_name"`
	MiddleName   string `form:"middle_name"`
	Email        string `form:"email"`
	Phone        string `form:"phone"`
	Gender       string `form:"gender"`
	DateOfBirth  string `form:"date_of_birth"`
	MaritalStatus string `form:"marital_status"`
	Occupation   string `form:"occupation"`
	AddressLine1 string `form:"address_line1"`
	City         string `form:"city"`
	Country      string `form:"country"`

	// Church & identity records
	IDNumber          string `form:"id_number"`
	Hometown          string `form:"hometown"`
	Region            string `form:"region"`
	SundaySchoolClass string `form:"sunday_school_class"`
	DayBorn           string `form:"day_born"`
	MembershipYear    int    `form:"membership_year"`

	// Spouse
	HasSpouse bool `form:"has_spouse"`
	SpouseID  int  `form:"spouse_id"`

	// Baptism
	IsBaptized        bool   `form:"is_baptized"`
	BaptizedBy        string `form:"baptized_by"`
	BaptismChurch     string `form:"baptism_church"`
	BaptismCertNumber string `form:"baptism_cert_number"`
	BaptismDate       string `form:"baptism_date"`

	validator.Validator `form:"-"`
}

type DepartmentDto struct {
	Name           string `form:"name"`
	Description    string `form:"description"`
	DepartmentType string `form:"department_type"`
	LeaderID       int    `form:"leader_id"`
	IsActive       bool   `form:"is_active"`

	validator.Validator `form:"-"`
}

type MilestoneDto struct {
	MilestoneType string `form:"milestone_type"`
	EventDate     string `form:"event_date"`
	Description   string `form:"description"`
	OfficiatedBy  string `form:"officiated_by"`

	validator.Validator `form:"-"`
}

type RelationshipDto struct {
	RelativeID   int    `form:"relative_id"`
	RelationType string `form:"relation_type"`
	Notes        string `form:"notes"`

	validator.Validator `form:"-"`
}

type CommunicationDto struct {
	Subject         string `form:"subject"`
	Body            string `form:"body"`
	RecipientFilter string `form:"recipient_filter"`

	validator.Validator `form:"-"`
}

// BudgetDto is used to create or update a budget plan.
type BudgetDto struct {
	Name       string `form:"name"`
	FiscalYear int    `form:"fiscal_year"`
	Period     string `form:"period"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Status     string `form:"status"`
	Notes      string `form:"notes"`

	validator.Validator `form:"-"`
}

// BudgetLineDto is used to add a line item to a budget.
type BudgetLineDto struct {
	Category        string  `form:"category"`
	LineType        string  `form:"line_type"`
	AllocatedAmount float64 `form:"allocated_amount"`
	Currency        string  `form:"currency"`
	Notes           string  `form:"notes"`

	validator.Validator `form:"-"`
}
