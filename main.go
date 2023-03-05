package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	gologger "github.com/wawakakakyakya/GolangLogger"
	"github.com/wawakakakyakya/check_logs_by_mail/config"
	"github.com/wawakakakyakya/check_logs_by_mail/file"
	"github.com/wawakakakyakya/check_logs_by_mail/smtp"
)

var (
	wg     sync.WaitGroup
	mailWg sync.WaitGroup
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Println("[ERROR] read config failed")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	logConfig := config.GlobalConfig.LogConfig
	logger := gologger.NewLogger("main", logConfig.Path, logConfig.MaxSize, logConfig.MaxBackups, logConfig.MaxAge, logConfig.Compress, logConfig.Level)

	ctx := context.Background()
	mailCtx, mailCancelFunc := context.WithCancel(ctx)
	mailQueue := make(chan *smtp.SMTPData, 100)
	defer close(mailQueue)
	logger.Debug("add mail waitgroup")
	mailWg.Add(1)
	//config <> smtpの依存関係はここで切り離す
	go smtp.Main(config.GlobalConfig.SMTP.Host, config.GlobalConfig.SMTP.Port, config.GlobalConfig.SMTP.UserName, config.GlobalConfig.SMTP.PassWord, config.GlobalConfig.SMTP.Timeout, config.GlobalConfig.SMTP.From, mailQueue, mailCtx, logger, &mailWg)
	for _, f := range config.Files {
		wg.Add(1)
		logger.Debug("add waitgroup")
		go file.Main(f, logger, mailQueue, &wg)
	}

	wg.Wait()
	mailCancelFunc()
	logger.Debug("waiting mail process end...")

	mailWg.Wait()

	logger.Info("done.")
}
