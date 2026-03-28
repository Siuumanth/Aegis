package resp

type Parser interface {
	Parse([]byte) (*Command, error)
}
