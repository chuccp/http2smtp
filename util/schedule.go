package util

import "github.com/robfig/cron/v3"

func ParserCron(cronStr string) error {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := parser.Parse(cronStr)
	if err != nil {
		return err
	}
	return nil
}
