package serror

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/DeniesKresna/gohelper/utlog"
)

type Serror struct {
	Comment      string      `json:"comment"`
	Message      string      `json:"-"`
	ErrorMessage string      `json:"error_message"`
	ErrorLine    []string    `json:"error_lines"`
	Error        error       `json:"-"`
	StatusCode   int         `json:"-"`
	Validation   interface{} `json:"validation"`
}

// Can use this wherever in code
//
// # Error is used for show error detail from system
//
// # Usually comment is used for showing which process end with error
//
// [Status Code List] https://restfulapi.net/http-status-codes/
func NewWithErrorComment(err error, statusCode int, comment string) SError {
	comment = strings.Trim(comment, " ")
	sysErr := &Serror{
		Comment:    comment,
		StatusCode: 500,
		ErrorLine:  getErrorFlow(),
	}

	if err != nil {
		sysErr.Error = err
		sysErr.ErrorMessage = err.Error()
		utlog.Error(err)
	}

	if statusCode != 0 {
		sysErr.StatusCode = statusCode
	}
	return sysErr
}

// Usually use this in usecase
//
// # Message is used for build message in response
//
// # Comment is used for showing which process end with error
//
// [Status Code List] https://restfulapi.net/http-status-codes/
func NewWithCommentMessage(statusCode int, comment string, message string) (sysErr *Serror) {
	comment = strings.Trim(comment, " ")
	message = strings.Trim(message, " ")
	sysErr = &Serror{
		Message:    message,
		Comment:    comment,
		StatusCode: 500,
		ErrorLine:  getErrorFlow(),
	}

	if sysErr.Error == nil {
		sysErr.Error = errors.New(message)
		sysErr.ErrorMessage = message
		utlog.Error(message)
	}

	if statusCode != 0 {
		sysErr.StatusCode = statusCode
	}
	return sysErr
}

// Usually use this in usecase
//
// # Error is used for show error detail from system
//
// # Message is used for build message in response
//
// # Comment is used for showing which process end with error
//
// [Status Code List] https://restfulapi.net/http-status-codes/
func NewWithErrorCommentMessage(err error, statusCode int, comment string, message string) (sysErr *Serror) {
	comment = strings.Trim(comment, " ")
	message = strings.Trim(message, " ")
	sysErr = &Serror{
		Message:    message,
		Comment:    comment,
		StatusCode: 500,
		ErrorLine:  getErrorFlow(),
	}

	if err != nil {
		sysErr.Error = err
		sysErr.ErrorMessage = err.Error()
		utlog.Error(err)
	}

	if statusCode != 0 {
		sysErr.StatusCode = statusCode
	}
	return sysErr
}

// Usually use this in usecase specifically in validation process
//
// # Message is used for build message in response
//
// # Comment is used for showing which process end with error
//
// [Status Code List] https://restfulapi.net/http-status-codes/
func NewWithCommentMessageValidation(statusCode int, comment string, message string, validation map[string]string) (sysErr *Serror) {
	comment = strings.Trim(comment, " ")
	message = strings.Trim(message, " ")
	sysErr = &Serror{
		Message:    message,
		Comment:    comment,
		Validation: validation,
		StatusCode: 500,
		ErrorLine:  getErrorFlow(),
	}

	if sysErr.Error == nil {
		sysErr.Error = errors.New(message)
		sysErr.ErrorMessage = message
		utlog.Error(message)
	}

	if statusCode != 0 {
		sysErr.StatusCode = statusCode
	}
	return sysErr
}

func (s *Serror) AddComment(str string) {
	errVal := s.Comment
	str = strings.Trim(str, " ")
	errVal = errVal + " " + str
	s.Comment = strings.Trim(errVal, " ")
}

func (s *Serror) AddCommentMessage(com string, mes string) {
	errVal := s.Comment
	com = strings.Trim(com, " ")
	errVal = errVal + " " + com
	s.Comment = strings.Trim(errVal, " ")

	s.Message = strings.Trim(mes, " ")
}

func (s *Serror) AddMessage(mes string) {
	s.Message = strings.Trim(mes, " ")
}

func (s *Serror) AddValidation(data interface{}) {
	s.Validation = data
}

func (s *Serror) GetComment() string {
	return s.Comment
}

func (s *Serror) GetMessage() string {
	return s.Message
}

func (s *Serror) GetErrorMessage() string {
	return s.ErrorMessage
}

func (s *Serror) GetValidation() interface{} {
	return s.Validation
}

func (s *Serror) GetStatusCode() int {
	if s.StatusCode == 0 {
		return 200
	}
	return s.StatusCode
}

func (s *Serror) GetError() error {
	return s.Error
}

func (s *Serror) GetErrorLine() []string {
	return s.ErrorLine
}

func traceError() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}
