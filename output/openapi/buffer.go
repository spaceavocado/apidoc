package openapi

import (
	"fmt"
	"strings"
)

type buffer struct {
	indentChar string
	lines      []string
}

// Clear the buffer
func (b *buffer) Clear() {
	b.lines = make([]string, 0)
}

// Flush the buffer lines into the output string
func (b *buffer) Flush() string {
	return strings.Join(b.lines, "")
}

// Write into the buffer with a desired indentation
func (b *buffer) Write(content string, indent int) {
	b.lines = append(b.lines, fmt.Sprintf("%s%s", strings.Repeat(b.indentChar, indent), content))
}

// Line writes a new line into the buffer,
// with desired indentation, ended with a new line char
func (b *buffer) Line(content string, indent int) {
	b.Write(fmt.Sprintf("%s\n", content), indent)
}

// Label writes a label into the buffer,
// with desired indentation, ended with a new line char
func (b *buffer) Label(label string, indent int) {
	b.Write(fmt.Sprintf("%s:\n", strings.TrimSpace(label)), indent)
}

// KeyValue writes a key/value pair into the buffer,
// with desired indentation, ended with a new line char
func (b *buffer) KeyValue(key string, value string, indent int) {
	b.Write(fmt.Sprintf("%s: %s\n", strings.TrimSpace(key), value), indent)
}
