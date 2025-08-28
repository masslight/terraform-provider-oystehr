package client

type ClientConfig struct {
	ProjectID    *string
	AccessToken  *string
	ClientID     *string
	ClientSecret *string
}

type Client struct {
	config      *ClientConfig
	Application *applicationClient
	Fhir        *fhirClient
	Lab         *labClient
	M2M         *m2mClient
	Project     *projectClient
	Role        *roleClient
	Secret      *secretClient
	Z3          *z3Client
	Zambda      *zambdaClient
}

func (c *Client) Request() {

}

func New(config *ClientConfig) *Client {
	return &Client{
		config:      config,
		Application: newApplicationClient(config),
		Fhir:        newFhirClient(config),
		Lab:         newLabClient(config),
		M2M:         newM2MClient(config),
		Project:     newProjectClient(config),
		Role:        newRoleClient(config),
		Secret:      newSecretClient(config),
		Z3:          newZ3Client(config),
		Zambda:      newZambdaClient(config),
	}
}
