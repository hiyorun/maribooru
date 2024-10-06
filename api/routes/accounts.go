package routes

import (
	"maribooru/internal/account"
	"maribooru/internal/permission"
)

func (av *VersionOne) Accounts() {
	userHandler := account.NewUserHandler(av.db, av.cfg, av.log)
	adminHandler := account.NewAdminHandler(av.db, av.cfg, av.log)
	permissionHandler := permission.NewHandler(av.db, av.cfg, av.log)

	user := av.api.Group("/user")
	user.POST("/sign-in", userHandler.SignIn)
	user.POST("/sign-up", userHandler.SignUp)
	user.POST("/init-admin-create", adminHandler.InitialCreateAdmin)
	user.PUT("/change-password", userHandler.ChangePassword, av.mw.JWTMiddleware())
	user.GET("", userHandler.SelfGet, av.mw.JWTMiddleware())
	user.PUT("", userHandler.SelfUpdate, av.mw.JWTMiddleware())
	user.DELETE("", userHandler.SelfDelete, av.mw.JWTMiddleware())

	users := av.api.Group("/users")
	users.GET("/:id", userHandler.GetUserByID)
	users.GET("", userHandler.GetAllUsers)

	admin := av.api.Group("/admin", av.mw.JWTMiddleware(), av.mw.AdminMiddleware())
	adminManage := admin.Group("/manage")
	adminManage.POST("", adminHandler.CreateAdmin)
	adminManage.GET("", adminHandler.GetAllAdmin)
	adminManage.PUT("/:id", adminHandler.AssignAdmin)
	adminManage.DELETE("/:id", adminHandler.RemoveAdmin)

	adminUser := admin.Group("/user")
	adminUser.PUT("/:id", adminHandler.AdministrativeUserUpdate)
	adminUser.GET("/permission/:id", permissionHandler.GetByUserID)
	adminUser.PUT("/permission", permissionHandler.Set)
}
