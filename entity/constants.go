package entity

const (
	Pi            = 3.14159
	Max           = 100
	Name          = "Golang"
	ValidatePhone = "_ValidatePhone_"
	CodeService   = "code_service"
	UserService   = "user_service"
	QrcodeService = "qrcode_service"
	UserModel     = "user_model"
	HistoryModel  = "history_model"
	UserToken     = "user_token"
	AuthService   = "auth_service"
	UserRest      = "user_rest"
	TokenRest     = "token_rest"
	QrcodeRest    = "qrcode_rest"
	QrcodeModel   = "qrcode_model"
)
const (
	SUCCESS byte = iota
	WARM
	ERROR
)

// Token state constants
const (
	TokenStateInUse       uint8 = 0 // 使用中
	TokenStateUserDisabled  uint8 = 1 // 用户禁用
	TokenStateAdminDisabled uint8 = 2 // 管理员禁用
)
