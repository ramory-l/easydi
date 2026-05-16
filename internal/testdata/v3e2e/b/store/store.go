package store

import astore "github.com/ramory-l/easydi/internal/testdata/v3e2e/a/store"

type Svc struct{ name string }

func (s *Svc) Get() string { return s.name }

// di:provide name=Svc
func NewSvc(
	// di:param Config.Name
	name string,
	_ *astore.Repo,
) *Svc {
	return &Svc{name: name}
}
