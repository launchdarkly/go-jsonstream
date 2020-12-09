package commontest

import (
	"errors"
	"fmt"
	"reflect"
)

// These functions provide a simple mechanism for returning errors as errors instead of using
// assert or require.

func AssertEqual(expected, actual interface{}) error {
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}
	return nil
}

func AssertTrue(value bool, failureMessage string) error {
	if !value {
		return errors.New(failureMessage)
	}
	return nil
}

func AssertNoErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
