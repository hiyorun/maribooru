package routes

import (
	"maribooru/internal/handlers"

	echojwt "github.com/labstack/echo-jwt/v4"
)

func (av *VersionOne) Users() {
	handler := handlers.NewUserHandler(av.db, av.cfg, av.log)

	auth := av.api.Group("/auth")
	auth.POST("/sign-in", handler.SignIn)
	auth.POST("/sign-up", handler.SignUp)
	auth.POST("/admin-create", handler.CreateAdmin)

	user := av.api.Group("/users", echojwt.WithConfig(av.cfg.JWT.Config))
	user.GET("/:id", handler.GetByID)
}
