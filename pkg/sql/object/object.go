package object

import (
	"fmt"
	"hash/fnv"

	"github.com/google/uuid"
)

type BuiltinFunction func(args ...Object) Object
type ObjectType int

const (
	_ ObjectType = iota
	NULL
	ERROR
	INTEGER
	BOOLEAN
	STRING
	STRING_HASH_OBJ // This is only used internally. It is not a part of the language.
	UUID
)

var Objects = [...]string{
	NULL:            "NULL",
	ERROR:           "ERROR",
	INTEGER:         "INTEGER",
	BOOLEAN:         "BOOLEAN",
	STRING:          "STRING",
	STRING_HASH_OBJ: "STRING_HASH",
	UUID:            "UUID",
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hashable interface {
	HashKey() HashKey
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type UID struct {
	Value uuid.UUID
}

func (u *UID) Inspect() string  { return u.Value.String() }
func (u *UID) Type() ObjectType { return UUID }
func (u *UID) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(u.Value.String()))

	return HashKey{Type: u.Type(), Value: h.Sum64()}
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN }
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// This is only used internally. It is not a part of the language.
type StringHash struct {
	Value map[string]string
}

func (s *StringHash) Type() ObjectType { return STRING }
func (s *StringHash) Inspect() string  { return fmt.Sprintf("%v", s.Value) }
