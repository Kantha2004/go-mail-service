package service

type MailService interface {
	SendEmail(to string, subject string, body string, messageId string) error
}
