package client

type ClientConfig struct {
	ProjectID    *string
	AccessToken  *string
	ClientID     *string
	ClientSecret *string
}

type Client struct {
	config      ClientConfig
	Fhir        *fhirClient
	Application *applicationClient
	Role        *roleClient
	M2M         *m2mClient
	Secret      *secretClient
	Z3          *z3Client
	Zambda      *zambdaClient
}

func (c *Client) Request() {

}

func New(config ClientConfig) *Client {
	return &Client{
		config:      config,
		Fhir:        newFhirClient(config),
		Application: newApplicationClient(config),
		Role:        newRoleClient(config),
		M2M:         newM2MClient(config),
		Secret:      newSecretClient(config),
		Z3:          newZ3Client(config),
		Zambda:      newZambdaClient(config),
	}
}
