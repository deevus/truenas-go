package truenas

import (
	"errors"
	"testing"
)

func TestIsNotFoundError_Nil(t *testing.T) {
	if isNotFoundError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsNotFoundError_DoesNotExist(t *testing.T) {
	err := errors.New("object with id 42 does not exist")
	if !isNotFoundError(err) {
		t.Error("expected true for 'does not exist' error")
	}
}

func TestIsNotFoundError_ENOENT(t *testing.T) {
	err := errors.New("[ENOENT] No such file or directory")
	if !isNotFoundError(err) {
		t.Error("expected true for '[ENOENT]' error")
	}
}

func TestIsNotFoundError_Unrelated(t *testing.T) {
	err := errors.New("connection refused")
	if isNotFoundError(err) {
		t.Error("expected false for unrelated error")
	}
}

func TestIsNotFoundError_BothPatterns(t *testing.T) {
	err := errors.New("resource does not exist [ENOENT]")
	if !isNotFoundError(err) {
		t.Error("expected true when both patterns are present")
	}
}
