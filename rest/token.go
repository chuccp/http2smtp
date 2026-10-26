package rest

import (
	"errors"
	"time"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/service"
	"github.com/chuccp/http2smtp/util"
)

type Token struct {
	context      *core.Context
	tokenService *service.TokenService
	tokenModel   *model.TokenModel
}

func (token *Token) getOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	// The list already hands admins every user's tokens (SMTP credentials included), so
	// the row-level read must not refuse them — the token page offers "send test mail"
	// on those rows.
	if user.IsAdmin {
		return token.tokenService.GetOneAny(id)
	}
	return token.tokenService.GetOne(id, user.Id)
}

func (token *Token) deleteOne(req *web.Request) (any, error) {
	id := req.ParamUint("id")
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	tokenModel := core.GetModel[*model.TokenModel](token.context)
	exist, err := tokenModel.Query().Where("id = ? AND user_id = ?", id, user.Id).One()
	if err != nil {
		return nil, err
	}
	if exist == nil {
		return nil, errors.New("token not found")
	}
	err = tokenModel.DeleteByPK(id)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (token *Token) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	name := req.Query("name")
	adminOnly := req.Query("adminOnly") == "true"
	return token.tokenService.GetPage(page, user.Id, user.IsAdmin, name, adminOnly)
}

func (token *Token) postOne(req *web.Request) (any, error) {
	var st model.Token
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	// Create is create: a client-supplied id must not turn this into an UPDATE of an
	// existing row (gorm Save updates when the primary key is set), which would let any
	// signed-in user overwrite another account's token and take ownership of it.
	st.Id = 0
	st.UserId = user.Id
	if err = token.ensureSMTPOwnedBy(st.SMTPId, st.UserId); err != nil {
		return nil, err
	}
	if err = token.ensureRecipientsOwnedBy(st.ReceiveEmailIds, st.UserId); err != nil {
		return nil, err
	}
	st.State = token.resolveState(st.State, entity.TokenStateInUse, user.IsAdmin)
	err = token.tokenModel.Save(&st)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}
func (token *Token) putOne(req *web.Request) (any, error) {
	var st model.Token
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	exist, err := token.tokenModel.FindByPK(st.Id)
	if err != nil {
		return nil, err
	}
	// Report "not found" rather than "forbidden" so we don't leak existence
	if exist == nil || (exist.UserId != user.Id && !user.IsAdmin) {
		return nil, errors.New("token not found")
	}
	// The token and everything it points at must belong to its owner, not to the caller:
	// an admin editing someone else's row keeps the row's owner, and a user may not point
	// a token at another account's SMTP server or contact list.
	if err = token.ensureSMTPOwnedBy(st.SMTPId, exist.UserId); err != nil {
		return nil, err
	}
	if err = token.ensureRecipientsOwnedBy(st.ReceiveEmailIds, exist.UserId); err != nil {
		return nil, err
	}
	exist.Name = st.Name
	exist.Subject = st.Subject
	exist.SMTPId = st.SMTPId
	exist.ReceiveEmailIds = util.DeduplicateIds(st.ReceiveEmailIds)
	// 保留原有 token；前端不编辑令牌值，留空表示不改
	if st.Token != "" {
		exist.Token = st.Token
	}
	exist.State = token.resolveState(st.State, exist.State, user.IsAdmin)
	exist.UpdateTime = time.Now()
	// Save rather than UpdateByPK: the update path drops zero values, so re-enabling a
	// token (state = 0) would be ignored and the row would stay disabled.
	err = token.tokenModel.Save(exist)
	if err != nil {
		return nil, err
	}
	return web.Ok("ok"), nil
}

// ensureSMTPOwnedBy rejects a token that points at an SMTP server its owner does not hold.
// supplementToken resolves the server by id alone, so without this check a user could
// mint a token for someone else's SMTP id and send through that account's credentials
// from the public API.
func (token *Token) ensureSMTPOwnedBy(smtpId uint, userId uint) error {
	if smtpId == 0 {
		return nil
	}
	smtpModel := core.GetModel[*model.SMTPModel](token.context)
	byId, err := smtpModel.Query().Where("id = ? AND user_id = ?", smtpId, userId).One()
	if err != nil {
		return err
	}
	if byId == nil {
		return errors.New("SMTP not found")
	}
	return nil
}

// ensureRecipientsOwnedBy rejects a token whose recipient list references contacts of
// another account. Ids that no longer resolve are tolerated — supplementToken drops
// them as well — but an existing contact belonging to someone else is not.
func (token *Token) ensureRecipientsOwnedBy(receiveEmailIds string, userId uint) error {
	ids := util.StringToUintIds(receiveEmailIds)
	if len(ids) == 0 {
		return nil
	}
	mailModel := core.GetModel[*model.MailModel](token.context)
	mails, err := mailModel.FindAllByPK(ids...)
	if err != nil {
		return err
	}
	for _, mail := range mails {
		if mail.UserId != userId {
			return errors.New("mail not found")
		}
	}
	return nil
}

// resolveState maps the requested state onto the state the caller is allowed to set.
// current is the state already stored on the row (TokenStateInUse for a fresh create).
// Re-enabling is deliberately not the caller's call alone: "admin disabled" is the
// administrator's disposal, so a plain owner asking for state = 0 must not undo it.
func (token *Token) resolveState(state uint8, current uint8, isAdmin bool) uint8 {
	if state == entity.TokenStateInUse {
		if !isAdmin && current == entity.TokenStateAdminDisabled {
			return entity.TokenStateAdminDisabled
		}
		return entity.TokenStateInUse
	}
	// 前端传来非 0 值，后端根据角色决定具体状态
	if isAdmin {
		return entity.TokenStateAdminDisabled
	}
	return entity.TokenStateUserDisabled
}

func (token *Token) sendMail(req *web.Request) (any, error) {
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	return token.tokenService.SendMailByToken(req, user)
}

func (token *Token) sendMailById(req *web.Request) (any, error) {
	var st entity.SendMailByTokenId
	err := req.BindJSON(&st)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, token.context)
	if user == nil {
		return nil, err
	}
	return token.tokenService.SendMailByTokenId(user, &st)
}

func (token *Token) Init(context *core.Context) error {
	token.context = context
	token.tokenService = wf.GetService[*service.TokenService](token.context)
	token.tokenModel = wf.GetModel[*model.TokenModel](token.context)
	token.context.Get("/token/:id", token.getOne).WithMeta(auth2.WithLogin())
	token.context.Delete("/token/:id", token.deleteOne).WithMeta(auth2.WithLogin())
	token.context.Get("/token", token.getPage).WithMeta(auth2.WithLogin())
	token.context.Post("/token", token.postOne).WithMeta(auth2.WithLogin())
	token.context.Put("/token", token.putOne).WithMeta(auth2.WithLogin())
	token.context.Post("/sendMailByToken", token.sendMail).WithMeta(auth2.WithLogin())
	token.context.Post("/sendMailByTokenId", token.sendMailById).WithMeta(auth2.WithLogin())
	return nil
}
