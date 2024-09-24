package routes

import (
	"maribooru/internal/handlers"
)

func (av *VersionOne) Users() {
	handler := handlers.NewUserHandler(av.db, av.cfg, av.log)

	user := av.api.Group("/user")
	user.POST("/sign-in", handler.SignIn)
	user.POST("/sign-up", handler.SignUp)
	user.POST("/init-admin-create", handler.InitialCreateAdmin)
	user.POST("/change-password", handler.ChangePassword, av.mw.JWTMiddleware())
	user.GET("/", handler.SelfGet, av.mw.JWTMiddleware())
	user.PUT("/:id", handler.SelfUpdate, av.mw.JWTMiddleware())

	users := av.api.Group("/users")
	users.GET("/:id", handler.GetUserByID)
	users.GET("/", handler.GetAllUsers)
}
