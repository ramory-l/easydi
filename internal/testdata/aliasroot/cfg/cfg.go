package cfg

type Config struct{ Addr string }

type Server struct{ addr string }

// di:provide name=Server
// di:expose
func NewServer(
	// di:param Config.Addr
	addr string,
) *Server { return &Server{addr: addr} }
