package serializer

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 基础序列化器
type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

// ResponseUrl 响应URL链接
type ResponseUrl struct {
	Url string `json:"url"`
}

const (
	// CodeSuccess 成功状态码 200
	CodeSuccess = http.StatusOK
	// CodeNotLogin 未登录状态码 1250
	CodeNotLogin = 1250
	// CodeNotAuthError 未授权状态码 401
	CodeNotAuthError = http.StatusUnauthorized
	// CodeDBError 数据库错误状态码 500
	CodeDBError = http.StatusInternalServerError
	// CodeInternalError 内部错误状态码 500
	CodeInternalError = http.StatusInternalServerError
	// CodeError 通用错误状态码 404
	CodeError = http.StatusNotFound
	// CodeParamsError 参数错误状态码 50001
	CodeParamsError = 50001
)

// Success 返回成功响应
func Success(data interface{}) Response {
	return Response{
		Code: CodeSuccess,
		Msg:  "Success",
		Data: data,
	}
}

// NotAuthErr 使用消息构建未授权错误响应，如果消息为空
// 则默认消息为 "NotAuth"
func NotAuthErr(msg string) Response {
	if msg == "" {
		msg = "NotAuth"
	}
	return Response{
		Code: CodeNotAuthError,
		Msg:  msg,
	}
}

// NotLogin 返回未登录响应
func NotLogin(msg string) Response {
	if msg == "" {
		msg = "NotLogin"
	}
	return Response{
		Code: CodeNotLogin,
		Msg:  msg,
	}
}

// Err 返回通用错误响应
func Err(errCode int, msg string, err error) Response {
	res := Response{
		Code: errCode,
		Msg:  msg,
	}
	if err != nil && gin.Mode() != gin.ReleaseMode {
		res.Error = fmt.Sprintf("%+v", err)
	}
	return res
}

// DBErr 返回数据库错误响应
func DBErr(msg string, err error) Response {
	if msg == "" {
		msg = "DBerr"
	}
	return Err(CodeDBError, msg, err)
}

// InternalErr 返回内部错误响应
func InternalErr(msg string, err error) Response {
	if msg == "" {
		msg = "Internal"
	}
	return Err(CodeInternalError, msg, err)
}

// ParamsErr 返回参数错误响应
func ParamsErr(msg string, err error) Response {
	if msg == "" {
		msg = "ParamErr"
	}
	return Err(CodeParamsError, msg, err)
}

// ErrorResponse 返回错误消息
func ErrorResponse(err error) Response {
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return ParamsErr("JsonNotMatched", err)
	}

	return ParamsErr("", err)
}
