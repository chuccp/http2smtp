package service

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"sync"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/smtp"
	"go.uber.org/zap"
)

type TokenService struct {
	context    *core.Context
	mailModel  *model.MailModel
	smtpModel  *model.SMTPModel
	tokenModel *model.TokenModel
	logService *LogService
	lock       *sync.RWMutex
	cachePath  string
}

func (l *TokenService) GetOne(id uint, userId uint) (*model.Token, error) {
	tokenModel := core.GetModel[*model.TokenModel](l.context)
	token, err := tokenModel.Query().Where("id = ? AND user_id = ?", id, userId).One()
	if err != nil {
		return nil, err
	}
	if token != nil {
		l.supplementToken(token)
	}
	return token, nil
}

// GetOneAny loads a token without the owner filter, for admin views. The list endpoint
// already shows admins every user's tokens, so this exposes nothing extra.
func (l *TokenService) GetOneAny(id uint) (*model.Token, error) {
	tokenModel := core.GetModel[*model.TokenModel](l.context)
	token, err := tokenModel.Query().Where("id = ?", id).One()
	if err != nil {
		return nil, err
	}
	if token != nil {
		l.supplementToken(token)
	}
	return token, nil
}
func (l *TokenService) GetPage(page *web.Page, userId uint, isAdmin bool, name string, adminOnly bool) (any, error) {
	var tokens []*model.Token
	var i int
	var err error
	if isAdmin {
		query := l.tokenModel.Query()
		if adminOnly {
			query = query.Where("user_id IN (SELECT id FROM t_user WHERE is_admin = ?)", true)
		}
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		tokens, i, err = query.Page(page)
	} else {
		query := l.tokenModel.Query().Where("user_id = ?", userId)
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		tokens, i, err = query.Page(page)
	}
	if err != nil {
		return nil, err
	}
	l.supplementToken(tokens...)
	if isAdmin {
		userIds := make([]uint, 0)
		for _, t := range tokens {
			if t.UserId > 0 {
				userIds = append(userIds, t.UserId)
			}
		}
		userService := wf.GetService[*UserService](l.context)
		userService.FillUserNames(userIds, func(uid uint, name string) {
			for _, t := range tokens {
				if t.UserId == uid {
					t.UserName = name
				}
			}
		})
		admins := userService.AdminIds(userIds)
		for _, t := range tokens {
			t.CreatedByAdmin = admins[t.UserId]
		}
	}
	return web.ToPage[*model.Token](int64(i), tokens), nil
}

func (l *TokenService) SendApiCallMail(schedule *model.Schedule) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	byToken, err := l.tokenModel.FindByPK(schedule.TokenId)
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	if !l.ownerEnabled(byToken) {
		return errors.New("token owner is disabled")
	}
	if byToken.IsUse() {
		l.supplementToken(byToken)
		var body string
		var err0 error
		if byToken.SMTP == nil {
			err0 = errors.New("SMTP not found")
		} else {
			body, err0 = smtp.SendAPIMail(schedule, byToken.SMTP, byToken.ReceiveEmails)
		}
		if err0 != nil {
			log.Error("SendAPIMail log error", zap.Error(err0))
		}
		// The scheduled call belongs to whoever owns the schedule, so that owner is who
		// sees it in their log list.
		err2 := l.logService.Log(schedule.UserId, byToken.SMTP, byToken.ReceiveEmails, nil, byToken.Name, byToken.Token, schedule.Name, body, err0)
		if err2 != nil {
			log.Error("SendAPIMail log error", zap.Error(err2))
		}
		return err0
	}
	return errors.New("token is not use")
}

func (l *TokenService) sendMailWithToken(byToken *model.Token, recipients []string, subject string, content string, files []*smtp.File) (any, error) {
	if !l.ownerEnabled(byToken) {
		return nil, errors.New("token owner is disabled")
	}
	l.supplementToken(byToken)
	if byToken.SMTP == nil {
		return nil, errors.New("SMTP not found")
	}
	byToken.ReceiveEmails = MergeRecipients(byToken.ReceiveEmails, recipients)
	if len(byToken.ReceiveEmails) == 0 {
		return nil, errors.New("no recipients: the token has none configured and the request carried none")
	}
	if subject == "" {
		subject = byToken.Subject
	}
	err2 := smtp.SendAllMsg(byToken.SMTP, byToken.ReceiveEmails, files, subject, content)
	err := l.logService.Log(byToken.UserId, byToken.SMTP, byToken.ReceiveEmails, files, byToken.Name, byToken.Token, subject, content, err2)
	if err != nil {
		log.Error("sendMailWithToken log error", zap.Error(err))
	}
	if err2 == nil {
		return web.Ok("ok"), nil
	}
	return "error", err2
}

