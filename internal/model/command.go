package model

type Command struct {
	Name string // GET, SET, AEGIS.*
	Key  string
	Args []string
	Raw  []byte // original RESP (important)
}
