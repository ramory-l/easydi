package store

type Lookup interface{ Get() string }

type Repo struct{}

func (Repo) Get() string { return "repo" }

// di:provide name=Repo
func NewRepo() *Repo { return &Repo{} }
