package helpers

type (
	GenericPagedQuery struct {
		Limit    int    `query:"limit"`
		Offset   int    `query:"offset"`
		Keywords string `query:"keywords"`
		Sort     string `query:"sort"`
	}
)
