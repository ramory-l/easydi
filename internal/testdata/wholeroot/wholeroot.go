package wholeroot

// di:root
type Config struct{ N int }

// di:provide
func NewThing(
	// di:param Config
	c Config,
) Thing {
	return Thing{N: c.N}
}

type Thing struct{ N int }
