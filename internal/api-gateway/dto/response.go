package dto

// Response 统一响应结构
// @Description API 统一响应格式
type Response struct {
	Code    int         `json:"code" example:"0"`                    // 错误码,0表示成功
	Message string      `json:"message" example:"success"`           // 响应消息
	Data    interface{} `json:"data,omitempty" swaggertype:"string"` // 响应数据
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// HelloRequest 问候请求
// @Description 问候请求参数
type HelloRequest struct {
	// 可以添加请求参数,这里暂时为空
}

// HelloResponse 问候响应数据
type HelloResponse struct {
	Message string `json:"message" example:"Hello World"` // 问候消息
}
