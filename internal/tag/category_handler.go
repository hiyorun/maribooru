package tag

import (
	"errors"
	"maribooru/internal/account"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	CategoryCreate struct {
		Slug string `json:"slug" validate:"required"`
		Name string `json:"name"`
	}

	CategoryUpdate struct {
		ID   uuid.UUID `json:"id" validate:"required"`
		Slug string    `json:"slug"`
		Name string    `json:"name"`
	}

	CategoryResponse struct {
		ID        uuid.UUID            `json:"id"`
		Slug      string               `json:"slug"`
		Name      string               `json:"name"`
		CreatedAt time.Time            `json:"created_at"`
		UpdatedAt time.Time            `json:"updated_at"`
		DeletedAt time.Time            `json:"deleted_at"`
		CreatedBy account.UserResponse `json:"created_by"`
		UpdatedBy account.UserResponse `json:"updated_by"`
		DeletedBy account.UserResponse `json:"deleted_by"`
	}

	CategoryHandler struct {
		db    *gorm.DB
		model *CategoryModel
		cfg   *config.Config
		log   *zap.Logger
	}
)

func (c *CategoryCreate) ToTable() TagCategory {
	return TagCategory{
		Slug: c.Slug,
		Name: c.Name,
	}
}

func (c *CategoryUpdate) ToTable() TagCategory {
	return TagCategory{
		ID:   c.ID,
		Slug: c.Slug,
		Name: c.Name,
	}
}

func (t *TagCategory) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: t.DeletedAt.Time,
		CreatedBy: t.CreatedBy.ToResponse(false),
		UpdatedBy: t.UpdatedBy.ToResponse(false),
		DeletedBy: t.DeletedBy.ToResponse(false),
	}
}

func (t TagCategorySlice) ToResponse() []CategoryResponse {
	response := make([]CategoryResponse, 0)
	for _, category := range t {
		response = append(response, category.ToResponse())
	}
	return response
}

func NewCategoryHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		db,
		NewCategoryModel(db),
		cfg,
		log,
	}
}

func (ch *CategoryHandler) CreateCategory(c echo.Context) error {
	ch.log.Debug("CategoryHandler: Create")
	userID, err := helpers.GetUserID(c, ch.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request CategoryCreate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	category := request.ToTable()
	category.CreatedByID = userID

	data, err := ch.model.Create(category)
	if err != nil {
		ch.log.Error("Failed to create tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to create tag category")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (ch *CategoryHandler) GetCategoryByID(c echo.Context) error {
	ch.log.Debug("CategoryHandler: Get")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	data, err := ch.model.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag category not found")
		}
		ch.log.Error("Failed to get tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag category")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (ch *CategoryHandler) GetCategories(c echo.Context) error {
	ch.log.Debug("CategoryHandler: GetAll")
	params := helpers.GenericPagedQuery{
		Limit:    50,
		Offset:   0,
		Sort:     "slug asc",
		Keywords: "",
	}

	if err := c.Bind(&params); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	data, count, err := ch.model.GetAll(params)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ch.log.Error("Failed to get tag categories", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag categories")
	}

	paged := helpers.PageData(data.ToResponse(), int(count), params.Offset, params.Limit)

	return helpers.Response(c, http.StatusOK, paged, "")
}

func (ch *CategoryHandler) UpdateCategory(c echo.Context) error {
	ch.log.Debug("CategoryHandler: Update")
	userID, err := helpers.GetUserID(c, ch.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request CategoryUpdate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	category := request.ToTable()
	category.UpdatedByID = userID

	data, err := ch.model.Update(category)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag category not found")
		}
		ch.log.Error("Failed to update tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update tag category")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (ch *CategoryHandler) DeleteCategory(c echo.Context) error {
	ch.log.Debug("CategoryHandler: Delete")
	userID, err := helpers.GetUserID(c, ch.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	if err := ch.model.Delete(id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag category not found")
		}
		ch.log.Error("Failed to delete tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to delete tag category")
	}

	return helpers.Response(c, http.StatusOK, nil, "")
}
