package audit

import (
	"errors"

	"github.com/dmitastr/yp_observability_service/internal/domain/audit/data"
	"github.com/dmitastr/yp_observability_service/internal/domain/audit/listener"
)

type IAuditor interface {
	AddListener(listener.IListener) IAuditor
	Notify(*data.Data) error
}

type Auditor struct {
	listeners []listener.IListener
}

func NewAuditor() *Auditor {
	return &Auditor{listeners: make([]listener.IListener, 0)}
}

func (a *Auditor) AddListener(l listener.IListener) IAuditor {
	if l != nil {
		a.listeners = append(a.listeners, l)
	}
	return a
}

func (a *Auditor) Notify(data *data.Data) error {
	var errs error

	for _, l := range a.listeners {
		if err := l.Notify(data); err != nil {
			errors.Join(errs, err)
		}
	}
	return errs
}
