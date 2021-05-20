package ecode

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"
)

var (
	// ClientError is the error code representing client error
	ClientError int32 = 4000
	// ServerError is the error code representing server error
	ServerError int32 = 5000
)

// Server return the server side tars.Error
func Server(format string, args ...interface{}) *tars.Error {
	return &tars.Error{Code: ServerError, Message: fmt.Sprintf(format, args...)}
}

// Client return the client side tars.Error
func Client(format string, args ...interface{}) *tars.Error {
	return &tars.Error{Code: ClientError, Message: fmt.Sprintf(format, args...)}
}

// IsClientErrorCode returns true if the eCode represents the client error
func IsClientErrorCode(eCode int32) bool {
	return ClientError <= eCode && eCode < ServerError
}
