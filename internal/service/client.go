package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Proxmox struct {
	client *proxmox.DefaultApiService
	IsRoot bool
}

type ClientConfig struct {
	Endpoint   string
	Token      string
	SkipVerify bool
	Username   string
	Password   string
}

func New(c ClientConfig) (*Proxmox, error) {
	tflog.Debug(context.Background(), "Unrecognized API response body")
	cfg := proxmox.NewConfiguration()
	cfg.Servers[0] = proxmox.ServerConfiguration{
		URL: fmt.Sprintf("%s/api2/json", c.Endpoint),
	}
	cfg.HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: c.SkipVerify},
		},
	}

	if c.Token != "" {
		return newTokenClient(c, cfg)
	}

	if c.Username != "" && c.Password != "" {
		return newBasicAuthClient(c, cfg)
	}

	return nil, fmt.Errorf("invalid client configuration")
}

func newTokenClient(c ClientConfig, cfg *proxmox.Configuration) (*Proxmox, error) {
	cfg.AddDefaultHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s", c.Token))
	client := proxmox.NewAPIClient(cfg)
	return &Proxmox{
		client: client.DefaultApi,
		IsRoot: false,
	}, nil
}

func newBasicAuthClient(c ClientConfig, cfg *proxmox.Configuration) (*Proxmox, error) {
	client := proxmox.NewAPIClient(cfg)
	p := &Proxmox{
		client: client.DefaultApi,
		IsRoot: c.Username == "root@pam",
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
