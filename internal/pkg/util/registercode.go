package util

import (
	"fmt"
	"strconv"

	"github.com/strayca7/siam/pkg/serrors"
	"github.com/strayca7/siam/pkg/util/consts"
)

// RegisterCode registers all error codes for the given service.
// All service names will be defined in the package https://github.com/strayca7/siam/internal/pkg/util
// using package-level variables.
func MustRegisterCode(service string) error {
	path, ok := CodePath[service]
	if !ok {
		return fmt.Errorf("unknown service: %s", service)
	}

	// for Code.C
	errorCodes, err := consts.ExtractInt(path)
	if err != nil {
		// this log leave to the caller processing
		return fmt.Errorf("failed to extract error codes: %w", err)
	}

	// for Code.HTTP
	errorHTTPStatus, err := consts.ExtractComment(path, consts.ParseErrHTTPStatus)
	if err != nil {
		return fmt.Errorf("failed to extract error HTTP status: %w", err)
	}

	// for Code.Ext
	errorExternal, err := consts.ExtractComment(path, consts.ParseErrExternal)
	if err != nil {
		// this log leave to the caller processing
		return fmt.Errorf("failed to extract error comments: %w", err)
	}

	for err, code := range errorCodes {
		if ext, ok := errorExternal[err]; ok {
			if http, ok := errorHTTPStatus[err]; ok {
				status, err := strconv.Atoi(http)
				if err != nil {
					return fmt.Errorf("failed to convert HTTP status to int: %w", err)
				}
				serrors.MustRegister(serrors.Code{
					C:    code,
					Ext:  ext,
					HTTP: status,
				})
			}
		}
	}
	return nil
}
