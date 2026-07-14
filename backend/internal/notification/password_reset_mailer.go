package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type PasswordResetMailer struct {
	config SMTPConfig
}

func NewPasswordResetMailer(config SMTPConfig) PasswordResetMailer {
	return PasswordResetMailer{config: config}
}

func (mailer PasswordResetMailer) SendPasswordReset(_ context.Context, recipient, displayName, code string) error {
	address := mailer.config.Host + ":" + mailer.config.Port
	from, err := mail.ParseAddress(mailer.config.From)
	if err != nil {
		return err
	}
	var auth smtp.Auth
	if mailer.config.Username != "" {
		auth = smtp.PlainAuth("", mailer.config.Username, mailer.config.Password, mailer.config.Host)
	}
	message := strings.Join([]string{
		"From: " + mailer.config.From,
		"To: " + recipient,
		"Subject: " + mime.QEncoding.Encode("UTF-8", "RentNestHub 密码重置验证码"),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		passwordResetHTML(displayName, code),
	}, "\r\n")
	if mailer.config.Port == "465" {
		return sendMailImplicitTLS(address, mailer.config.Host, auth, from.Address, []string{recipient}, []byte(message))
	}
	return smtp.SendMail(address, auth, from.Address, []string{recipient}, []byte(message))
}

func sendMailImplicitTLS(address, host string, auth smtp.Auth, from string, recipients []string, message []byte) error {
	connection, err := tls.Dial("tcp", address, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(connection, host)
	if err != nil {
		connection.Close()
		return err
	}
	defer client.Quit()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		writer.Close()
		return err
	}
	return writer.Close()
}

func passwordResetHTML(displayName, code string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN"><body style="margin:0;background:#eef4ef;font-family:Arial,'Microsoft YaHei',sans-serif;color:#18221e">
  <table width="100%%" cellpadding="0" cellspacing="0" role="presentation"><tr><td align="center" style="padding:36px 16px">
    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:560px;background:#ffffff;border:1px solid #dce2dd;border-radius:8px">
      <tr><td style="padding:32px 36px 20px"><div style="display:inline-block;padding:9px 12px;background:#176b4b;border-radius:6px;color:#ffffff;font-weight:700">RentNestHub</div>
      <h1 style="margin:26px 0 12px;font-size:25px">重置你的密码</h1><p style="margin:0;color:#69746e;line-height:1.7">%s，你正在重置 RentNestHub 账户密码。</p></td></tr>
      <tr><td style="padding:0 36px"><div style="margin:8px 0 24px;padding:20px;text-align:center;background:#e9f4ee;border-radius:8px;letter-spacing:8px;font-size:30px;font-weight:700;color:#10533a">%s</div>
      <p style="margin:0;color:#69746e;line-height:1.7">验证码 10 分钟内有效，验证成功后将立即失效。请勿将验证码发送给任何人。</p></td></tr>
      <tr><td style="padding:26px 36px 32px;color:#69746e;font-size:12px;line-height:1.6">这是一封系统邮件，请勿直接回复。若非本人操作，请忽略此邮件。</td></tr>
    </table>
  </td></tr></table>
</body></html>`, displayName, code)
}
