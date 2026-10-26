package smtp

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

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

func SendTestMsg(st *model.SMTP) error {
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
func SendTestMsg2(st *model.SMTP) error {
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

func SendContentMsgByRecipients(smtp *model.SMTP, recipients []string, subject, bodyString string) error {
	receiveEmails := make([]*model.Mail, 0)
	for _, recipient := range recipients {
		name, m, err := util.ParseMail(recipient)
		if err != nil {
			continue
		}
		receiveEmails = append(receiveEmails, &model.Mail{Name: name, Mail: m})
	}
	return SendContentMsg(smtp, receiveEmails, subject, bodyString)
}

func SendContentMsg(smtp *model.SMTP, mails []*model.Mail, subject, bodyString string) error {
	return SendFilesMsg(smtp, mails, nil, subject, bodyString)
}

func SendContentTemplateMsg(smtp *model.SMTP, mails []*model.Mail, subject, bodyString string, useTemplate bool, templateStr string) (string, error) {
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
func SendAllMsg(smtp *model.SMTP, mails []*model.Mail, files []*File, subject, bodyString string) error {
	return SendFilesMsg(smtp, mails, files, subject, bodyString)
}
func SendFilesMsg(smtp *model.SMTP, mails []*model.Mail, files []*File, subject, bodyString string) error {
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

	// A recipient go-mail refuses must fail the request, not vanish: silently skipping
	// it meant the log listed an address that never received anything while the call
	// still reported success.
	for _, email := range sendMsg.ReceiveEmails {
		if util.EqualsAnyIgnoreCase(email.Mail, invalidMails...) {
			continue
		}
		if len(email.Name) > 0 {
			if err := msg.AddToFormat(email.Name, email.Mail); err != nil {
				return fmt.Errorf("invalid recipient %q: %w", email.Mail, err)
			}
		} else if err := msg.AddTo(email.Mail); err != nil {
			return fmt.Errorf("invalid recipient %q: %w", email.Mail, err)
		}
	}
	msg.Subject(sendMsg.Subject)
	msg.SetBodyString(mail.TypeTextPlain, sendMsg.BodyString)
	for _, f := range sendMsg.ff {
		if err := msg.AttachReader(f.Name, f.File, mail.WithFileName(f.Name)); err != nil {
			return fmt.Errorf("attachment %q: %w", f.Name, err)
		}
	}
	options := []mail.Option{
		mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		mail.WithPort(sendMsg.SendMail.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(sendMsg.SendMail.Username),
		mail.WithPassword(sendMsg.SendMail.Password),
	}
	switch sendMsg.SendMail.Port {
	case mail.DefaultPortSSL:
		// 465 is implicit TLS ("smtps"): the session is encrypted from the first byte.
		// Without this option the port is dialled in the clear and the handshake fails,
		// which is why the usual Gmail/QQ/163 SSL setting could not be used at all.
		options = append(options, mail.WithSSL())
	default:
		// go-mail defaults to TLSMandatory, which refuses every server that does not
		// advertise STARTTLS — plain relays included. Opportunistic still upgrades when
		// the server offers STARTTLS.
		options = append(options, mail.WithTLSPolicy(mail.TLSOpportunistic))
	}
	c, err := mail.NewClient(sendMsg.SendMail.Host, options...)
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

func SendAPIMail(schedule *model.Schedule, smtp *model.SMTP, mails []*model.Mail) (string, error) {
	url := schedule.Url
	Method := schedule.Method
	request := util.NewRequest()
	var headers []model.Header
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

func SendAllMsg2(smtp *model.SMTP, mails []*model.Mail, files []*File, subject, bodyString string) error {
	return SendFilesMsg2(smtp, mails, files, subject, bodyString)
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
	var headers []model.Header
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
		// The alert mail went out, but the call itself failed: that failure is what the
		// caller and the mail log must see. Returning nil here used to log a permanently
		// broken schedule as SUCCESS with the error text as its result.
		return err.Error(), err
	}
	if schedule.IsOnlySendByError {
		if response.StatusCode >= 400 {
			body, errSend := SendContentTemplateMsg2(smtp, mails, schedule.Name, string(response.Body), schedule.UseTemplate, schedule.Template)
			if errSend != nil {
				return body, errSend
			}
			// Same rule as the transport error above: the alert was delivered, the API
			// call still failed, and the log records the failure.
			return body, fmt.Errorf("api returned status %d", response.StatusCode)
		}
		return string(response.Body), nil
	}
	return SendContentTemplateMsg2(smtp, mails, schedule.Name, string(response.Body), schedule.UseTemplate, schedule.Template)
}
