package simple

// di:provide
func NewHasher(
	// di:param Infra.DB.DSN()
	dsn string,
) Hasher {
	return Hasher{salt: dsn}
}

// di:provide
func NewRepo(h Hasher) Repo {
	return &repo{h: h}
}

// di:provide
// di:expose
func NewService(
	r Repo,
	// di:param Config.Auth.Secret
	sec string,
) (Service, error) {
	return Service{R: r, Sec: sec}, nil
}
