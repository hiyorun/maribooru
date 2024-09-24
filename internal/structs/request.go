package structs

type (
	PagedRequest struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}
)
