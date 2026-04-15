package service

import (
	"bytes"
	"errors"
	"os"
	"sync"

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

func (l *TokenService) GetOne(id uint) (*model.Token, error) {
	tokenModel := core.GetModel[*model.TokenModel](l.context)
	token, err := tokenModel.FindById(id)
	if err != nil {
		return nil, err
	}
	return token, nil
}
func (l *TokenService) GetPage(page *web.Page) (any, error) {

	tokens, i, err := l.tokenModel.Page(page)
	if err != nil {
		return nil, err
	}
	l.supplementToken(tokens...)
	return web.ToPage[*model.Token](int64(i), tokens), nil
}

func (l *TokenService) SendApiCallMail(schedule *model.Schedule) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	byToken, err := l.tokenModel.Query().Where("token=?", schedule.Token).One()
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	if byToken.IsUse {
		body, err := smtp.SendAPIMail2(schedule, byToken.SMTP, byToken.ReceiveEmails)
		if err != nil {
			log.Error("SendAPIMail log error", zap.Error(err))
		}
		err2 := l.logService.Log(byToken.SMTP, byToken.ReceiveEmails, nil, schedule.Token, schedule.Name, body, err)
		if err2 != nil {
			log.Error("SendAPIMail log error", zap.Error(err2))
		}
		return err
	}
	return errors.New("token is not use")
}

func (l *TokenService) SendMailByToken(req *web.Request) (any, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	var sendMailApi entity.SendMailApi
	var byToken *model.Token
	files := make([]*smtp.File, 0)
	err2 := func() (err error) {
		if util.ContainsAnyIgnoreCase(req.ContentType(), "application/json") {
			err = req.BindJSON(&sendMailApi)
			if err != nil {
				log.Error("SendMailByToken log error", zap.Error(err))
				return err
			}
		} else {
			sendMailApi.Token = req.GetFormParam("token")
			sendMailApi.Content = req.GetFormParam("content")
			sendMailApi.Subject = req.GetFormParam("subject")
			sendMailApi.Recipients = util.SplitAndDeduplicate(req.GetFormParam("recipients"), ",")
		}
		byToken, err = l.tokenModel.GetOneByToken(sendMailApi.Token)
		if err != nil {
			log.Error("SendMailByToken log error", zap.Error(err))
			return err
		}
		if !byToken.IsUse {
			return errors.New("token is not use")
		}

		for _, mail := range sendMailApi.Recipients {
			byToken.ReceiveEmails = append(byToken.ReceiveEmails, &model.Mail{Mail: mail})
		}
		if len(sendMailApi.Subject) == 0 {
			sendMailApi.Subject = byToken.Subject
		}
		if req.IsMultipartForm() {
			form, err := req.MultipartForm()
			if err != nil {
				log.Error("SendMailByToken log error", zap.Error(err))
				return err
			}
			fileHeaders, ok := form.File["files"]
			if ok {
				for _, fileHeader := range fileHeaders {
					filePath := util.GetCachePath(l.cachePath, fileHeader.Filename)
					err := web.SaveUploadedFile(fileHeader, filePath)
					if err != nil {
						log.Error("SendMailByToken log error", zap.Error(err))
						return err
					}
					file, err := os.Open(filePath)
					if err != nil {
						log.Error("SendMailByToken log error", zap.Error(err))
						return err
					}
					files = append(files, &smtp.File{File: file, Name: fileHeader.Filename, FilePath: filePath})
				}
			}

		}

		if len(sendMailApi.Files) > 0 {
			for _, file := range sendMailApi.Files {
				if len(file.Data) == 0 {
					continue
				}
				fileData, err := util.DecodeFileBase64(file.Data)
				if err != nil {
					log.Error("SendMailByToken log error", zap.Error(err))
					return err
				}
				if len(file.Name) == 0 {
					file.Name, err = util.CalculateMD5(fileData)
					if err != nil {
						log.Error("SendMailByToken log error", zap.Error(err))
						return err
					}
				}
				filePath := util.GetCachePath(l.cachePath, file.Name)
				err = util.WriteFile(fileData, filePath)
				if err != nil {
					log.Error("SendMailByToken log error", zap.Error(err))
					return err
				}
				vFile, err := os.Open(filePath)
				if err != nil {
					log.Error("SendMailByToken log error", zap.Error(err))
					return err
				}
				files = append(files, &smtp.File{File: vFile, Name: file.Name, FilePath: filePath})
			}
		}
		return nil
	}()
	if err2 == nil {
		err2 = smtp.SendAllMsg2(byToken.SMTP, byToken.ReceiveEmails, files, sendMailApi.Subject, sendMailApi.Content)
	}
	err := l.logService.Log(byToken.SMTP, byToken.ReceiveEmails, files, sendMailApi.Token, sendMailApi.Subject, sendMailApi.Content, err2)
	if err != nil {
		log.Error("SendMailByToken log error", zap.Error(err))
	}
	if err2 == nil {
		return "ok", nil
	}
	return "error", err2

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
