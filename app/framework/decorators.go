package framework

type CSPWrapper struct {
	CSP string
	View
}

func (c *CSPWrapper) Headers() Headers {
	temp := c.View.Headers()
	if temp == nil {
		temp = make(map[string]string)
	}
	temp["Content-Security-Policy"] = c.CSP
	return temp
}
