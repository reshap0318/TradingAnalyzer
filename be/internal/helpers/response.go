package helpers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/go-playground/validator/v10"
)

type TResponse struct {
	Code  int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type TResponseError struct {
	Code  int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

var (
	ErrNotFound = errors.New("data not found")
	ErrInternal = errors.New("internal server error")
)

func RespondWithMessage(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, TResponse{
		Code:    code,
		Message: message,
	})
}

func ResponsedWithData(c *gin.Context, code int, message string, data interface{}) {
	if message == "" {
		message = "success"
	}

	c.AbortWithStatusJSON(code, TResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func ResponseJsonNotValid(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, TResponseError{
		Code:    http.StatusBadRequest,
		Message: "Format JSON tidak valid",
	})
}

// func RespondValidation(c *gin.Context, err error) {
// 	if validationErrors, ok := err.(validator.ValidationErrors); ok {
// 		errorMessages := make(map[string][]string)
// 		for _, e := range validationErrors {
// 			field := e.Field()
// 			errorMessages[field] = append(errorMessages[field], e.Translate(TransValidate))
// 		}

// 		c.AbortWithStatusJSON(422, TResponseError{
// 			Code:  422,
// 			Error: errorMessages,
// 		})
// 		return
// 	}

// 	// Jika error lain (misal: JSON tidak valid)
// 	RespondWithMessage(c, 422, "Invalid request")
// }

func ReponseSuccess(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusOK, TResponse{
		Code:    http.StatusOK,
		Message: "success",
	})
}

func RespondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		RespondWithMessage(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInternal):
		RespondWithMessage(c, http.StatusInternalServerError, err.Error())
	default:
		RespondWithMessage(c, 400, err.Error())
	}
}
