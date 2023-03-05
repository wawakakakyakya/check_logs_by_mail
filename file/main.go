package file

import (
	"sync"

	gologger "github.com/wawakakakyakya/GolangLogger"
	"github.com/wawakakakyakya/check_logs_by_mail/config"
	"github.com/wawakakakyakya/check_logs_by_mail/smtp"
)

func sendResult(fc *config.FileConfig, mailQueue chan *smtp.SMTPData) {
	for _, wc := range fc.Words {
		mailQueue <- wc.SMTPData
	}
}

// ファイル処理のエントリーポイント
func Main(fc *config.FileConfig, logger *gologger.Logger, mailQueue chan *smtp.SMTPData, wg *sync.WaitGroup) {

	defer wg.Done()

	fileParser, err := NewFile(fc, logger)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	sendResult(fc, mailQueue)
	logParser := NewLogParser(fc, logger)
	readSize, rotated, err := fileParser.Parse(fileParser.Config.FileName, logParser)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	var lastLine int64
	if rotated {
		lastLine = readSize
	} else {
		lastLine = fileParser.PosFile.LastLine + readSize
	}
	if err = fileParser.UpdatePosition(lastLine); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Debug("parse job ended successfully")
}
