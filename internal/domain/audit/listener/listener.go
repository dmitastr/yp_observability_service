package listener

import (
	"github.com/dmitastr/yp_observability_service/internal/domain/audit/data"
)

type Type int

const (
	FileListenerType Type = iota
	URLListenerType
)

type IListener interface {
	Notify(*data.Data) error
}

func NewListener(lType Type, path *string) IListener {
	if path == nil || *path == "" {
		return nil
	}
	switch lType {
	case FileListenerType:
		return NewFileListener(*path)
	case URLListenerType:
		return NewURLListener(*path)
	default:
		return nil
	}
}
