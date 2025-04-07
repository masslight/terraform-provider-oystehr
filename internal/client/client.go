package client

type ClientConfig struct {
	ProjectID    *string
	AccessToken  *string
	ClientID     *string
	ClientSecret *string
}

type Client struct {
	config ClientConfig
	Fhir   *fhirClient
}

func (c *Client) Request() {

}

func New(config ClientConfig) *Client {
	return &Client{config: config,
		Fhir: newFhirClient(config)}
}
