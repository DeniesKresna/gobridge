package serror

type SError interface {
	// System add comment to error so it will be easy to tracking the error flow process in code
	AddComment(comment string)

	// Usually use this in usecase
	//
	// System add comment to error so it will be easy to tracking the error flow process in code
	//
	// Message is used for build message in response.
	AddCommentMessage(comment string, message string)

	// System will replace validation notes in error
	AddValidation(data interface{})

	// Usually use this in usecase
	//
	// System will add message in error.
	//
	// Message is used for build message in response.
	AddMessage(message string)
	GetMessage() string
	GetErrorMessage() string
	GetComment() string
	GetStatusCode() int
	GetError() error
	GetErrorLine() []string
	GetValidation() interface{}
}
