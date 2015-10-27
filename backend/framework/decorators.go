package framework

// CORSWrapper will wrap a Handler and return CORS-compliant headers
// for the AppURL domain.
type CORSWrapper struct {
	Origin string
	View
}

// Headers add the CORS headers to the request:
//
// "Access-Control-Allow-Origin"
// "Access-Control-Allow-Headers"
func (c *CORSWrapper) Headers() Headers {
	hdrs := c.View.Headers()
	if hdrs == nil {
		hdrs = make(map[string]string)
	}

	// Include the CORS Headers
	hdrs["Access-Control-Allow-Origin"] = c.Origin
	hdrs["Access-Control-Allow-Headers"] = "Content-Type"
	return hdrs
}

// CSPWrapper includes headers for Content Security Policy.
type CSPWrapper struct {
	CSP string
	View
}

// Headers add the CSP headers to the request:
//
// "Content-Security-Policy"
func (c *CSPWrapper) Headers() Headers {
	hdrs := c.View.Headers()
	if hdrs == nil {
		hdrs = make(map[string]string)
	}

	hdrs["Content-Security-Policy"] = c.CSP
	return hdrs
}
