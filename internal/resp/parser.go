package resp

import "Aegis/internal/model"

type Parser interface {
	Parse([]byte) (*model.Command, error)
}
