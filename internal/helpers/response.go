package helpers

import (
	"math"

	"github.com/labstack/echo/v4"
)

type (
	JSONResponse struct {
		Status  int         `json:"status"`
		Data    interface{} `json:"data"`
		Message string      `json:"message"`
	}
	PagedData struct {
		List interface{} `json:"list"`
		Meta Metadata    `json:"meta"`
	}
	Metadata struct {
		PerPage int `json:"per_page"`
		Page    int `json:"page"`
		Total   int `json:"total"`
	}
)

func Response(c echo.Context, status int, data interface{}, msg string) error {
	response := JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	}
	return c.JSON(status, response)
}

func ResponseWithSettings(c echo.Context, status int, data interface{}, msg string) error {
	response := JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	}
	return c.JSON(status, response)
}

func PageData(data interface{}, total, offset, limit int) PagedData {
	page := int(math.Ceil((float64(offset) + 1) / float64(limit)))

	paged := PagedData{
		List: data,
		Meta: Metadata{
			PerPage: limit,
			Page:    page,
			Total:   total,
		},
	}
	return paged
}
