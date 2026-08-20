package server

import (
	"sort"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ActivationCode is the closed, server-owned vocabulary used by activation
// resolvers. The browser maps these values to one product-copy and affordance
// pair; callers cannot construct a GraphQL activation error with an arbitrary
// code because the only constructors below each bind one allowlisted value.
type ActivationCode string

const (
	ActivationCodeProviderUnavailable ActivationCode = "provider_unavailable"
	ActivationCodePreviewError        ActivationCode = "preview_error"
	ActivationCodeGuestSessionError   ActivationCode = "guest_session_error"
	ActivationCodeCreateError         ActivationCode = "create_error"
)

var activationCodeAllowlist = map[ActivationCode]struct{}{
	ActivationCodeProviderUnavailable: {},
	ActivationCodePreviewError:        {},
	ActivationCodeGuestSessionError:   {},
	ActivationCodeCreateError:         {},
}

// ActivationCodes returns a sorted copy of the complete activation-code set.
// Widening the vocabulary requires changing the allowlist and its closure test.
func ActivationCodes() []ActivationCode {
	codes := make([]ActivationCode, 0, len(activationCodeAllowlist))
	for code := range activationCodeAllowlist {
		codes = append(codes, code)
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })
	return codes
}

func newProviderUnavailableError(message string, cause error) *gqlerror.Error {
	return newActivationError(ActivationCodeProviderUnavailable, message, cause)
}

func newPreviewError(message string, cause error) *gqlerror.Error {
	return newActivationError(ActivationCodePreviewError, message, cause)
}

func newGuestSessionError(message string, cause error) *gqlerror.Error {
	return newActivationError(ActivationCodeGuestSessionError, message, cause)
}

func newCreateError(message string, cause error) *gqlerror.Error {
	return newActivationError(ActivationCodeCreateError, message, cause)
}

// newActivationError deliberately keeps product copy and the server-side
// cause in separate fields. gqlgen serializes Message and Extensions, while
// Err remains available to errors.Unwrap and is never sent to the browser.
func newActivationError(code ActivationCode, message string, cause error) *gqlerror.Error {
	return &gqlerror.Error{
		Err:     cause,
		Message: message,
		Extensions: map[string]interface{}{
			"code": string(code),
		},
	}
}
