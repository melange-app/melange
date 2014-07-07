package melange

type Server interface {
  Run(port int) error
}

type Delegate interface {
  Post(f *Server, m *Message)
}

type Message struct {

}
