package permissions

// Permission key constants
const (
	// Members
	ViewMembers       = "view_members"
	ManageMembers     = "manage_members"
	ViewMemberGiving  = "view_member_giving"
	ImportMembers     = "import_members"

	// Events
	ViewEvents     = "view_events"
	ManageEvents   = "manage_events"
	TakeAttendance = "take_attendance"

	// Finance
	ViewFinance        = "view_finance"
	RecordFinance      = "record_finance"
	ViewBudgets        = "view_budgets"
	ManageBudgets      = "manage_budgets"
	ExportFinance      = "export_finance"

	// Ministry
	ViewGroups       = "view_groups"
	ManageGroups     = "manage_groups"
	ViewRosters      = "view_rosters"
	ManageRosters    = "manage_rosters"
	ViewSermons      = "view_sermons"
	ManageSermons    = "manage_sermons"
	ViewPrayer       = "view_prayer"
	ManagePrayer     = "manage_prayer"
	ViewPastoral     = "view_pastoral"
	ManagePastoral   = "manage_pastoral"
	ViewDepartments  = "view_departments"
	ManageDepartments = "manage_departments"

	// Community
	ViewCalendar         = "view_calendar"
	ManageCalendar       = "manage_calendar"
	ViewAnnouncements    = "view_announcements"
	ManageAnnouncements  = "manage_announcements"
	ViewVisitors         = "view_visitors"
	ManageVisitors       = "manage_visitors"
	ViewDocuments        = "view_documents"
	ManageDocuments      = "manage_documents"
	ViewMilestones       = "view_milestones"
	SendCommunications   = "send_communications"

	// Reports
	ViewReports           = "view_reports"
	ViewFinanceReports    = "view_finance_reports"
	ViewAttendanceReports = "view_attendance_reports"

	// Workers
	ViewWorkers   = "view_workers"
	InviteWorkers = "invite_workers"
)

// PermissionKey is a permission key with its display label.
type PermissionKey struct {
	Key   string
	Label string
}

// PermissionGroup groups related permissions under a heading.
type PermissionGroup struct {
	Name        string
	Permissions []PermissionKey
}

// PermissionGroups is the ordered list of permission groups used for UI rendering.
var PermissionGroups = []PermissionGroup{
	{
		Name: "Members",
		Permissions: []PermissionKey{
			{ViewMembers, "View Members"},
			{ManageMembers, "Manage Members"},
			{ViewMemberGiving, "View Member Giving"},
			{ImportMembers, "Import Members"},
		},
	},
	{
		Name: "Events",
		Permissions: []PermissionKey{
			{ViewEvents, "View Events"},
			{ManageEvents, "Manage Events"},
			{TakeAttendance, "Take Attendance"},
		},
	},
	{
		Name: "Finance",
		Permissions: []PermissionKey{
			{ViewFinance, "View Finance"},
			{RecordFinance, "Record Finance"},
			{ViewBudgets, "View Budgets"},
			{ManageBudgets, "Manage Budgets"},
			{ExportFinance, "Export Finance"},
		},
	},
	{
		Name: "Ministry",
		Permissions: []PermissionKey{
			{ViewGroups, "View Groups"},
			{ManageGroups, "Manage Groups"},
			{ViewRosters, "View Rosters"},
			{ManageRosters, "Manage Rosters"},
			{ViewSermons, "View Sermons"},
			{ManageSermons, "Manage Sermons"},
			{ViewPrayer, "View Prayer"},
			{ManagePrayer, "Manage Prayer"},
			{ViewPastoral, "View Pastoral Care"},
			{ManagePastoral, "Manage Pastoral Care"},
			{ViewDepartments, "View Departments"},
			{ManageDepartments, "Manage Departments"},
		},
	},
	{
		Name: "Community",
		Permissions: []PermissionKey{
			{ViewCalendar, "View Calendar"},
			{ManageCalendar, "Manage Calendar"},
			{ViewAnnouncements, "View Announcements"},
			{ManageAnnouncements, "Manage Announcements"},
			{ViewVisitors, "View Visitors"},
			{ManageVisitors, "Manage Visitors"},
			{ViewDocuments, "View Documents"},
			{ManageDocuments, "Manage Documents"},
			{ViewMilestones, "View Milestones"},
			{SendCommunications, "Send Communications"},
		},
	},
	{
		Name: "Reports",
		Permissions: []PermissionKey{
			{ViewReports, "View Reports"},
			{ViewFinanceReports, "View Financial Reports"},
			{ViewAttendanceReports, "View Attendance Reports"},
		},
	},
	{
		Name: "Workers",
		Permissions: []PermissionKey{
			{ViewWorkers, "View Workers"},
			{InviteWorkers, "Invite Workers"},
		},
	},
}

