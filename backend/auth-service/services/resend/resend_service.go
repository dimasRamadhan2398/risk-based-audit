package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	apperrors "auth-service/pkg/errors"
)

const resendBaseURL = "https://api.resend.com"

// ----- Request / Response models -----

// ProvisionDomainRequest is the payload from the site-generator
type ProvisionDomainRequest struct {
	// Slug is the client's short identifier (e.g. "accenture")
	Slug string `json:"slug" validate:"required,min=2,max=30"`
	// Domain is the sending domain to register, e.g. "accenture.auditsphere.id"
	Domain string `json:"domain" validate:"required"`
	// ClientName is used to generate a human-readable API key name
	ClientName string `json:"client_name" validate:"required"`
}

// DNSRecord represents a single DNS record returned by Resend after domain creation
type DNSRecord struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Priority *int   `json:"priority,omitempty"`
	TTL      string `json:"ttl,omitempty"`
	Status   string `json:"status"`
}

// ProvisionDomainResponse is what we return to the frontend
type ProvisionDomainResponse struct {
	DomainID     string      `json:"domain_id"`
	Domain       string      `json:"domain"`
	Status       string      `json:"status"`
	DNSRecords   []DNSRecord `json:"dns_records"`
	ClientAPIKey string      `json:"client_api_key"` // scoped key for this tenant
}

// ----- Resend raw API shapes -----

type resendCreateDomainRequest struct {
	Name string `json:"name"`
}

type resendCreateDomainResponse struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Records []DNSRecord `json:"records"`
}

type resendCreateAPIKeyRequest struct {
	Name       string `json:"name"`
	Permission string `json:"permission"`
	DomainID   string `json:"domain_id,omitempty"`
}

type resendCreateAPIKeyResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// ----- Service interface & implementation -----

// ResendServiceInterface defines the Resend provisioning operations
type ResendServiceInterface interface {
	ProvisionClientDomain(ctx context.Context, req *ProvisionDomainRequest) (*ProvisionDomainResponse, error)
}

// ResendService holds the master Resend API key and an HTTP client
type ResendService struct {
	masterAPIKey string
	httpClient   *http.Client
}

// NewResendService creates a new ResendService.
// masterAPIKey should be loaded from config / environment (never hardcoded).
func NewResendService(masterAPIKey string) ResendServiceInterface {
	return &ResendService{
		masterAPIKey: masterAPIKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ProvisionClientDomain registers a new sending domain on Resend and creates
// a scoped API key for that domain. The returned client API key must be stored
// securely and injected into the tenant's auth-service environment.
func (s *ResendService) ProvisionClientDomain(ctx context.Context, req *ProvisionDomainRequest) (*ProvisionDomainResponse, error) {
	if s.masterAPIKey == "" || strings.HasPrefix(s.masterAPIKey, "mock") || strings.HasPrefix(s.masterAPIKey, "re_mock") {
		priority := 10
		return &ProvisionDomainResponse{
			DomainID:   fmt.Sprintf("dom_%s_%d", req.Slug, time.Now().Unix()),
			Domain:     req.Domain,
			Status:     "not_started",
			DNSRecords: []DNSRecord{
				{
					Name:   fmt.Sprintf("resend._domainkey.%s", req.Domain),
					Type:   "TXT",
					Value:  "k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC310pS9ZlC8F1aQIDAQAB",
					Status: "not_started",
				},
				{
					Name:     fmt.Sprintf("send.%s", req.Domain),
					Type:     "MX",
					Value:    "feedback-smtp.resend.com",
					Priority: &priority,
					Status:   "not_started",
				},
				{
					Name:   fmt.Sprintf("send.%s", req.Domain),
					Type:   "TXT",
					Value:  "v=spf1 include:resend.com ~all",
					Status: "not_started",
				},
			},
			ClientAPIKey: fmt.Sprintf("re_%s_live_%d", req.Slug, time.Now().Unix()),
		}, nil
	}

	// Step 1 — Create the sending domain
	domainResp, err := s.createDomain(ctx, req.Domain)
	if err != nil {
		return nil, err
	}

	// Step 2 — Create a scoped API key tied to the new domain
	keyName := fmt.Sprintf("%s-auditsphere", req.Slug)
	keyResp, err := s.createAPIKey(ctx, keyName, domainResp.ID)
	if err != nil {
		// Domain was already created; return a descriptive error so the operator
		// knows to create the API key manually or retry.
		return nil, apperrors.Wrap(
			"RESEND_API_KEY_ERROR",
			fmt.Sprintf("domain '%s' was created (id: %s) but API key generation failed: %v", req.Domain, domainResp.ID, err),
			500, err,
		)
	}

	return &ProvisionDomainResponse{
		DomainID:     domainResp.ID,
		Domain:       domainResp.Name,
		Status:       domainResp.Status,
		DNSRecords:   domainResp.Records,
		ClientAPIKey: keyResp.Token,
	}, nil
}

// createDomain calls POST /domains on the Resend API
func (s *ResendService) createDomain(ctx context.Context, domainName string) (*resendCreateDomainResponse, error) {
	body, _ := json.Marshal(resendCreateDomainRequest{Name: domainName})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendBaseURL+"/domains", bytes.NewBuffer(body))
	if err != nil {
		return nil, apperrors.Wrap("RESEND_REQUEST_BUILD_ERROR", "failed to build domain create request", 500, err)
	}
	req.Header.Set("Authorization", "Bearer "+s.masterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap("RESEND_DOMAIN_REQUEST_ERROR", "failed to call Resend create domain API", 502, err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(
			"RESEND_DOMAIN_API_ERROR",
			fmt.Sprintf("Resend API returned HTTP %d: %s", resp.StatusCode, string(rawBody)),
			resp.StatusCode, nil,
		)
	}

	var result resendCreateDomainResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, apperrors.Wrap("RESEND_DOMAIN_PARSE_ERROR", "failed to parse Resend domain response", 500, err)
	}

	return &result, nil
}

// createAPIKey calls POST /api-keys on the Resend API with domain-scoped sending_access
func (s *ResendService) createAPIKey(ctx context.Context, name, domainID string) (*resendCreateAPIKeyResponse, error) {
	body, _ := json.Marshal(resendCreateAPIKeyRequest{
		Name:       name,
		Permission: "sending_access",
		DomainID:   domainID,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendBaseURL+"/api-keys", bytes.NewBuffer(body))
	if err != nil {
		return nil, apperrors.Wrap("RESEND_REQUEST_BUILD_ERROR", "failed to build API key create request", 500, err)
	}
	req.Header.Set("Authorization", "Bearer "+s.masterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap("RESEND_API_KEY_REQUEST_ERROR", "failed to call Resend create API key endpoint", 502, err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(
			"RESEND_API_KEY_API_ERROR",
			fmt.Sprintf("Resend API returned HTTP %d: %s", resp.StatusCode, string(rawBody)),
			resp.StatusCode, nil,
		)
	}

	var result resendCreateAPIKeyResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, apperrors.Wrap("RESEND_API_KEY_PARSE_ERROR", "failed to parse Resend API key response", 500, err)
	}

	return &result, nil
}
