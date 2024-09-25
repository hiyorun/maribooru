package helpers

import (
	"maribooru/internal/structs"
	"math"

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

func PageData(data interface{}, total, offset, limit int) structs.PagedData {
	page := int(math.Ceil((float64(offset) + 1) / float64(limit)))

	paged := structs.PagedData{
		List: data,
		Meta: structs.Metadata{
			PerPage: limit,
			Page:    page,
			Total:   total,
		},
	}
	return paged
}
