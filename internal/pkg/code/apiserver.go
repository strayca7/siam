package code

// code for API server

// siam-apiserver: secret errors.
const (
	// ErrUserNotFound - 404: User not found.
	ErrUserNotFound = iota + 110001

	// ErrUserAlreadyExists - 409: User already exists.
	ErrUserAlreadyExists
)

// siam-apiserver: secret errors.
const (
	// ErrReachMaxCount - 429: Reach max count.
	ErrReachMaxCount = iota + 110101

	// ErrSecretNotFound - 404: Secret not found.
	ErrSecretNotFound
)

// siam-apiserver: policy errors.
const (
	// ErrPolicyNotFound - 404: Policy not found.
	ErrPolicyNotFound = iota + 110201
)