// SendMailByToken serves both the public API (user == nil, token is the only credential)
// and the management endpoint (user set, so a token can only be used by its owner).
func (l *TokenService) SendMailByToken(req *web.Request, user *model.User) (any, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	var sendMailApi entity.SendMailApi
	files := make([]*smtp.File, 0)
	// Until the files are recorded in the mail log they belong to this request alone:
	// every error return after this point removes them, or a failed base64 field would
	// leave copies in the cache that nothing references and nothing ever deletes.
	// Registered before the close below so the handles are closed first (LIFO).
	logged := false
	defer func() {
		if !logged {
			for _, file := range files {
				if file != nil && file.FilePath != "" {
					_ = os.Remove(file.FilePath)
				}
			}
		}
	}()
	// go-mail reads these readers but never closes them, and the handles below are
	// *os.File: without this they stay open until the GC finalizer gets around to them,
	// which is how a busy instance runs out of file descriptors.
	defer func() {
		for _, file := range files {
			if file != nil && file.File != nil {
				_ = file.File.Close()
			}
		}
	}()
	if util.ContainsAnyIgnoreCase(req.ContentType(), "application/json") {
		if err := req.BindJSON(&sendMailApi); err != nil {
			return nil, err
		}
	} else {
		sendMailApi.Token = req.GetFormParam("token")
		sendMailApi.Content = req.GetFormParam("content")
		sendMailApi.Subject = req.GetFormParam("subject")
		sendMailApi.Recipients = util.SplitAndDeduplicate(req.GetFormParam("recipients"), ",")
	}
	byToken, err := l.tokenModel.GetOneByToken(sendMailApi.Token)
	if err != nil {
		return nil, err
	}
	if byToken == nil {
		return nil, errors.New("token not found")
	}
	// Report "not found" rather than "forbidden" so we don't leak token existence
	if user != nil && !user.IsAdmin && byToken.UserId != user.Id {
		return nil, errors.New("token not found")
	}
	if !byToken.IsUse() {
		return nil, errors.New("token is not use")
	}
	if req.IsMultipartForm() {
		form, err := req.MultipartForm()
		if err != nil {
			return nil, err
		}
		fileHeaders, ok := form.File["files"]
		if ok {
			for _, fileHeader := range fileHeaders {
				filePath := util.GetCachePath(l.cachePath, fileHeader.Filename)
				if err := web.SaveUploadedFile(fileHeader, filePath); err != nil {
					return nil, err
				}
				file, err := os.Open(filePath)
				if err != nil {
					return nil, err
				}
				files = append(files, &smtp.File{File: file, Name: fileHeader.Filename, FilePath: filePath})
			}
		}
	}
	files, err = l.appendBase64Files(files, sendMailApi.Files)
	if err != nil {
		return nil, err
	}
	// sendMailWithToken always writes the log entry — even for a failed send — and the
	// entry records these paths, so from here on the log's cleanup owns the files.
	logged = true
	return l.sendMailWithToken(byToken, sendMailApi.Recipients, sendMailApi.Subject, sendMailApi.Content, files)
}

// SendMailByTokenId sends through one of the caller's own tokens. An admin may use any
// token: the token list already shows admins every user's tokens together with the SMTP
// credentials behind them, so refusing the send would restrict nothing they cannot do by
// copying that SMTP configuration into a record of their own.
func (l *TokenService) SendMailByTokenId(user *model.User, req *entity.SendMailByTokenId) (any, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	byToken, err := l.tokenModel.FindByPK(req.TokenId)
	if err != nil {
		return nil, err
	}
	if byToken == nil || (byToken.UserId != user.Id && !user.IsAdmin) {
		return nil, errors.New("token not found")
	}
	if !byToken.IsUse() {
		return nil, errors.New("token is not use")
	}
	// Attachments are this request's until the log entry records them; a failure before
	// that point removes them again.
	files := make([]*smtp.File, 0)
	logged := false
	defer func() {
		if !logged {
			for _, file := range files {
				if file != nil && file.FilePath != "" {
					_ = os.Remove(file.FilePath)
				}
			}
		}
	}()
	defer func() {
		for _, file := range files {
			if file != nil && file.File != nil {
				_ = file.File.Close()
			}
		}
	}()
	files, err = l.appendBase64Files(files, req.Files)
	if err != nil {
		return nil, err
	}
	// sendMailWithToken always writes the log entry (even for a failed send), and the
	// entry records these paths, so from here on the log's cleanup owns the files.
	logged = true
	return l.sendMailWithToken(byToken, req.Recipients, req.Subject, req.Content, files)
}