// allPerms returns a map with every permission set to true.
func allPerms() map[string]bool {
	m := make(map[string]bool)
	for _, g := range PermissionGroups {
		for _, p := range g.Permissions {
			m[p.Key] = true
		}
	}
	return m
}

// DefaultPermissionsForRole returns a sensible default permission set for each
// built-in enum role. Admins always pass via a separate code path; this handles
// the remaining roles so degraded custom-role workers still have basic access.
func DefaultPermissionsForRole(role string) map[string]bool {
	switch role {
	case "super_admin", "branch_admin":
		return allPerms()

	case "pastor":
		return map[string]bool{
			ViewMembers:           true,
			ViewMemberGiving:      true,
			ViewEvents:            true,
			TakeAttendance:        true,
			ViewFinance:           true,
			ViewBudgets:           true,
			ViewGroups:            true,
			ViewRosters:           true,
			ViewSermons:           true,
			ManageSermons:         true,
			ViewPrayer:            true,
			ManagePrayer:          true,
			ViewPastoral:          true,
			ManagePastoral:        true,
			ViewDepartments:       true,
			ViewCalendar:          true,
			ViewAnnouncements:     true,
			ViewVisitors:          true,
			ViewDocuments:         true,
			ViewMilestones:        true,
			ViewReports:           true,
			ViewAttendanceReports: true,
			ViewWorkers:           true,
		}

	case "secretary":
		return map[string]bool{
			ViewMembers:           true,
			ManageMembers:         true,
			ImportMembers:         true,
			ViewEvents:            true,
			ManageEvents:          true,
			TakeAttendance:        true,
			ViewGroups:            true,
			ViewRosters:           true,
			ManageRosters:         true,
			ViewDepartments:       true,
			ViewCalendar:          true,
			ManageCalendar:        true,
			ViewAnnouncements:     true,
			ManageAnnouncements:   true,
			ViewVisitors:          true,
			ManageVisitors:        true,
			ViewDocuments:         true,
			ManageDocuments:       true,
			ViewMilestones:        true,
			ViewReports:           true,
			ViewAttendanceReports: true,
			ViewWorkers:           true,
		}

	case "records_keeper":
		return map[string]bool{
			ViewMembers:           true,
			ManageMembers:         true,
			ImportMembers:         true,
			ViewMemberGiving:      true,
			ViewEvents:            true,
			TakeAttendance:        true,
			ViewGroups:            true,
			ViewDepartments:       true,
			ViewCalendar:          true,
			ViewAnnouncements:     true,
			ViewDocuments:         true,
			ManageDocuments:       true,
			ViewMilestones:        true,
			ViewReports:           true,
			ViewAttendanceReports: true,
		}

	case "finance_officer":
		return map[string]bool{
			ViewMembers:         true,
			ViewMemberGiving:    true,
			ViewFinance:         true,
			RecordFinance:       true,
			ViewBudgets:         true,
			ManageBudgets:       true,
			ExportFinance:       true,
			ViewReports:         true,
			ViewFinanceReports:  true,
		}

	case "member":
		return map[string]bool{
			ViewMembers:       true,
			ViewEvents:        true,
			ViewSermons:       true,
			ViewPrayer:        true,
			ViewCalendar:      true,
			ViewAnnouncements: true,
		}

	default:
		return map[string]bool{}
	}
}
