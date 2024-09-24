package routes

import (
	"maribooru/internal/handlers"
)

func (av *VersionOne) Users() {
	handler := handlers.NewUserHandler(av.db, av.cfg, av.log)

	auth := av.api.Group("/auth")
	auth.POST("/sign-in", handler.SignIn)
	auth.POST("/sign-up", handler.SignUp)
	auth.POST("/init-admin-create", handler.InitialCreateAdmin)
	auth.POST("/change-password", handler.ChangePassword, av.mw.JWTMiddleware())
	auth.GET("/", handler.GetSelf, av.mw.JWTMiddleware())
	auth.PUT("/:id", handler.UserUpdate, av.mw.JWTMiddleware())

	user := av.api.Group("/users")
	user.GET("/:id", handler.GetUserByID)
}
