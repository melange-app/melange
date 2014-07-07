package melange

type Server interface {
  Run(port int) error
}

type Enabler struct {
  Server
}

func (t *Enabler) Enable(port int) error {
  return t.Server.Run(port)
}
