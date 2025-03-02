package errors

import "fmt"

// Error specifies an API that must be fullfilled by all error types.
type Error interface {

	// Error implements the error interface.
	Error() string

	// Msg returns the error message.
	Msg() string

	// Err returns wrapped error.
	Err() Error
}

var _ Error = (*customError)(nil)

// customError is a custom error struct representing a vIoT error.
type customError struct {
	msg string
	err Error
}

func (ce *customError) Error() string {
	if ce != nil {
		if ce.err != nil {
			return fmt.Sprintf("%s: %s", ce.msg, ce.err.Error())
		}
		return ce.msg
	}
	return ""
}

func (ce *customError) Msg() string {
	return ce.msg
}

func (ce *customError) Err() Error {
	return ce.err
}

// Contains inspects if Error's message is same as error
// in argument. If not it continues further unwrapping
// layers of Error until it founds it or unwrap all layers
func Contains(ce Error, e error) bool {
	if ce == nil || e == nil {
		return ce == nil
	}
	if ce.Msg() == e.Error() {
		return true
	}
	if ce.Err() == nil {
		return false
	}

	return Contains(ce.Err(), e)
}

// Wrap returns an Error that wraps the given error with the given message.
func Wrap(wrapper Error, err error) Error {
	if wrapper == nil || err == nil {
		return nil
	}
	return &customError{
		msg: wrapper.Msg(),
		err: cast(err),
	}
}

func cast(err error) Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok {
		return e
	}
	return &customError{
		msg: err.Error(),
		err: nil,
	}
}

// New returns an Error with the given message.
func New(msg string) Error {
	return &customError{
		msg: msg,
		err: nil,
	}
}