package riskclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// PofElement is one fulfilment element of the COSO PoF catalogue
// (risk-service GET /api/v1/rcm/coso-pof-catalogue).
type PofElement struct {
	Code                        string   `json:"code"`
	Index                       int      `json:"index"`
	Type                        string   `json:"type"`
	DescriptionEN               string   `json:"description_en"`
	DescriptionID               string   `json:"description_id"`
	Weight                      float64  `json:"weight"`
	DefaultMethods              []string `json:"default_methods"`
	ScaleType                   string   `json:"scale_type"`
	RequiresIndependentAssessor bool     `json:"requires_independent_assessor"`
}

// Pof is one point of focus
type Pof struct {
	Code              string       `json:"code"`
	Index             int          `json:"index"`
	NameEN            string       `json:"name_en"`
	NameID            string       `json:"name_id"`
	Weight            float64      `json:"weight"`
	DefaultOwnerRoles []string     `json:"default_owner_roles"`
	Elements          []PofElement `json:"elements"`
}

// PofPrinciple is one principle with its PoFs
type PofPrinciple struct {
	Code              string   `json:"code"`
	Number            int      `json:"number"`
	Index             int      `json:"index"`
	NameEN            string   `json:"name_en"`
	NameID            string   `json:"name_id"`
	Weight            float64  `json:"weight"`
	DefaultOwnerRoles []string `json:"default_owner_roles"`
	Pofs              []Pof    `json:"pofs"`
}

// PofComponent is one component
type PofComponent struct {
	Code       string         `json:"code"`
	Number     int            `json:"number"`
	NameEN     string         `json:"name_en"`
	NameID     string         `json:"name_id"`
	Weight     float64        `json:"weight"`
	Principles []PofPrinciple `json:"principles"`
}

// PofCatalogue is one catalogue version
type PofCatalogue struct {
	Framework  string         `json:"framework"`
	Version    string         `json:"version"`
	Name       string         `json:"name"`
	IsCurrent  bool           `json:"is_current"`
	Components []PofComponent `json:"components"`
}

// GetPofCatalogue reads one catalogue version (version "" = the current one)
// with the caller's token. Any non-200 answer is an error.
func (c *RiskClient) GetPofCatalogue(ctx context.Context, authorization, version string) (*PofCatalogue, error) {
	v := url.Values{}
	if strings.TrimSpace(version) != "" {
		v.Set("version", strings.TrimSpace(version))
	}
	path := "/api/v1/rcm/coso-pof-catalogue"
	if len(v) > 0 {
		path += "?" + v.Encode()
	}
	data, err := c.getData(ctx, authorization, path)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Catalogue PofCatalogue `json:"catalogue"`
	}
	if err := json.Unmarshal(data, &wrap); err != nil {
		return nil, fmt.Errorf("decode coso pof catalogue: %w", err)
	}
	if len(wrap.Catalogue.Components) == 0 {
		return nil, fmt.Errorf("coso pof catalogue is empty")
	}
	return &wrap.Catalogue, nil
}
