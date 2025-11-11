package sms

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

type Sender interface {
	Send(ctx context.Context, phone, message string) error
}

type LoggerSender struct {
	log zerolog.Logger
}

func NewLoggerSender(log zerolog.Logger) *LoggerSender {
	return &LoggerSender{log: log}
}

func (s *LoggerSender) Send(_ context.Context, phone, message string) error {
	s.log.Info().
		Str("phone", phone).
		Str("message", message).
		Msg("sms sent")
	return nil
}

func FormatAuthCode(code string) string {
	return fmt.Sprintf("Ваш код для входа: %s", code)
}
