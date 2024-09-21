package middlewares

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func (m *Middleware) JWTMiddleware() echo.MiddlewareFunc {
	m.log.Debug("JWTMiddleware:Authenticating")
	return echojwt.WithConfig(m.cfg.JWT.Config)
}
