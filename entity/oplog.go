package entity

// Operation-log actions and target types. The database stores these stable values; the
// UI translates them for display.
const (
	OpCreate  = "create"
	OpUpdate  = "update"
	OpDelete  = "delete"
	OpEnable  = "enable"
	OpDisable = "disable"
	OpTrigger = "trigger"
	OpCleanup = "cleanup"
	OpRestart = "restart"
	OpLogin   = "login"
)

const (
	TargetSMTP     = "smtp"
	TargetMail     = "mail"
	TargetToken    = "token"
	TargetSchedule = "schedule"
	TargetUser     = "user"
	TargetLog      = "log"
	TargetSettings = "settings"
)
