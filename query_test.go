package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestRegexMessage(t *testing.T) {
	// check standard statement
	message := "duration: 0.083 ms  statement: SET time zone 'UTC'"
	result := regexMessage(message)
	assert.Equal(t, "statement", result["preparedStep"])

	// check prepared statement
	message = "duration: 0.066 ms  bind <unnamed>: select * from servers where id = 1"
	result = regexMessage(message)
	assert.Equal(t, "bind", result["preparedStep"])

	// check non-greedy to colon
	message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah'"
	result = regexMessage(message)
	assert.Equal(t, "bind", result["preparedStep"])

	// check multiline
	message = `duration: 0.066 ms  bind <unnamed>: select * from servers
	where name = 'blah:blah'`
	multiLine := `select * from servers
	where name = 'blah:blah'`
	result = regexMessage(message)
	assert.Equal(t, multiLine, result["grokQuery"])
}

func TestComments(t *testing.T) {
	var q = new(query)
	// No comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah'"
	assert.NotPanics(t, func() { extractComments(q) })

	// Legit comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*application:Rails,controller:users,action:search,line:monitor.rb:214:in '::Weekly::Digest/mon_synchronize'*/"
	assert.NotPanics(t, func() { extractComments(q) })

	// not complete comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*application:Rails*/"
	assert.NotPanics(t, func() { extractComments(q) })

	// Empty comment
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /**/"
	assert.NotPanics(t, func() { extractComments(q) })

	// illegit comment with colon
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*ask:yourmom*/"
	assert.NotPanics(t, func() { extractComments(q) })

	q.message = "/*ask:you:*/"
	assert.NotPanics(t, func() { extractComments(q) })

	q.message = "/*:*/"
	assert.NotPanics(t, func() { extractComments(q) })

	q.message = `duration: 0.066 ms  bind <unnamed>: select * from multiline /*
	: blah */`
	assert.NotPanics(t, func() { extractComments(q) })

	// Multiples
	q.message = "duration: 0.066 ms  bind <unnamed>: select * from servers where name = 'blah:blah' /*application:Rails*/ /*application:Sidekiq*/"
	assert.NotPanics(t, func() { extractComments(q) })
}
