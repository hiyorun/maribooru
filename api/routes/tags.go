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
}
