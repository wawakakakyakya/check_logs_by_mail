package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	// "github.com/wawakakakyakya/check_logs_by_mail/localnet/smtp"

	logger "github.com/wawakakakyakya/GolangLogger"
)

var (
	CRLF string = "\r\n"
)

type SMTPClient struct {
	hostName string
	port     int
	userName string
	password string
	timeout  int
	from     string
	logger   *logger.Logger
	client   *smtp.Client
}

func NewSMTPClient(hostname string, port int, username string, password string, timeout int, from string, logger *logger.Logger) *SMTPClient {
	smtpLogger := logger.Child("smtpClient")
	c := &smtp.Client{}
	return &SMTPClient{hostName: hostname, port: port, userName: username, password: password, timeout: timeout, from: from, client: c, logger: smtpLogger}
}

func (s *SMTPClient) newClient(addr string, port int) error {
	var plainConn *net.Conn
	var tlsConn *tls.Conn
	var err error
	var client *smtp.Client

	server := fmt.Sprintf("%s:%d", addr, port)
	switch port {
	case 25:
		s.logger.Debug("plain connection start")
		plainConn, err = plainConnection(addr, port)
		if err != nil {
			s.logger.Error("make connection failed")
			return err
		}
		client, err = smtp.NewClient(*plainConn, server)
		if err != nil {
			s.logger.Error("make client failed")
			return err
		}
	case 465:
		s.logger.Debug("tls connection start")
		tlsConn, err = tlsConnection(addr, port)
		if err != nil {
			s.logger.Error("make connection failed")
			return err
		}
		client, err = smtp.NewClient(tlsConn, server)
		if err != nil {
			s.logger.Error("make client failed")
			return err
		}
	}
	s.client = client
	return nil
}

func (s *SMTPClient) addr() string {
	return fmt.Sprintf("%s:%d", s.hostName, s.port)
}

func (s *SMTPClient) getAuth() smtp.Auth {
	// s.logger.DebugF(s.userName, s.password, s.hostName)
	return smtp.PlainAuth("", s.userName, s.password, fmt.Sprintf("%s:%d", s.hostName, s.port))
}

func (s *SMTPClient) send(addr string, port int, from string, to []string, msg []byte) error {
	var err error
	s.newClient(addr, port)

	if err = s.hello(); err != nil {
		s.logger.Error("call hello failed")
		return err
	}

	if err = s.client.Auth(s.getAuth()); err != nil {
		s.logger.Error("make auth failed")
		return err
	}

	//set mail From
	if err = s.client.Mail(from); err != nil {
		s.logger.Error("set from failed")
		return err
	}

	//set mail To
	if err = s.setRecipients(to); err != nil {
		s.logger.Error("set recipients failed")
		return err
	}

	w, err := s.client.Data()
	if err != nil {
		s.logger.Error("set data failed")
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		s.logger.Error("write data failed")
		return err
	}
	defer w.Close()

	s.client.Quit()
	return nil
}

// https://zenn.dev/hsaki/books/golang-context/viewer/deadline
func (s *SMTPClient) Send(data *SMTPData) error {
	s.logger.DebugF("send mail to %s start", data.recipients)
	timeout, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
	defer cancel()
	// auth := smtp.CRAMMD5Auth(s.userName, s.password)
	mailRes := make(chan error)
	defer close(mailRes)
	s.logger.DebugF("body: %s", data.Body())
	go func() {
		mailRes <- s.send(s.hostName, s.port, s.from, data.recipients, data.Body())
	}()

	select {
	case err := <-mailRes:
		if err != nil {
			return err
		} else {
			s.logger.Debug("send mail was ended normally")
		}
	case <-timeout.Done():
		return fmt.Errorf("send mail failed by timeout(%dsec)", s.timeout)
	}
	return nil
}

// reuse need server setting
// w, err := s.client.Data()
// if err != nil {
// 	return err
// }
// defer w.Close()
// msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", strings.Join(data.Recipients, ","), data.Subject, "bodybody")
// if _, err := io.WriteString(w, msg); err != nil {
// 	return err
// }

// err = s.client.Quit()
// if err != nil {
// 	return err
// }
// return nil

// func (s *SMTPClient) Dial() error {

// 	if s.client != nil {
// 		return errors.New("smtp client is not nil")
// 	}

// 	client, err := smtp.Dial(s.addr())
// 	if err != nil {
// 		return err
// 	}
// 	s.client = client
// 	return nil
// }

// func (s *SMTPClient) isConnect() error {

// 	if s.client == nil {
// 		return errors.New("smtp client client is nil")
// 	} else if err := s.client.Noop(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *SMTPClient) Auth() error {
// 	err := s.isConnect()
// 	if err != nil {
// 		return err
// 	}
// 	return s.client.Auth(smtp.CRAMMD5Auth(s.userName, s.password))
// }

// send hello
// IPアドレスを[]で囲むとうまくいく
func (s *SMTPClient) hello() error {
	return s.client.Hello("[127.0.0.1]")
}

func (s *SMTPClient) setRecipients(recipents []string) error {
	for _, addr := range recipents {
		if err := s.client.Rcpt(addr); err != nil {
			return err
		}
	}
	return nil
}

// func (s *SMTPClient) setMailFrom(mailFrom string) error {
// 	if err := s.client.Mail(mailFrom); err != nil {
// 		return err
// 	}
// 	return nil
// }
