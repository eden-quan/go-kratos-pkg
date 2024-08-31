package mongopkg

import "go.mongodb.org/mongo-driver/bson"

// PaginatorArgs 分页参数
type PaginatorArgs struct {
	// PageOption 分页
	PageOption PageOption
	// PageOrders 排序
	PageOrders bson.D
	// PageWheres 条件
	PageWheres []*Where
}

// PageOption .
type PageOption struct {
	Limit  int64
	Offset int64
}

// Order 排序
type Order struct {
	// Field 排序的字段(例子：uid)
	Field string
	// Order 排序的方向
	Order interface{}
}

// Where 条件
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
