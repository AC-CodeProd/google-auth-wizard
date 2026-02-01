package googlescopes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type GoogleServices map[string][]Scope

type Scope struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type scopeResponse struct {
	Description string `json:"description"`
}

type apiInfoResponse struct {
	IconURL string                     `json:"iconUrl"`
	Scopes  []map[string]scopeResponse `json:"scopes"`
}

type getScopesResponse struct {
	Success bool                       `json:"success"`
	Apis    map[string]apiInfoResponse `json:"apis"`
}

type Client struct {
	httpClient    *http.Client
	baseURL       string
	scopeEndpoint string
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithScopeEndpoint(scopeEndpoint string) ClientOption {
	return func(c *Client) {
		c.scopeEndpoint = scopeEndpoint
	}
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:       "https://developers.google.com/oauthplayground",
		scopeEndpoint: "getScopes",
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func (c *Client) FetchScopes() (*GoogleServices, error) {
	return c.FetchScopesWithContext(context.Background())
}

func (c *Client) FetchScopesWithContext(ctx context.Context) (*GoogleServices, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, c.scopeEndpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, resp.Status)
	}

	var tempAPIs getScopesResponse
	if err := json.NewDecoder(resp.Body).Decode(&tempAPIs); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	if !tempAPIs.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return c.reorganizeScopes(tempAPIs.Apis), nil
}

func (c *Client) reorganizeScopes(tempAPIs map[string]apiInfoResponse) *GoogleServices {
	googleServices := make(GoogleServices, len(tempAPIs))

	for apiName, apiInfo := range tempAPIs {
		scopes := make([]Scope, 0, len(apiInfo.Scopes))

		for _, scopeMap := range apiInfo.Scopes {
			for scopeURL, scopeInfo := range scopeMap {
				scope := Scope{
					URL:         scopeURL,
					Description: scopeInfo.Description,
				}
				scopes = append(scopes, scope)
			}
		}

		sort.Slice(scopes, func(i, j int) bool {
			return scopes[i].URL < scopes[j].URL
		})

		googleServices[apiName] = scopes
	}

	return &googleServices
}

func (gs *GoogleServices) GetScopesForService(serviceName string) ([]Scope, bool) {
	scopes, exists := (*gs)[serviceName]
	return scopes, exists
}

func (gs *GoogleServices) GetAllServiceNames() []string {
	names := make([]string, 0, len(*gs))
	for name := range *gs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (gs *GoogleServices) GetTotalScopeCount() int {
	total := 0
	for _, scopes := range *gs {
		total += len(scopes)
	}
	return total
}

func (gs *GoogleServices) FindScopesByURL(searchURL string) map[string][]Scope {
	results := make(map[string][]Scope)

	for serviceName, scopes := range *gs {
		for _, scope := range scopes {
			if scope.URL == searchURL {
				results[serviceName] = append(results[serviceName], scope)
			}
		}
	}

	return results
}

func (gs *GoogleServices) FindScopesByDescription(searchTerm string) map[string][]Scope {
	results := make(map[string][]Scope)
	searchLower := strings.ToLower(searchTerm)

	for serviceName, scopes := range *gs {
		for _, scope := range scopes {
			if strings.Contains(strings.ToLower(scope.Description), searchLower) {
				results[serviceName] = append(results[serviceName], scope)
			}
		}
	}

	return results
}

func (gs *GoogleServices) GetScopeByURL(searchURL string) (*Scope, string, bool) {
	for serviceName, scopes := range *gs {
		for _, scope := range scopes {
			if scope.URL == searchURL {
				return &scope, serviceName, true
			}
		}
	}
	return nil, "", false
}

func (gs *GoogleServices) IsEmpty() bool {
	return len(*gs) == 0
}

func (gs *GoogleServices) HasService(serviceName string) bool {
	_, exists := (*gs)[serviceName]
	return exists
}

func (gs *GoogleServices) GetServiceCount() int {
	return len(*gs)
}

func (gs *GoogleServices) ToJSON() ([]byte, error) {
	return json.MarshalIndent(gs, "", "  ")
}

func FromJSON(data []byte) (*GoogleServices, error) {
	var gs GoogleServices
	if err := json.Unmarshal(data, &gs); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	return &gs, nil
}
