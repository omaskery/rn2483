package testutils

import (
	"errors"
	"fmt"

	"github.com/poy/onpar/matchers"
)

type errorMatcher struct {
	expected error
}

// MatchError creates a matcher for asserting if an error is convertible to the expect error using errors.Is
func MatchError(err error) *errorMatcher {
	return &errorMatcher{
		expected: err,
	}
}

var _ matchers.Matcher = (*errorMatcher)(nil)

// Match implements the matchers.Matcher interface
func (e errorMatcher) Match(actual interface{}) (resultValue interface{}, err error) {
	actualErr, ok := actual.(error)
	if !ok {
		return actual, fmt.Errorf("expected an error type, got type %T", actual)
	}

	if !errors.Is(actualErr, e.expected) {
		return actual, fmt.Errorf("expected error type %T, got type %T", e.expected, actualErr)
	}

	return actual, nil
}
