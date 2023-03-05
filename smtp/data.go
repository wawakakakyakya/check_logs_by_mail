package smtp

import (
	"strings"
)

type SMTPData struct {
	recipients []string
	subject    string
	queue      []string
}

const crlf = "\r\n"

func (s *SMTPData) Body() []byte {
	to := "To: " + strings.Join(s.recipients, ",")
	subject := "Subject: " + s.subject + crlf
	msg := strings.Join(s.queue, crlf)
	return []byte(strings.Join([]string{to, subject, msg}, crlf))
}

func (s *SMTPData) AddMsg(msg string) {
	s.queue = append(s.queue, msg)
}

func NewSMTPData(recipients []string, subject string) *SMTPData {
	return &SMTPData{recipients: recipients, subject: subject}
}
