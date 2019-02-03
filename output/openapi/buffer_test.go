package openapi

import "testing"

func TestClear(t *testing.T) {
	b := buffer{
		indentChar: "  ",
		lines: []string{
			"hello",
			"world",
		},
	}
	b.Clear()
	res := b.Flush()
	if res != "" {
		t.Errorf("Expected \"%s\", got \"%s\"", "", res)
	}
}

func TestFlush(t *testing.T) {
	b := buffer{
		indentChar: "  ",
		lines: []string{
			"hello",
			"world",
		},
	}

	res := b.Flush()
	if res != "helloworld" {
		t.Errorf("Expected \"%s\", got \"%s\"", "helloworld", res)
	}
}

func TestWrite(t *testing.T) {
	b := buffer{
		indentChar: "  ",
		lines:      []string{""},
	}

	b.Write("hello", 0)
	b.Write("world", 1)
	res := b.Flush()
	if res != "hello  world" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello  world", res)
	}
}

func TestLine(t *testing.T) {
	b := buffer{
		indentChar: "  ",
		lines:      []string{""},
	}

	b.Line("hello", 0)
	b.Line("world", 1)
	res := b.Flush()
	if res != "hello\n  world\n" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello\n  world\n", res)
	}
}

func TestLabel(t *testing.T) {
	b := buffer{
		indentChar: "  ",
		lines:      []string{""},
	}

	b.Label("hello", 0)
	b.Label("world", 1)
	res := b.Flush()
	if res != "hello:\n  world:\n" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello:\n  world:\n", res)
	}
}

func TestKeyValue(t *testing.T) {
	b := buffer{
		indentChar: "  ",
		lines:      []string{""},
	}

	b.KeyValue("hello", "lorem", 0)
	b.KeyValue("world", "ipsum", 1)
	res := b.Flush()
	if res != "hello: lorem\n  world: ipsum\n" {
		t.Errorf("Expected \"%s\", got \"%s\"", "hello: lorem\n  world: ipsum\n", res)
	}
}
