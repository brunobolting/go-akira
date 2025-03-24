package auth

import (
	"akira/internal/entity"
	"context"
)

var _ entity.CaptchaService = (*TurnstileCaptcha)(nil)

type TurnstileCaptcha struct {
	SecretKey string
}

func NewTurnstileCaptcha(secretKey string) *TurnstileCaptcha {
	return &TurnstileCaptcha{
		SecretKey: secretKey,
	}
}

func (t *TurnstileCaptcha) Verify(ctx context.Context, captcha string) error {
	return nil
}
