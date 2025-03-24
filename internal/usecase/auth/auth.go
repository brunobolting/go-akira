package auth

import (
	"akira/internal/entity"
	"context"
)

func Make(ctx context.Context, user entity.UserService, logger entity.Logger) entity.AuthService {
	captcha := MakeCaptcha()
	return NewService(ctx, user, captcha, logger)
}
