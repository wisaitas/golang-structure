package gormx

type Condition struct {
	Query interface{}
	Args  []interface{}
}

type Relation struct {
	Query string
	Args  []interface{}
}

type PaginationQuery struct {
	Page     *int    `query:"page"`
	PageSize *int    `query:"page_size"`
	Sort     *string `query:"sort"`
	Order    *string `query:"order"`
}
