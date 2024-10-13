package routes

import "maribooru/internal/tag"

func (av *VersionOne) Tags() {
	categoryHandler := tag.NewCategoryHandler(av.db, av.cfg, av.log)

	category := av.api.Group("/tag-categories", av.mw.JWTMiddleware(), av.mw.AdminMiddleware())
	category.POST("", categoryHandler.CreateCategory)
	category.PUT("", categoryHandler.UpdateCategory)
	category.DELETE("/:id", categoryHandler.DeleteCategory)

	publicCategory := av.api.Group("/tag-categories")
	publicCategory.GET("", categoryHandler.GetCategories)
	publicCategory.GET("/:id", categoryHandler.GetCategoryByID)

	tagHandler := tag.NewTagHandler(av.db, av.cfg, av.log)

	tag := av.api.Group("/tags", av.mw.JWTMiddleware())
	tag.POST("", tagHandler.Create)
	tag.PUT("", tagHandler.Update)
	tag.DELETE("/:id", tagHandler.Delete)

	publicTag := av.api.Group("/tags")
	publicTag.GET("", tagHandler.GetAll)
	publicTag.GET("/:id", tagHandler.GetByID)
	publicTag.GET("/name/:name", tagHandler.GetByName)
}
