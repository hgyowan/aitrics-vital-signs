package error

import (
	"errors"
)

type BusinessError struct {
	error
	Status *Status
}

func (e *BusinessError) Error() string {
	if e == nil || e.error == nil {
		return ""
	}
	return e.error.Error()
}

func (e *BusinessError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.error
}

var errEmptyBusiness = errors.New("empty business error")

func EmptyBusinessError() error {
	return errEmptyBusiness
}

func newBusinessError(err error, status Status) error {
	return &BusinessError{
		error:  wrapBusiness(err, status.Message),
		Status: &status,
	}
}

func Wrap(err error) error {
	return wrap(err)
}

// WrapWithCode
// 새로운 code 의 business error 를 생성하고자 하는 경우 첫번째 인자에 EmptyBusinessError() 를 넣어주세요.
func WrapWithCode(err error, code Code, details ...string) error {
	if err == nil {
		return nil
	}

	base := businessCodeMap[None]
	if s, ok := businessCodeMap[code]; ok {
		base = s
	}

	if errors.Is(err, errEmptyBusiness) {
		err = errors.New(base.Message)
	}

	if len(details) > 0 {
		base.Detail = details
	}

	return newBusinessError(err, base)
}

func CastBusinessError(err error) (*BusinessError, bool) {
	for err != nil {
		var be *BusinessError
		if errors.As(err, &be) {
			return be, true
		}
		err = errors.Unwrap(err)
	}
	return nil, false
}

func CompareBusinessError(err error, code Code) bool {
	if t, ok := CastBusinessError(err); ok {
		if t.Status.Code == int(code) {
			return true
		}
	}

	return false
}
