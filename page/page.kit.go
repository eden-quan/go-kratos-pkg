package pagepkg

import (
	"regexp"
	"strings"
)

const (
	DefaultPageNumber = 1  // goto page number : which page (default : 1)
	DefaultPageSize   = 20 // show records number (default : 20)

	DefaultOrderColumn = "id"   // default order column
	DefaultOrderAsc    = "ASC"  // order direction : asc
	DefaultOrderDesc   = "DESC" // order direction : desc

	DefaultPlaceholder = "?" // param placeholder
)

var (
	// regColumn 正则表达式:列
	regColumn = regexp.MustCompile("^[A-Za-z-_]+$")
)

// DefaultPageRequest 默认分页请求
func DefaultPageRequest() *PageRequest {
	return &PageRequest{
		Page:     DefaultPageNumber,
		PageSize: DefaultPageSize,
	}
}

// PageOption .
type PageOption struct {
	Limit  int
	Offset int
}

// ConvertToPageOption 转换为分页选项
func ConvertToPageOption(pageRequest *PageRequest) *PageOption {
	opts := &PageOption{
		Limit:  int(pageRequest.PageSize),
		Offset: int(pageRequest.PageSize * (pageRequest.Page - 1)),
	}
	return opts
}

// PaginatorArgs 列表参数
type PaginatorArgs struct {
	// PageOption 分页
	PageOption *PageOption
	// PageOrders 排序
	PageOrders []*Order
	// PageWheres 条件
	PageWheres []*Where
}

// Order 排序(例子：order by id desc)
type Order struct {
	// Field 排序的字段(例子：id)
	Field string
	// Order 排序的方向(例子：desc)
	Order string
}

// NewOrder order
func NewOrder(field, orderDirection string) *Order {
	return &Order{
		Field: field,
		Order: orderDirection,
	}
}

// AssembleSQL 组装排序
func (o *Order) AssembleSQL() string {
	if o.Field == "" {
		return ""
	}

	column := o.Field
	if !IsValidField(column) {
		//column = DefaultOrderColumn
		column = "bad_order_from_invalid_column"
	}
	return column + " " + ParseOrderDirection(o.Order)
}

// AssembleUnsafeSQL 不安全的组装排序
func (o *Order) AssembleUnsafeSQL() string {
	if o.Field == "" {
		return ""
	}
	return o.Field + " " + o.Order
}

// IsValidField 判断是否为有效的字段名
func IsValidField(field string) bool {
	return regColumn.MatchString(field)
}

// ParseOrderDirection 排序方向
func ParseOrderDirection(orderDirection string) string {
	if orderDirection = strings.ToUpper(orderDirection); orderDirection == DefaultOrderAsc {
		return DefaultOrderAsc
	}
	return DefaultOrderDesc
}

// Where 条件；例：where id = ?(where id = 1)
type Where struct {
	// Field 字段
	Field string
	// Operator 运算符
	Operator string
	// Placeholder 占位符
	Placeholder string
	// Value 数据
	Value interface{}
}

// NewWhere where
func NewWhere(field, operator string, value interface{}) *Where {
	return &Where{
		Field:       field,
		Operator:    operator,
		Placeholder: DefaultPlaceholder,
		Value:       value,
	}
}

// NewWhereWithPlaceholder where
func NewWhereWithPlaceholder(field, operator, placeholder string, value interface{}) *Where {
	return &Where{
		Field:       field,
		Operator:    operator,
		Placeholder: placeholder,
		Value:       value,
	}
}

func (w *Where) AssembleSQL() string {
	if w.Field == "" {
		return ""
	}
	column := w.Field
	if !IsValidField(column) {
		//column = DefaultOrderColumn
		column = "bad_where_from_invalid_column"
	}
	return column + " " + w.Operator + " " + w.Placeholder
}

func (w *Where) AssembleUnsafeSQL() string {
	if w.Field == "" {
		return ""
	}
	return w.Field + " " + w.Operator + " " + w.Placeholder
}
