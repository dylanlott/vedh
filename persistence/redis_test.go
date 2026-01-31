package persistence

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedis(t *testing.T) {
	r, err := NewRedis("localhost:6379", "")
	if err != nil {
		t.Logf("FAILED: %s", err)
		t.Fail()
	}

	if r == nil {
		t.Logf("FAILED: Redis can't be nil - %+v", r)
		t.Fail()
	}

	val, ok, err := r.Get("testkey")
	if err != nil && (strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "operation not permitted")) {
		t.Skip("redis not available on localhost:6379")
	}
	if ok {
		t.Fail()
	}
	if err == nil {
		t.Fail()
	}
	if val != Value("") {
		t.Logf("val: %+v", val)
		t.Fail()
	}

	val, err = r.Put(Key("abc"), Value("value"))
	if err != nil && (strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), "operation not permitted")) {
		t.Skip("redis not available on localhost:6379")
	}
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, val, Value("value"))
}
