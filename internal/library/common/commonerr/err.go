package commonerr

import (
	"encoding/json"
	"fmt"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

const (
	InternalServerName        = "internal_server_error"
	InternalServerDescription = "The server is unable to complete your request"
)

var (
	Err404 = SetNewError(http.StatusNotFound, "404", "404 not found")
)

// DefaultInputBody return bad request for bad body request
var DefaultInputBody = ErrorFormat{
	ErrorName:        "bad_request",
	ErrorDescription: "Your body request resulted in error",
}

type ErrorMessage struct {
	ErrorList  []*ErrorFormat `json:"error_list"`
	Code       int            `json:"code"`
	Translator ut.Translator  `json:"-"`
}

// Get error byte
func (errorMessage *ErrorMessage) Marshal() []byte {
	b, _ := json.Marshal(errorMessage)
	return b
}

// Get string
func (errorMessage *ErrorMessage) ToString() string {
	return string(errorMessage.Marshal())
}

// Error to implement error interface
func (errorMessage *ErrorMessage) Error() string {
	return errorMessage.ToString()
}

// Errorln is for print error
func (errorMessage *ErrorMessage) Errorln(listErr ...interface{}) *ErrorMessage {
	log.Logger.Error().Msg(fmt.Sprintln(listErr...))
	return errorMessage
}

type ErrorFormat struct {
	ErrorName        string `json:"error_name"`
	ErrorDescription string `json:"error_description"`
}

// SetErrorValidator containts setter error from github.com/go-playground/validator
func (errorMessage *ErrorMessage) SetErrorValidator(err error) *ErrorMessage {
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return errorMessage
		}

		for _, errorItem := range err.(validator.ValidationErrors) {
			translatedError := errorItem.Translate(errorMessage.Translator)
			if translatedError == "" {
				errorMessage.Append(errorItem.Field(), errorItem.Tag())
			} else {
				errorMessage.Append(errorItem.Field(), translatedError)
			}

		}

	}
	return errorMessage
}

// Create new error message
func NewErrorMessage() *ErrorMessage {
	return &ErrorMessage{}
}

// SetNewError is function return new error message.
// It support to set code, error name, and error description
func SetNewError(code int, errorName, errDesc string) *ErrorMessage {
	return &ErrorMessage{
		Code: code,
		ErrorList: []*ErrorFormat{
			{
				ErrorName:        errorName,
				ErrorDescription: errDesc,
			},
		},
	}
}

// SetNewBadRequest returns a new error message with the standard code for a bad request (400).
// It allows for setting both the error name and error description using the specified error format.
func SetNewBadRequestByFormat(ef *ErrorFormat) *ErrorMessage {
	return &ErrorMessage{
		Code: http.StatusBadRequest,
		ErrorList: []*ErrorFormat{
			ef,
		},
	}
}

// Set error new custom reuseable error message by validation tag
func (errorMessage *ErrorMessage) SetTranslator(translator ut.Translator) *ErrorMessage {
	errorMessage.Translator = translator
	return errorMessage
}

// SetNewBadRequest set error as bad request code 400 and custom message
func SetNewBadRequest(errorName, errDesc string) *ErrorMessage {
	return SetNewError(http.StatusBadRequest, errorName, errDesc)
}

// SetBadRequest set error as bad request code 400
func (errorMessage *ErrorMessage) SetBadRequest() *ErrorMessage {
	errorMessage.Code = http.StatusBadRequest
	return errorMessage
}

// Append is function add error to existing error message.
// It support to set error name and error description.
func (errorMessage *ErrorMessage) Append(errorName, errDesc string) *ErrorMessage {
	errorMessage.ErrorList = append(errorMessage.ErrorList, &ErrorFormat{
		ErrorName:        errorName,
		ErrorDescription: errDesc,
	})
	return errorMessage
}

// Set404 is a function that returns a new error message with the standard code for "not found" (404).
// It allows for specifying both the error name and error description.
func Set404() *ErrorMessage {
	return Err404
}

// SetDefaultErrBodyRequest generates an error response for a request with a default body.
func SetDefaultErrBodyRequest() *ErrorMessage {
	return SetNewBadRequestByFormat(&DefaultInputBody)
}

// SetNewInternalError is function return new error message with internal server error standard code(500).
func SetNewInternalError() *ErrorMessage {
	return SetNewError(http.StatusInternalServerError, InternalServerName, InternalServerDescription)
}

// SetNewUnauthorizedError is function return new error message with unauthorized error code(401).
// It support to set error name and error description
func SetNewUnauthorizedError(errorName, errDesc string) *ErrorMessage {
	return SetNewError(http.StatusUnauthorized, errorName, errDesc)
}

func SetNewTokenExpiredError() *ErrorMessage {
	return SetNewUnauthorizedError("unauthorized", "expired token")
}

func SetNewUnauthorizedAPICall() *ErrorMessage {
	return SetNewUnauthorizedError("api call is unauthorized", "api is unauthorized for user")
}
