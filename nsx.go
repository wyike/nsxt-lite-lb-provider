package main

import (
	"crypto/tls"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// NSXClient struct to hold the Resty client
type NSXClient struct {
	client *resty.Client
	router string
	vip    string
}

// NewNSXClient initializes the NSX client
func NewNSXClient(nsxHost, nsxUser, nsxPassword, clusterVIP, clusterRouter string) *NSXClient {
	client := resty.New().
		SetBaseURL(nsxHost).
		SetBasicAuth(nsxUser, nsxPassword).
		SetHeaders(map[string]string{
			"Content-Type":      "application/json",
			"X-Allow-Overwrite": "true",
		}).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &NSXClient{
		client: client,
		router: clusterRouter,
		vip:    clusterVIP,
	}
}

type Tier1StaticRoutes struct {
	// the type is ipv4 or ipv6, the format is CIDR
	Network *string `json:"network"`
	// next hops
	NextHops []*NextHop `json:"next_hops"`
}

type NextHop struct {
	// admin distance
	// Maximum: 255
	// Minimum: 1
	AdminDistance int32 `json:"admin_distance,omitempty"`

	// ip address
	IPAddress *string `json:"ip_address"`

	// Configure the interface paths of tier-0
	Scope []string `json:"scope"`
}

type Tier1StaticRoutesListResult struct {
	Results []*Tier1StaticRoutes
}

func (nc *NSXClient) GetStaticRoute() ([]*Tier1StaticRoutes, error) {
	// Make the request to get static routes
	url := fmt.Sprintf("/policy/api/v1/infra/tier-1s/%s/static-routes", nc.router)
	resp, err := nc.client.R().
		SetResult(&Tier1StaticRoutesListResult{}). // Set the result type, resps will be marshal to it
		Get(fmt.Sprintf(url))

	if err != nil {
		return nil, fmt.Errorf("failed to get static routes: %v", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: %s", resp.Status())
	}

	// Parse the response
	routes := resp.Result().(*Tier1StaticRoutesListResult)
	return routes.Results, nil
}

func (nc *NSXClient) PatchStaticRoute(nextHops []string) error {
	// Define the static route payload
	r := fmt.Sprintf("to-%s", nc.vip)
	n := fmt.Sprintf("%s/32", nc.vip)

	hops := []map[string]interface{}{}
	for _, h := range nextHops {
		hops = append(hops, map[string]interface{}{
			"ip_address":     h,
			"admin_distance": 1, // Optional field
		})
	}
	payload := map[string]interface{}{
		"id":           r,
		"display_name": r,
		"network":      n,
		"next_hops":    hops,
	}

	// Make the request to create a static route
	url := fmt.Sprintf("/policy/api/v1/infra/tier-1s/%s/static-routes/%s", nc.router, r)
	resp, err := nc.client.R().
		SetBody(payload).
		Patch(url)

	if err != nil {
		return fmt.Errorf("failed to create static route: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("NSX API error: %s", resp.String())
	}

	log.WithName("nsxt").Info("static route is created", "name", r, "hops", hops)
	return nil
}

func testNSXClient() {
	nc := NewNSXClient("https://x", "admin", "y", "10.10.10.1", "z")
	if _, err := nc.GetStaticRoute(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	return
}
