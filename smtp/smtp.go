package smtp

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/chuccp/http2smtp/db"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/util"
	"github.com/wneessen/go-mail"
)

type Mail struct {
	Name string
	Mail string
}
type SMTP struct {
	Username string
	Mail     string
	Password string
	Host     string
	Port     int
}

type File struct {
	File     *os.File `json:"-"`
	Name     string   `json:"name"`
	FilePath string   `json:"filePath"`
}

type SendMsg struct {
	Subject       string
	BodyString    string
	ReceiveEmails []*Mail
	SendMail      *SMTP
	ff            []*File
}

func sendTestMsg(Smtp *SMTP) error {
	var sendMsg SendMsg
	sendMsg.SendMail = Smtp
	sendMsg.ReceiveEmails = []*Mail{{Name: Smtp.Username, Mail: Smtp.Mail}}
	sendMsg.Subject = "mail test"
	sendMsg.BodyString = " this is a test"
	return SendMail(&sendMsg)
}

func SendTestMsg(st *db.SMTP) error {
	if len(st.Username) == 0 {
		return errors.New("username cannot be empty")
	}
	if len(st.Password) == 0 {
		return errors.New("password cannot be empty")
	}
	if len(st.Host) == 0 {
		return errors.New("host cannot be empty")
	}
	return sendTestMsg(&SMTP{Username: st.Username, Mail: st.Mail, Password: st.Password, Host: st.Host, Port: st.Port})
}

func SendContentMsgByRecipients(smtp *db.SMTP, recipients []string, subject, bodyString string) error {
	receiveEmails := make([]*db.Mail, 0)
	for _, recipient := range recipients {
		name, m, err := util.ParseMail(recipient)
		if err != nil {
			continue
		}
		receiveEmails = append(receiveEmails, &db.Mail{Name: name, Mail: m})
	}
	return SendContentMsg(smtp, receiveEmails, subject, bodyString)
}

func SendContentMsg(smtp *db.SMTP, mails []*db.Mail, subject, bodyString string) error {
	return SendFilesMsg(smtp, mails, nil, subject, bodyString)
}

func SendContentTemplateMsg(smtp *db.SMTP, mails []*db.Mail, subject, bodyString string, useTemplate bool, templateStr string) (string, error) {
	if useTemplate {
		template, err := util.ParseTemplate(templateStr, bodyString)
		if err != nil {
			return bodyString, SendContentMsg(smtp, mails, subject, bodyString)
		} else {
			return template, SendContentMsg(smtp, mails, subject, template)
		}
	}
	return bodyString, SendContentMsg(smtp, mails, subject, bodyString)
}
func SendAllMsg(smtp *db.SMTP, mails []*db.Mail, files []*File, subject, bodyString string) error {
	return SendFilesMsg(smtp, mails, files, subject, bodyString)
}
func SendFilesMsg(smtp *db.SMTP, mails []*db.Mail, files []*File, subject, bodyString string) error {
	SMTP := &SMTP{Username: smtp.Username, Mail: smtp.Mail, Password: smtp.Password, Host: smtp.Host, Port: smtp.Port}
	receiveEmails := make([]*Mail, 0)
	for _, d := range mails {
		receiveEmails = append(receiveEmails, &Mail{Name: d.Name, Mail: d.Mail})
	}
	if len(receiveEmails) == 0 {
		return errors.New("receiveEmails is empty")
	}
	var sendMsg SendMsg
	sendMsg.SendMail = SMTP
	sendMsg.ReceiveEmails = receiveEmails
	sendMsg.Subject = subject
	sendMsg.BodyString = bodyString
	for _, file := range files {
		sendMsg.ff = append(sendMsg.ff, file)
	}
	return SendMail(&sendMsg)
}

func SendMail(sendMsg *SendMsg, invalidMails ...string) error {

	msg := mail.NewMsg()
	if len(sendMsg.SendMail.Username) > 0 {
		err := msg.FromFormat(sendMsg.SendMail.Username, sendMsg.SendMail.Mail)
		if err != nil {
			return err
		}
	} else {
		if err := msg.From(sendMsg.SendMail.Mail); err != nil {
			return err
		}
	}

	for _, email := range sendMsg.ReceiveEmails {
		if !util.EqualsAnyIgnoreCase(email.Mail, invalidMails...) {
			if len(email.Name) > 0 {
				err := msg.AddToFormat(email.Name, email.Mail)
				if err != nil {
					continue
				}
			} else {
				msg.AddTo(email.Mail)
			}
		}
	}
	msg.Subject(sendMsg.Subject)
	msg.SetBodyString(mail.TypeTextPlain, sendMsg.BodyString)
	for _, f := range sendMsg.ff {
		err := msg.AttachReader(f.Name, f.File, mail.WithFileName(f.Name))
		if err != nil {
			continue
		}
	}
	c, err := mail.NewClient(sendMsg.SendMail.Host, mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), mail.WithPort(sendMsg.SendMail.Port), mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(sendMsg.SendMail.Username), mail.WithPassword(sendMsg.SendMail.Password))
	if err != nil {
		return err
	}
	if err := c.DialAndSend(msg); err != nil {
		if strings.Contains(err.Error(), "not found") && len(invalidMails) == 0 {
			mails := util.ExtractEmails(err.Error())
			err := SendMail(sendMsg, mails...)
			if err != nil {
				return err
			} else {
				return ToUserNotFoundError(mails)
			}
		}
		return err
	}
	return nil
}

