package require

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func NoError(t testing.TB, err error) {
	if err == nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	t.Fatalf("\n"+
		"Error Trace: %s:%d\n"+
		"Error:       Received unexpected error:\n"+
		"             %+v\n"+
		"Test:        %s",
		file, line, err, t.Name(),
	)
}

func Equal(t testing.TB, expected, actual any) {
	if reflect.DeepEqual(expected, actual) {
		return
	}

	var expectedJson, actualJson []byte
	var err error
	if s, ok := expected.(string); ok {
		expectedJson = []byte(s)
	} else {
		expectedJson, err = json.Marshal(expected)
		if err != nil {
			expectedJson = []byte(fmt.Sprintf("%+v", expected))
		}
	}
	if s, ok := actual.(string); ok {
		actualJson = []byte(s)
	} else {
		actualJson, err = json.Marshal(actual)
		if err != nil {
			actualJson = []byte(fmt.Sprintf("%+v", actual))
		}
	}

	_, file, line, _ := runtime.Caller(1)
	t.Fatalf("\n"+
		"Error Trace: %s:%d\n"+
		"Error:       Not equal:\n"+
		"             expected: %s\n"+
		"             actual  : %s\n"+
		"Test:        %s",
		file, line, expectedJson, actualJson, t.Name(),
	)
}

func NotEqual(t testing.TB, expected, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		return
	}

	var expectedJson []byte
	var err error
	if s, ok := expected.(string); ok {
		expectedJson = []byte(s)
	} else {
		expectedJson, err = json.Marshal(expected)
		if err != nil {
			expectedJson = []byte(fmt.Sprintf("%+v", expected))
		}
	}

	_, file, line, _ := runtime.Caller(1)
	t.Fatalf("\n"+
		"Error Trace: %s:%d\n"+
		"Error:       Should not be:\n"+
		"             %s\n"+
		"Test:        %s",
		file, line, expectedJson, t.Name(),
	)
}
