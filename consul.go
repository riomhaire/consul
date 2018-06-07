package consul

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

//Client provides an interface for getting data out of Consul
type Client interface {
	// Get a Service from consul
	Service(string, string) ([]string, error)
	// Register a service with local agent
	Register(string, int) error
	// Deregister a service with local agent
	DeRegister(string) error
}

type ConsulClient struct {
	Consul *consul.Client
}

//NewConsul returns a Client interface for given consul address
func NewConsulClient(addr string) (*ConsulClient, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return &ConsulClient{}, err
	}
	return &ConsulClient{Consul: c}, nil
}

// Register a service with consul local agent
func (c *ConsulClient) Register(id, name, host string, port int, path, health string) error {

	reg := &consul.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Port:    port,
		Address: host,
		Check: &consul.AgentServiceCheck{
			CheckID:       id,
			Name:          "HTTP API health",
			HTTP:          health,
			TLSSkipVerify: true,
			Method:        "GET",
			Interval:      "10s",
			Timeout:       "1s",
		},
		Tags: []string{
			"traefik.backend=" + name,
			"traefik.frontend.rule=PathPrefix:" + path,
		},
	}
	return c.Consul.Agent().ServiceRegister(reg)
}

// DeRegister a service with consul local agent
func (c *ConsulClient) DeRegister(id string) error {
	return c.Consul.Agent().ServiceDeregister(id)
}

// Service return a service
func (c *ConsulClient) Service(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.Consul.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}
