package tag

import (
	"errors"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	CategoryCreate struct {
		PhoneticCategory string `json:"phonetic_category" validate:"required"`
		ReadableCategory string `json:"readable_category"`
	}

	CategoryUpdate struct {
		ID uuid.UUID `json:"id" validate:"required"`
		CategoryCreate
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
		PhoneticCategory: c.PhoneticCategory,
		ReadableCategory: c.ReadableCategory,
	}
}

func NewCategoryHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		db,
		NewCategoryModel(db),
		cfg,
		log,
	}
}

func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	var request CategoryCreate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	data, err := h.model.Create(request.ToTable())
	if err != nil {
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to create tag category")
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (h *CategoryHandler) GetCategoryByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	data, err := h.model.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag category not found")
		}
		h.log.Error("Failed to get tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag category")
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (h *CategoryHandler) GetCategories(c echo.Context) error {
	params := helpers.GenericPagedQuery{
		Limit:    50,
		Offset:   0,
		Sort:     "name asc",
		Keywords: "",
	}

	if err := c.Bind(&params); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}

	data, count, err := h.model.GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag categories not found")
		}
		h.log.Error("Failed to get tag categories", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag categories")
	}

	paged := helpers.PageData(data, int(count), params.Offset, params.Limit)

	return helpers.Response(c, http.StatusOK, paged, "")
}

func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	var request CategoryUpdate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	data, err := h.model.Update(request.ToTable())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag category not found")
		}
		h.log.Error("Failed to update tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update tag category")
	}

	return helpers.Response(c, http.StatusOK, data, "")
}

func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	if err := h.model.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag category not found")
		}
		h.log.Error("Failed to delete tag category", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to delete tag category")
	}

	return helpers.Response(c, http.StatusOK, nil, "")
}
