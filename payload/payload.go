package payload

type Payload struct {
	Header       map[string]string
	Body         []byte
	UrlParameter string
	StatusCode   int
	TimeCost     int
}
