package config

import (
	"regexp"

	"github.com/wawakakakyakya/check_logs_by_mail/smtp"
)

type FileConfig struct {
	FileName string        `yaml:"fileName"`
	PosFile  string        `yaml:"posFile"`
	MaxLine  int           `yaml:"maxLine"`
	Words    []*WordConfig `yaml:"words"`
}

type WordConfig struct {
	Word          string   `yaml:"word"` // regexp
	SkipThreshold int      `yaml:"skipThreshold"`
	Subject       string   `yaml:"subject"`
	Recipients    []string `yaml:"recipients"`
	Regexp        *regexp.Regexp
	SMTPData      *smtp.SMTPData
}
