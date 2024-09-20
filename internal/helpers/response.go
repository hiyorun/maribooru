package helpers

import (
	"maribooru/internal/structs"

	"github.com/labstack/echo/v4"
)

func Response(c echo.Context, status int, data interface{}, msg string) error {
	response := structs.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	}
	return c.JSON(status, response)
}

func ResponseWithSettings(c echo.Context, status int, data interface{}, msg string) error {
	response := structs.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	}
	return c.JSON(status, response)
}

func PageData(data interface{}, total, page, perPage int) structs.PagedData {
	paged := structs.PagedData{
		List: data,
		Meta: structs.Metadata{
			PerPage: perPage,
			Page:    page,
			Total:   total,
		},
	}
	return paged
}
