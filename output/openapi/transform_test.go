package openapi

import "testing"

func TestTrsEmpty(t *testing.T) {
	res := trsEmpty("hello")
	if res != "hello" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello", res)
	}
}

func TestTrsChain(t *testing.T) {
	trs1 := func(input string) string {
		return "a" + input
	}
	trs2 := func(input string) string {
		return input + "b"
	}
	chain := trsChain([]transformation{trs1, trs2})
	res := chain("hello")
	if res != "ahellob" {
		t.Errorf("Expected \"%s\", got \"%s\"", "ahellob", res)
	}
}

func TestTrsType(t *testing.T) {
	trs := newTrsType()
	res := trs("int16")
	if res != "integer" {
		t.Errorf("Expected \"%s\", got \"%s\"", "integer", res)
	}
	res = trs("unknown")
	if res != "unknown" {
		t.Errorf("Expected \"%s\", got \"%s\"", "unknown", res)
	}
}

func TestTrsMediaType(t *testing.T) {
	trs := newTrsMediaType()
	res := trs("multipart")
	if res != "multipart/form-data" {
		t.Errorf("Expected \"%s\", got \"%s\"", "multipart/form-data", res)
	}
	res = trs("application/json")
	if res != "application/json" {
		t.Errorf("Expected \"%s\", got \"%s\"", "application/json", res)
	}
}

func TestTrsQuote(t *testing.T) {
	trs := newTrsQuote()
	res := trs("hello")
	if res != "\"hello\"" {
		t.Errorf("Expected \"%s\", got \"%s\"", "\"hello\"", res)
	}
	res = trs("\"hello\"")
	if res != "\"hello\"" {
		t.Errorf("Expected \"%s\", got \"%s\"", "\"hello\"", res)
	}
}

func TestTrsSpecialChars(t *testing.T) {
	trs := newTrsSpecialChars()
	res := trs("[hel{}lo]]")
	if res != "hello" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello", res)
	}
	res = trs("hello")
	if res != "hello" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello", res)
	}
}

func TestTrsArray(t *testing.T) {
	trs := newTrsArray()
	res := trs("{[]object}")
	if res != "{array object}" {
		t.Errorf("Expected \"%s\", got \"%s\"", "{array object}", res)
	}
	res = trs("[]object")
	if res != "[]object" {
		t.Errorf("Expected \"%s\", got \"%s\"", "[]object", res)
	}
	res = trs("object")
	if res != "object" {
		t.Errorf("Expected \"%s\", got \"%s\"", "object", res)
	}
}
