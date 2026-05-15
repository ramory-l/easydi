package simple

// Config is a di:root.
//
// di:root
type Config struct {
	Auth AuthConfig
}

type AuthConfig struct {
	Secret string
}

// Infra is a di:root.
//
// di:root
type Infra struct {
	DB DB
}

type DB struct{ dsn string }

func (d DB) DSN() string { return d.dsn }

type Hasher struct{ salt string }

type Repo interface{ Find() string }

type repo struct{ h Hasher }

func (r *repo) Find() string { return r.h.salt }

type Service struct {
	R   Repo
	Sec string
}
