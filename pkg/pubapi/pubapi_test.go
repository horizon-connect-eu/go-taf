package pubapi_test

import (
	"testing"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-scalability-test/pkg/pubapi"
)

func TestAdd(t *testing.T) {
	expected := 3
	actual := pubapi.Add(1, 2)
	if expected != actual {
		t.Errorf("expected %d, got %d", expected, actual)
	}
}
