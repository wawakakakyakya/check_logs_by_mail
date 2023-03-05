package smtp

import (
	"context"
	"sync"

	gologger "github.com/wawakakakyakya/GolangLogger"
)

func Main(host string, port int, userName string, password string, timeout int, from string, queue chan *SMTPData, cancelCtx context.Context, logger *gologger.Logger, wg *sync.WaitGroup) {

	defer wg.Done()
	smtpLogger := logger.Child("smtpMain")
	smtpLogger.Debug("start sendMail process")

	smtpClient := NewSMTPClient(host, port, userName, password, timeout, from, smtpLogger)
	for {
		smtpLogger.Info("waiting mailqueue...")
		select {
		case smtpData := <-queue:
			smtpLogger.Info("send mail")
			if err := smtpClient.Send(smtpData); err != nil {
				smtpLogger.Error("send mail failed")
				smtpLogger.Error(err.Error())
			}
			smtpLogger.Info("mail done")
		case <-cancelCtx.Done():
			smtpLogger.Info("sendmail process was ended")
			return
		}
	}
}
