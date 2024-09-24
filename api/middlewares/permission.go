package middlewares

import (
	"maribooru/internal/helpers"
	"maribooru/internal/models"
	"maribooru/internal/structs"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (m *Middleware) PermissionMiddleware(requiredPermission structs.PermissionLevel) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m.log.Debug("PermissionMiddleware:Authenticating")
			userID, err := helpers.GetUserID(c, m.cfg.JWT.Secret)
			if err != nil {
				m.log.Error("Failed to get user from token", zap.Error(err))
				return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
			}

			mod := models.NewPermissionModel(m.db)
			userPermission, err := mod.GetByUserID(userID)
			if err != nil {
				m.log.Error("Failed to get user permission", zap.Error(err))
				return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
			}

			if userPermission.Permission&requiredPermission == 0 {
				return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
			}

			return next(c)
		}
	}
}

func (m *Middleware) AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m.log.Debug("AdminMiddleware:Authenticating")
			userID, err := helpers.GetUserID(c, m.cfg.JWT.Secret)
			if err != nil {
				m.log.Error("Failed to get user from token", zap.Error(err))
				return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
			}

			mod := models.NewPermissionModel(m.db)
			isAdmin, err := mod.IsAdmin(userID)
			if err != nil {
				return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
			}
			if !isAdmin {
				return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
			}

			return next(c)
		}
	}
}
