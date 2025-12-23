package service

import (
	"errors"
	"os"
	"sync"

	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/smtp"
	"github.com/chuccp/http2smtp/util"
	"github.com/chuccp/http2smtp/web"
	"go.uber.org/zap"
)

type Token struct {
	db        *db.DB
	log       *Log
	zapLog    *zap.Logger
	cachePath string
}

func NewToken(db *db.DB, zapLog *zap.Logger, cachePath string) *Token {
	return &Token{db: db, log: NewLog(db), cachePath: cachePath, zapLog: zapLog}
}
func (token *Token) GetPage(page *web.Page) (any, error) {
	p, err := token.db.GetTokenModel().Page(page)
	if err != nil {
		return nil, err
	}
	token.supplementToken(p.List...)
	return p, nil
}

func (token *Token) supplementToken(st ...*db.Token) {
	mailIds := make([]uint, 0)
	stmpIds := make([]uint, 0)
	for _, d := range st {
		d.Name = d.Token
		mailIds = append(mailIds, util.StringToUintIds(d.ReceiveEmailIds)...)
		stmpIds = append(stmpIds, d.SMTPId)
	}
	mailMap, err := token.db.GetMailModel().GetMapByIds(mailIds)
	if err == nil {
		for _, d := range st {
			mailIds := util.StringToUintIds(d.ReceiveEmailIds)
			d.ReceiveEmails = db.GetMails(mailIds, mailMap)
			d.ReceiveEmailsStr = db.GetMailsStr(d.ReceiveEmails)
		}
	}
	idsMap, err := token.db.GetSMTPModel().GetMapByIds(stmpIds)
	if err == nil {
		for _, d := range st {
			d.SMTP = idsMap[d.SMTPId]
			if d.SMTP != nil {
				d.SMTPStr = d.SMTP.Name
			}
		}
	}
}

var lock = new(sync.Mutex)

func (token *Token) SendApiCallMail(schedule *db.Schedule) error {
	lock.Lock()
	defer lock.Unlock()
	byToken, err := token.GetOneByToken(schedule.Token)
	if err != nil {
		return err
	}
	if byToken == nil {
		return errors.New("token not found")
	}
	if byToken.IsUse {
		body, err := smtp.SendAPIMail(schedule, byToken.SMTP, byToken.ReceiveEmails)
		if err != nil {
			token.zapLog.Error("SendAPIMail log error", zap.Error(err))
		}
		err2 := token.log.Log(byToken.SMTP, byToken.ReceiveEmails, nil, schedule.Token, schedule.Name, body, err)
		if err2 != nil {
			token.zapLog.Error("SendAPIMail log error", zap.Error(err2))
		}
		return err
	} else {
		return errors.New("token is not use")
	}
}

func (token *Token) SendMailByToken(req *web.Request) (any, error) {
	lock.Lock()
	defer lock.Unlock()
	var sendMailApi entity.SendMailApi
	var byToken *db.Token
	files := make([]*smtp.File, 0)
	err2 := func() (err error) {
		if util.ContainsAnyIgnoreCase(req.GetContext().ContentType(), "application/json") {
			err = req.ShouldBindBodyWithJSON(&sendMailApi)
			if err != nil {
				token.zapLog.Error("SendMailByToken log error", zap.Error(err))
				return err
			}
		} else {
			sendMailApi.Token = req.FormValue("token")
			sendMailApi.Content = req.FormValue("content")
			sendMailApi.Subject = req.FormValue("subject")
			sendMailApi.Recipients = util.SplitAndDeduplicate(req.FormValue("recipients"), ",")
		}
		byToken, err = token.GetOneByToken(sendMailApi.Token)
		if err != nil {
			token.zapLog.Error("SendMailByToken log error", zap.Error(err))
			return err
		}
		if !byToken.IsUse {
			return errors.New("token is not use")
		}

		for _, mail := range sendMailApi.Recipients {
			byToken.ReceiveEmails = append(byToken.ReceiveEmails, &db.Mail{Mail: mail})
		}
		if len(sendMailApi.Subject) == 0 {
			sendMailApi.Subject = byToken.Subject
		}
		if req.IsMultipartForm() {
			form, err := req.MultipartForm()
			if err != nil {
				token.zapLog.Error("SendMailByToken log error", zap.Error(err))
				return err
			}
			fileHeaders, ok := form.File["files"]
			if ok {
				for _, fileHeader := range fileHeaders {
					filePath := util.GetCachePath(token.cachePath, fileHeader.Filename)
					err := web.SaveUploadedFile(fileHeader, filePath)
					if err != nil {
						token.zapLog.Error("SendMailByToken log error", zap.Error(err))
						return err
					}
					file, err := os.Open(filePath)
					if err != nil {
						token.zapLog.Error("SendMailByToken log error", zap.Error(err))
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
					token.zapLog.Error("SendMailByToken log error", zap.Error(err))
					return err
				}
				if len(file.Name) == 0 {
					file.Name, err = util.CalculateMD5(fileData)
					if err != nil {
						token.zapLog.Error("SendMailByToken log error", zap.Error(err))
						return err
					}
				}
				filePath := util.GetCachePath(token.cachePath, file.Name)
				err = util.WriteFile(fileData, filePath)
				if err != nil {
					token.zapLog.Error("SendMailByToken log error", zap.Error(err))
					return err
				}
				vFile, err := os.Open(filePath)
				if err != nil {
					token.zapLog.Error("SendMailByToken log error", zap.Error(err))
					return err
				}
				files = append(files, &smtp.File{File: vFile, Name: file.Name, FilePath: filePath})
			}
		}
		return nil
	}()
	if err2 == nil {
		err2 = smtp.SendAllMsg(byToken.SMTP, byToken.ReceiveEmails, files, sendMailApi.Subject, sendMailApi.Content)
	}
	err := token.log.Log(byToken.SMTP, byToken.ReceiveEmails, files, sendMailApi.Token, sendMailApi.Subject, sendMailApi.Content, err2)
	if err != nil {
		token.zapLog.Error("SendMailByToken log error", zap.Error(err))
	}
	if err2 == nil {
		return "ok", nil
	}
	return "error", err2

}

func (token *Token) GetOne(id int) (*db.Token, error) {
	one, err := token.db.GetTokenModel().GetOne(uint(id))
	if err != nil {
		return nil, err
	}
	token.supplementToken(one)
	return one, nil
}
func (token *Token) GetOneByToken(tokenStr string) (*db.Token, error) {
	byToken, err := token.db.GetTokenModel().GetOneByToken(tokenStr)
	if err != nil {
		return nil, err
	}
	token.supplementToken(byToken)
	return byToken, err
}