// appendBase64Files writes the base64 attachments of a send request into the cache and
// appends them to files. On error the files written so far are still returned, so the
// caller's cleanup can remove them.
func (l *TokenService) appendBase64Files(files []*smtp.File, specs []entity.File) ([]*smtp.File, error) {
	for _, spec := range specs {
		if len(spec.Data) == 0 {
			continue
		}
		data, err := util.DecodeFileBase64(spec.Data)
		if err != nil {
			return files, err
		}
		name := spec.Name
		if len(name) == 0 {
			name, err = util.CalculateMD5(data)
			if err != nil {
				return files, err
			}
		}
		filePath := util.GetCachePath(l.cachePath, name)
		if err := util.WriteFile(data, filePath); err != nil {
			return files, err
		}
		vFile, err := os.Open(filePath)
		if err != nil {
			return files, err
		}
		files = append(files, &smtp.File{File: vFile, Name: name, FilePath: filePath})
	}
	return files, nil
}

// ownerEnabled reports whether the account a token belongs to may still send. A token is
// a credential of its own — the public API carries no session — so it outlives the account
// it was minted for: without this check, disabling or deleting a user would leave every
// token they hold sending mail. A lookup failure fails closed.
//
// user_id = 0 means the row predates token ownership (it was assigned by hand and is
// invisible to every list, which filters on user_id); there is no account to consult, so
// it is left alone rather than silently broken.
func (l *TokenService) ownerEnabled(byToken *model.Token) bool {
	if byToken.UserId == 0 {
		return true
	}
	userModel := core.GetModel[*model.UserModel](l.context)
	owner, err := userModel.FindByPK(byToken.UserId)
	if err != nil {
		log.Error("token owner lookup failed", zap.Uint("userId", byToken.UserId), zap.Error(err))
		return false
	}
	return owner != nil && owner.IsUse
}

func (l *TokenService) supplementToken(st ...*model.Token) {
	mailIds := make([]uint, 0)
	stmpIds := make([]uint, 0)
	for _, d := range st {
		mailIds = append(mailIds, util.StringToUintIds(d.ReceiveEmailIds)...)
		stmpIds = append(stmpIds, d.SMTPId)
	}
	mailMap, err := l.mailModel.FindMapByIds(mailIds)
	if err == nil {
		for _, d := range st {
			mailIds := util.StringToUintIds(d.ReceiveEmailIds)
			d.ReceiveEmails = GetMails(mailIds, mailMap)
			d.ReceiveEmailsStr = GetMailsStr(d.ReceiveEmails)
		}
	}
	idsMap, err := l.smtpModel.FindMapByIds(stmpIds)
	if err == nil {
		for _, d := range st {
			d.SMTP = idsMap[d.SMTPId]
			if d.SMTP != nil {
				d.SMTPStr = d.SMTP.Name
			}
		}
	}
}

func (l *TokenService) Init(context *core.Context) error {
	l.context = context
	l.cachePath = context.GetConfig().GetStringOrDefault("core.cachePath", "cache")
	l.lock = new(sync.RWMutex)
	l.mailModel = core.GetModel[*model.MailModel](l.context)
	l.smtpModel = core.GetModel[*model.SMTPModel](l.context)
	l.tokenModel = core.GetModel[*model.TokenModel](l.context)
	l.logService = core.GetService[*LogService](l.context)
	return nil
}

// MergeRecipients combines the recipients configured on the token with the ones the
// caller sent: the configured list is the default, the request adds to it, and an
// address present in both is sent to once. Addresses are compared case-insensitively,
// since the mailbox is the same either way.
func MergeRecipients(configured []*model.Mail, requested []string) []*model.Mail {
	merged := make([]*model.Mail, 0, len(configured)+len(requested))
	seen := make(map[string]bool, len(configured)+len(requested))
	for _, mail := range configured {
		if mail == nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(mail.Mail))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, mail)
	}
	for _, address := range requested {
		address = strings.TrimSpace(address)
		key := strings.ToLower(address)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, &model.Mail{Mail: address})
	}
	return merged
}

func GetMails(ids []uint, mailMap map[uint]*model.Mail) []*model.Mail {
	mails := make([]*model.Mail, 0)
	for _, id := range ids {
		v, ok := mailMap[id]
		if ok {
			mails = append(mails, v)
		}
	}
	return mails
}
func GetMailsStr(mails []*model.Mail) string {
	buffer := new(bytes.Buffer)
	for _, mail := range mails {
		buffer.WriteString("," + util.FormatMail(mail.Name, mail.Mail))
	}
	if buffer.Len() == 0 {
		return ""
	}
	return buffer.String()[1:]
}
