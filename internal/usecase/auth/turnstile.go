package auth

import (
	"akira/internal/config/env"
	"akira/internal/entity"
	"context"

	"github.com/brunobolting/go-snowy"
)

var _ entity.CaptchaService = (*TurnstileCaptcha)(nil)

type TurnstileCaptcha struct {
	SecretKey string
	URL       string
}

func NewTurnstileCaptcha(secretKey string) *TurnstileCaptcha {
	return &TurnstileCaptcha{
		SecretKey: secretKey,
		URL:       "https://challenges.cloudflare.com/turnstile/v0/siteverify",
	}
}

func (t *TurnstileCaptcha) Verify(ctx context.Context, captcha string) error {
	if !env.ISPROD {
		return nil
	}
	remoteip, ok := ctx.Value(entity.REMOTEIP_NAME).(string)
	if !ok {
		remoteip = ""
	}
	res, err := snowy.Post[map[string]any](
		snowy.Config{},
		t.URL,
		snowy.Headers{},
		snowy.RequestData{
			FormData: map[string]string{
				"secret":  t.SecretKey,
				"response": captcha,
				"remoteip": remoteip,
			},
		},
	)
	if err != nil {
		return err
	}
	data := *res.Data
	if data["success"] != true {
		return entity.ErrInvalidCaptcha
	}
	return nil
}
