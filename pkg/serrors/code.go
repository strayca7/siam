package serrors

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var (
	unknownCoder Code = Code{
		1,
		http.StatusInternalServerError,
		"An internal server error occurred",
		"http://github.com/strayca7/pkg/serrors/README.md",
	}
)

// Coder defines an interface for an error code detail information.
type Coder interface {
	// Code returns the code of the coder
	Code() int

	// HTTP status that should be used for the associated error code.
	HTTPStatus() int

	// External (user) facing error text.
	External() string

	// Reference returns the detail documents for user.
	Reference() string
}

// Code is a basic struct that implements the Coder interface.
// You can use it to Register(coder Coder) or MustRegister(coder Coder) your own error codes.
type Code struct {
	// C refers to the integer code of the ErrCode.
	C int

	// HTTP status that should be used for the associated error code.
	HTTP int

	// External (user) facing error text.
	Ext string

	// Ref specify the reference document.
	Ref string
}

// HTTPStatus returns the associated HTTP status code, if any. Otherwise,
// returns 200.
func (coder Code) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}

	return coder.HTTP
}

// Code returns the integer code of the coder.
func (coder Code) Code() int {
	return coder.C
}

// String implements stringer. String returns the external error message,
// if any.
func (coder Code) External() string {
	return coder.Ext
}

// Reference returns the reference document.
func (coder Code) Reference() string {
	return coder.Ref
}

// codes contains a map of error codes to metadata.
var codes = map[int]Coder{}
var mu sync.Mutex

// Register register a user define error code.
// It will overrid the exist code.
func Register(coder Coder) {
	if coder.Code() == 0 {
		panic("code `0` is reserved by `github.com/strayca7/siam/pkg/serrors` as unknownCode error code")
	}

	mu.Lock()
	defer mu.Unlock()

	codes[coder.Code()] = coder
}

// MustRegister register a user define error code.
// It will panic when the same Code already exist.
func MustRegister(coder Coder) {
	if coder.Code() == 0 {
		panic("code '0' is reserved by 'github.com/strayca7/siam/pkg/serrors' as ErrUnknown error code")
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := codes[coder.Code()]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code()))
	}

	codes[coder.Code()] = coder
}

// ParseCoder parse any error into *withCode.
// nil error will return nil direct.
// None withStack error will be parsed as ErrUnknown.
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}

	var wc *withCode
	if errors.As(err, &wc) {
		if coder, ok := codes[wc.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// IsCode reports whether any error in err's chain contains the given error code.
func IsCode(err error, code int) bool {
	var wc *withCode
	if errors.As(err, &wc) {
		if wc.code == code {
			return true
		}

		if wc.cause != nil {
			return IsCode(wc.cause, code)
		}

		return false
	}

	return false
}

func init() {
	codes[unknownCoder.Code()] = unknownCoder
}

// Codes returns all registered error codes.
func Codes() map[int]Coder {
	return codes
}
