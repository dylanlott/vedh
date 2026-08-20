package server

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestActivationErrors(t *testing.T) {
	internal := errors.New(`pq: SQLSTATE 23505: duplicate key value violates constraint "users_username_unique"`)

	tests := []struct {
		name string
		code ActivationCode
		new  func(string, error) *gqlerror.Error
	}{
		{name: "provider unavailable", code: ActivationCodeProviderUnavailable, new: newProviderUnavailableError},
		{name: "preview error", code: ActivationCodePreviewError, new: newPreviewError},
		{name: "guest session error", code: ActivationCodeGuestSessionError, new: newGuestSessionError},
		{name: "create error", code: ActivationCodeCreateError, new: newCreateError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const productCopy = "Please try that step again."
			got := tt.new(productCopy, internal)
			if got == nil {
				t.Fatal("constructor returned nil")
			}
			if got.Message != productCopy {
				t.Fatalf("Message = %q, want product copy %q", got.Message, productCopy)
			}
			if got.Extensions["code"] != string(tt.code) {
				t.Fatalf("extensions.code = %#v, want %q", got.Extensions["code"], tt.code)
			}
			if !errors.Is(got, internal) {
				t.Fatal("underlying error is not reachable through errors.Unwrap")
			}
			for _, forbidden := range []string{"pq:", "SQLSTATE", "users_username_unique"} {
				if strings.Contains(got.Message, forbidden) {
					t.Fatalf("client-facing message leaked %q: %q", forbidden, got.Message)
				}
			}
		})
	}
}

func TestActivationErrors_ClosedVocabulary(t *testing.T) {
	want := []ActivationCode{
		ActivationCodeCreateError,
		ActivationCodeGuestSessionError,
		ActivationCodePreviewError,
		ActivationCodeProviderUnavailable,
	}
	if got := ActivationCodes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ActivationCodes() = %#v, want exact closed set %#v", got, want)
	}
}

func assertActivationCode(t *testing.T, err error, want ActivationCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected activation error code %q", want)
	}
	var gqlErr *gqlerror.Error
	if !errors.As(err, &gqlErr) {
		t.Fatalf("error type = %T, want *gqlerror.Error: %v", err, err)
	}
	if got := gqlErr.Extensions["code"]; got != string(want) {
		t.Fatalf("extensions.code = %#v, want %q", got, want)
	}
}
