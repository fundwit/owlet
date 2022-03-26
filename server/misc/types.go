package misc

type PagedBody struct {
	Items interface{} `json:"items"`
	Total uint64      `json:"total"`
}
