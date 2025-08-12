/*
Package code defines error codes for siam platform.
On the base of http error codes below we define our own error codes.
Each custom error code corresponds to a specific http error code.
Each http error code at least corresponds to one custom error code.

siam code only allowed the following http code:

	StatusOK                           = 200 // RFC 7231, 6.3.
	StatusBadRequest                   = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 = 401 // RFC 7235, 3.1
	StatusForbidden                    = 403 // RFC 7231, 6.5.3
	StatusNotFound                     = 404 // RFC 7231, 6.5.4
	StatusInternalServerError          = 500 // RFC 7231, 6.6.1
*/
package code
