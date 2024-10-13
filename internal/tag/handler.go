package tag

import (
	"errors"
	"maribooru/internal/config"
	"maribooru/internal/helpers"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	TagCreate struct {
		Slug       string    `json:"slug" validate:"required"`
		Name       string    `json:"name"`
		CategoryID uuid.UUID `json:"category_id" validate:"required"`
	}

	TagParams struct {
		helpers.GenericPagedQuery
		CategoryID uuid.UUID `query:"category_id"`
	}

	TagUpdate struct {
		ID         uuid.UUID `json:"id" validate:"required"`
		Slug       string    `json:"slug"`
		Name       string    `json:"name"`
		CategoryID uuid.UUID `json:"category_id"`
	}

	TagResponse struct {
		ID           uuid.UUID `json:"id"`
		Slug         string    `json:"slug"`
		Name         string    `json:"name"`
		CategoryID   uuid.UUID `json:"category_id"`
		CategorySlug string    `json:"category_slug"`
		CategoryName string    `json:"category_name"`
	}

	TagHandler struct {
		db    *gorm.DB
		model *TagModel
		cfg   *config.Config
		log   *zap.Logger
	}
)

func (t *TagCreate) ToTable() Tag {
	return Tag{
		Slug:       strings.ToLower(helpers.RemoveSpaces(t.Slug)),
		Name:       t.Name,
		CategoryID: t.CategoryID,
	}
}

func (t *TagUpdate) ToTable() Tag {
	return Tag{
		ID:         t.ID,
		Slug:       strings.ToLower(helpers.RemoveSpaces(t.Slug)),
		Name:       t.Name,
		CategoryID: t.CategoryID,
	}
}

func (t *Tag) ToResponse() TagResponse {
	return TagResponse{
		ID:           t.ID,
		Slug:         t.Slug,
		Name:         t.Name,
		CategoryID:   t.CategoryID,
		CategorySlug: t.Category.Slug,
		CategoryName: t.Category.Name,
	}
}

func (t TagSlice) ToResponse() []TagResponse {
	data := make([]TagResponse, len(t))
	for i, v := range t {
		data[i] = v.ToResponse()
	}
	return data
}

func NewTagHandler(db *gorm.DB, cfg *config.Config, log *zap.Logger) *TagHandler {
	return &TagHandler{
		db:    db,
		model: NewTagModel(db),
		cfg:   cfg,
		log:   log,
	}
}

func (t *TagHandler) Create(c echo.Context) error {
	t.log.Debug("TagHandler: Create")
	userID, err := helpers.GetUserID(c, t.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request TagCreate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	tag := request.ToTable()
	tag.CreatedByID = userID

	data, err := t.model.Create(tag)
	if err != nil {
		t.log.Error("Failed to create tag", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to create tag")
	}

	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (t *TagHandler) GetByName(c echo.Context) error {
	t.log.Debug("TagHandler: GetByName")
	name := c.Param("name")
	data, err := t.model.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag not found")
		}
		t.log.Error("Failed to get tag", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (t *TagHandler) GetByID(c echo.Context) error {
	t.log.Debug("TagHandler: Get")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}

	data, err := t.model.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag not found")
		}
		t.log.Error("Failed to get tag", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (t *TagHandler) GetAll(c echo.Context) error {
	t.log.Debug("TagHandler: GetAll")
	params := TagParams{
		GenericPagedQuery: helpers.GenericPagedQuery{
			Limit:    50,
			Offset:   0,
			Sort:     "slug asc",
			Keywords: "",
		},
		CategoryID: uuid.Nil,
	}
	if err := c.Bind(&params); err != nil {
		t.log.Debug("params not set, using default values")
	}

	data, count, err := t.model.GetAll(params)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.log.Error("Failed to get tag", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to get tag")
	}
	paged := helpers.PageData(data.ToResponse(), int(count), params.Offset, params.Limit)
	return helpers.Response(c, http.StatusOK, paged, "")
}

func (t *TagHandler) Update(c echo.Context) error {
	t.log.Debug("TagHandler: Update")
	userID, err := helpers.GetUserID(c, t.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	var request TagUpdate
	if err := c.Bind(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "Invalid request")
	}
	if err := c.Validate(&request); err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, err.Error())
	}

	tag := request.ToTable()
	tag.UpdatedByID = userID

	data, err := t.model.Update(tag)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag not found")
		}
		t.log.Error("Failed to update tag", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to update tag")
	}
	return helpers.Response(c, http.StatusOK, data.ToResponse(), "")
}

func (t *TagHandler) Delete(c echo.Context) error {
	t.log.Debug("TagHandler: Delete")
	userID, err := helpers.GetUserID(c, t.cfg.JWT.Secret)
	if err != nil {
		return helpers.Response(c, http.StatusUnauthorized, nil, "Unauthorized")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return helpers.Response(c, http.StatusBadRequest, nil, "ID is needed")
	}
	err = t.model.Delete(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.Response(c, http.StatusNotFound, nil, "Tag not found")
		}
		t.log.Error("Failed to delete tag", zap.Error(err))
		return helpers.Response(c, http.StatusInternalServerError, nil, "Failed to delete tag")
	}
	return helpers.Response(c, http.StatusOK, nil, "")
}
