package composer

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composite"
)

func makeComposite(jsonStr string) *resource.Composite {
	xr := composite.New()
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		panic(err)
	}
	xr.SetUnstructuredContent(data)
	return &resource.Composite{Resource: xr}
}

func TestComposer_GetString(t *testing.T) {
	c := New(makeComposite(`{
		"apiVersion": "test/v1",
		"kind": "Test",
		"spec": {"name": "test-value"}
	}`))

	val := c.GetString("spec.name")
	if val != "test-value" {
		t.Errorf("expected test-value, got %s", val)
	}
	if c.Err() != nil {
		t.Errorf("expected no error, got %v", c.Err())
	}
}

func TestComposer_GetString_NotFound(t *testing.T) {
	c := New(makeComposite(`{
		"apiVersion": "test/v1",
		"kind": "Test",
		"spec": {}
	}`))

	val := c.GetString("spec.missing")
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
	if c.Err() == nil {
		t.Error("expected error")
	}
	if !errors.Is(c.Err(), ErrPathNotFound) {
		t.Errorf("expected ErrPathNotFound, got %v", c.Err())
	}
}

func TestComposer_GetBool(t *testing.T) {
	c := New(makeComposite(`{
		"apiVersion": "test/v1",
		"kind": "Test",
		"spec": {"enabled": true}
	}`))

	val := c.GetBool("spec.enabled")
	if !val {
		t.Error("expected true")
	}
	if c.Err() != nil {
		t.Errorf("expected no error, got %v", c.Err())
	}
}

func TestComposer_Add(t *testing.T) {
	c := New(makeComposite(`{
		"apiVersion": "test/v1",
		"kind": "Test"
	}`))

	c.Add("resource-a", "value-a")
	c.Add("resource-b", "value-b")

	if len(c.Desired()) != 2 {
		t.Errorf("expected 2 desired, got %d", len(c.Desired()))
	}
}

func TestComposer_ClearErrs(t *testing.T) {
	c := New(makeComposite(`{
		"apiVersion": "test/v1",
		"kind": "Test",
		"spec": {}
	}`))

	c.GetString("spec.missing")
	if c.Err() == nil {
		t.Error("expected error")
	}

	c.ClearErrs()
	if c.Err() != nil {
		t.Error("expected no error after clear")
	}
}
