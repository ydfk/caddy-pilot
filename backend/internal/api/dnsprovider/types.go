package dnsprovider

import (
	"time"

	model "go-fiber-starter/internal/model/dnsprovider"

	"github.com/google/uuid"
)

type DNSProviderCreatePayload struct {
	Name            string `json:"name" minLength:"1" maxLength:"128"`
	ProviderType    string `json:"provider_type" enum:"alidns"`
	AccessKeyID     string `json:"access_key_id" minLength:"1" maxLength:"256"`
	AccessKeySecret string `json:"access_key_secret" minLength:"1" maxLength:"256"`
	RegionID        string `json:"region_id,omitempty" maxLength:"64"`
	Enabled         bool   `json:"enabled"`
}

type DNSProviderUpdatePayload struct {
	Name            string `json:"name" minLength:"1" maxLength:"128"`
	AccessKeyID     string `json:"access_key_id" minLength:"1" maxLength:"256"`
	AccessKeySecret string `json:"access_key_secret,omitempty" maxLength:"256" doc:"留空时保留原密钥"`
	RegionID        string `json:"region_id,omitempty" maxLength:"64"`
	Enabled         bool   `json:"enabled"`
}

type DNSProviderCreateInput struct{ Body DNSProviderCreatePayload }
type DNSProviderIDInput struct {
	ID string `path:"id" format:"uuid"`
}
type DNSProviderUpdateInput struct {
	ID   string `path:"id" format:"uuid"`
	Body DNSProviderUpdatePayload
}

type DNSProviderResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	ProviderType    string    `json:"provider_type"`
	AccessKeyIDHint string    `json:"access_key_id_hint"`
	RegionID        string    `json:"region_id"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DNSProviderOutput struct{ Body DNSProviderResponse }
type DNSProviderListOutput struct{ Body []DNSProviderResponse }

func responseFromModel(provider model.DNSProvider, accessKeyID, regionID string) DNSProviderResponse {
	return DNSProviderResponse{
		ID: provider.Id, Name: provider.Name, ProviderType: provider.ProviderType,
		AccessKeyIDHint: maskValue(accessKeyID), RegionID: regionID, Enabled: provider.Enabled,
		CreatedAt: provider.CreatedAt, UpdatedAt: provider.UpdatedAt,
	}
}

func maskValue(value string) string {
	if len(value) <= 6 {
		return "******"
	}
	return value[:3] + "***" + value[len(value)-3:]
}
