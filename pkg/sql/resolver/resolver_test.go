package resolver

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestResolveAlias(t *testing.T) {
	maskParams := false
	t1 := time.Now()

	tests := []struct {
		input  string
		output string
	}{
		{"select u.id from users u", "(SELECT users.id FROM users u);"},
		{"select u.id, u.name from users u", "(SELECT users.id, users.name FROM users u);"},
		{"select u.id + 7 as my_alias from users u", "(SELECT (users.id + 7) AS my_alias FROM users u);"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		r := New(program)
		r.Resolve(r.ast, env)
		checkResolveErrors(t, r, tt.input)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
	}
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestResolveAlias, Elapsed Time: %s\n", timeDiff)
}

func checkResolveErrors(t *testing.T, r *Resolver, input string) {
	errors := r.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("input: %s\nresolver has %d errors", input, len(errors))
	for _, msg := range errors {
		t.Errorf("resolver error: %q", msg)
	}
	t.FailNow()
}
