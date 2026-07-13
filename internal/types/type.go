package types

type Type uint8

const (
	_ Type = iota
	Bool
	Char // Rune
	Int
	Uint
	Float
	Complex
	String

	Vec
	Map

	Declaration // enum, var, struct
	Null        // null
)