func SendAPIMail(schedule *db.Schedule, smtp *db.SMTP, mails []*db.Mail) (string, error) {
	url := schedule.Url
	Method := schedule.Method
	request := util.NewRequest()
	var headers []db.Header
	var dataMap = make(map[string]string)
	if len(schedule.HeaderStr) > 0 {
		err := json.Unmarshal([]byte(schedule.HeaderStr), &headers)
		if err == nil {
			for _, header := range headers {
				dataMap[header.Name] = header.Value
			}
		} else {
			return "", err
		}
	}
	if len(schedule.Headers) > 0 {
		for _, header := range schedule.Headers {
			dataMap[header.Name] = header.Value
		}
	}
	response, err := request.CallApiForResponse(url, dataMap, Method, []byte(schedule.Body))
	if err != nil {
		errSend := SendContentMsg(smtp, mails, schedule.Name, err.Error())
		if errSend != nil {
			return errSend.Error(), err
		}
		return err.Error(), nil
	}
	if schedule.IsOnlySendByError {
		if response.StatusCode >= 400 {
			return SendContentTemplateMsg(smtp, mails, schedule.Name, string(response.Body), schedule.UseTemplate, schedule.Template)
		}
		return string(response.Body), nil
	} else {
		return SendContentTemplateMsg(smtp, mails, schedule.Name, string(response.Body), schedule.UseTemplate, schedule.Template)
	}
}

func SendFilesMsg2(smtp *model.SMTP, mails []*model.Mail, files []*File, subject, bodyString string) error {
	SMTP := &SMTP{Username: smtp.Username, Mail: smtp.Mail, Password: smtp.Password, Host: smtp.Host, Port: smtp.Port}
	receiveEmails := make([]*Mail, 0)
	for _, d := range mails {
		receiveEmails = append(receiveEmails, &Mail{Name: d.Name, Mail: d.Mail})
	}
	if len(receiveEmails) == 0 {
		return errors.New("receiveEmails is empty")
	}
	var sendMsg SendMsg
	sendMsg.SendMail = SMTP
	sendMsg.ReceiveEmails = receiveEmails
	sendMsg.Subject = subject
	sendMsg.BodyString = bodyString
	for _, file := range files {
		sendMsg.ff = append(sendMsg.ff, file)
	}
	return SendMail(&sendMsg)
}
func SendContentTemplateMsg2(smtp *model.SMTP, mails []*model.Mail, subject, bodyString string, useTemplate bool, templateStr string) (string, error) {
	if useTemplate {
		template, err := util.ParseTemplate(templateStr, bodyString)
		if err != nil {
			return bodyString, SendContentMsg2(smtp, mails, subject, bodyString)
		} else {
			return template, SendContentMsg2(smtp, mails, subject, template)
		}
	}
	return bodyString, SendContentMsg2(smtp, mails, subject, bodyString)
}

func SendContentMsg2(smtp *model.SMTP, mails []*model.Mail, subject, bodyString string) error {
	return SendFilesMsg2(smtp, mails, nil, subject, bodyString)
}
func SendAPIMail2(schedule *model.Schedule, smtp *model.SMTP, mails []*model.Mail) (string, error) {
	url := schedule.Url
	Method := schedule.Method
	request := util.NewRequest()
	var headers []db.Header
	var dataMap = make(map[string]string)
	if len(schedule.HeaderStr) > 0 {
		err := json.Unmarshal([]byte(schedule.HeaderStr), &headers)
		if err == nil {
			for _, header := range headers {
				dataMap[header.Name] = header.Value
			}
		} else {
			return "", err
		}
	}
	if len(schedule.Headers) > 0 {
		for _, header := range schedule.Headers {
			dataMap[header.Name] = header.Value
		}
	}
	response, err := request.CallApiForResponse(url, dataMap, Method, []byte(schedule.Body))
	if err != nil {
		errSend := SendContentMsg2(smtp, mails, schedule.Name, err.Error())
		if errSend != nil {
			return errSend.Error(), err
		}
		return err.Error(), nil
	}
	if schedule.IsOnlySendByError {
		if response.StatusCode >= 400 {
			return SendContentTemplateMsg2(smtp, mails, schedule.Name, string(response.Body), schedule.UseTemplate, schedule.Template)
		}
		return string(response.Body), nil
	} else {
		return SendContentTemplateMsg2(smtp, mails, schedule.Name, string(response.Body), schedule.UseTemplate, schedule.Template)
	}
}
