package entity

type LoginUser struct {
	Id       uint
	Username string
	Password string
	Response string
	Nonce    string
	Salt     string
	IsAdmin  bool
}
