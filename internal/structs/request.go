package structs

type (
	PagedRequest struct {
		Limit    int    `query:"limit"`
		Offset   int    `query:"offset"`
		Keywords string `query:"keywords"`
		Sort     string `query:"sort"`
	}
)
