package auth

import (
	"akira/internal/config/env"
	"akira/internal/entity"
)

func MakeCaptcha() entity.CaptchaService {
	return NewTurnstileCaptcha(env.TURNSTILE_SECRET_KEY)
}
