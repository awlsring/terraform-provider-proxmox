package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
)

type Proxmox struct {
	client *proxmox.DefaultApiService
}

type ClientConfig struct {
	Endpoint string
	Token   string
	SkipVerify bool
	Username string
	Password string
}

func New(c ClientConfig) (*Proxmox, error) {
	cfg := proxmox.NewConfiguration()
	cfg.Servers[0] = proxmox.ServerConfiguration{
		URL: fmt.Sprintf("%s/api2/json", c.Endpoint),
	}
	cfg.HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: c.SkipVerify},
		},
	}
	
	if c.Token != "" {
		fmt.Println("making with token")
		return newTokenClient(c, cfg)
	}

	if c.Username != "" && c.Password != "" {
		fmt.Println("making with basic")
		return newBasicAuthClient(c, cfg)
	}

	return nil, fmt.Errorf("invalid client configuration")
}

func newTokenClient(c ClientConfig, cfg *proxmox.Configuration) (*Proxmox, error) {
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s", c.Token))
	client := proxmox.NewAPIClient(cfg)
	return &Proxmox{
		client: client.DefaultApi,
	}, nil
}

func newBasicAuthClient(c ClientConfig, cfg *proxmox.Configuration) (*Proxmox, error) {
	client := proxmox.NewAPIClient(cfg)
	p := &Proxmox{
		client: client.DefaultApi,
	}

	ticket, err := p.login(c.Username, c.Password)
	if err != nil {
		return nil, err
	}
	client.GetConfig().DefaultHeader["Authorization"] = fmt.Sprintf("PVEAuthCookie=%s", *ticket.Ticket)
	client.GetConfig().DefaultHeader["CSRFPreventionToken"] = *ticket.CSRFPreventionToken
	return p, nil
}

func (r *Proxmox) login(username string, password string) (*proxmox.Ticket, error) {
	request := r.client.CreateTicket(context.TODO())
	request = request.CreateTicketRequestContent(
		proxmox.CreateTicketRequestContent{
			Username: username,
			Password: password,
		},
	)
	resp, _, err := r.client.CreateTicketExecute(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if resp.Data.Ticket == nil {
		return nil, fmt.Errorf("no ticket")
	}

	if resp.Data.CSRFPreventionToken == nil {
		return nil, fmt.Errorf("no CSRF token")
	}

	return &resp.Data, nil
}