package structs

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
