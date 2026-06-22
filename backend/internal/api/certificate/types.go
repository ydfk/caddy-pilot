package certificate

import (
	"time"

	model "go-fiber-starter/internal/model/certificate"

	"github.com/google/uuid"
)

type CertificateProfilePayload struct {
	Name            string     `json:"name" minLength:"1" maxLength:"128"`
	CertificateType string     `json:"certificate_type" enum:"single,wildcard"`
	Subjects        []string   `json:"subjects" minItems:"1"`
	ChallengeType   string     `json:"challenge_type" enum:"http,dns"`
	DNSProviderID   *uuid.UUID `json:"dns_provider_id,omitempty"`
	Enabled         bool       `json:"enabled"`
}

type CertificateProfileCreateInput struct{ Body CertificateProfilePayload }
type CertificateProfileIDInput struct {
	ID string `path:"id" format:"uuid"`
}
type CertificateProfileUpdateInput struct {
	ID   string `path:"id" format:"uuid"`
	Body CertificateProfilePayload
}

type CertificateProfileResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	CertificateType string     `json:"certificate_type"`
	Subjects        []string   `json:"subjects"`
	ChallengeType   string     `json:"challenge_type"`
	DNSProviderID   *uuid.UUID `json:"dns_provider_id,omitempty"`
	Enabled         bool       `json:"enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CertificateProfileOutput struct{ Body CertificateProfileResponse }
type CertificateProfileListOutput struct{ Body []CertificateProfileResponse }

func responseFromModel(profile model.CertificateProfile, subjects []string) CertificateProfileResponse {
	return CertificateProfileResponse{
		ID: profile.Id, Name: profile.Name, CertificateType: profile.CertificateType,
		Subjects: subjects, ChallengeType: profile.ChallengeType, DNSProviderID: profile.DNSProviderID,
		Enabled: profile.Enabled, CreatedAt: profile.CreatedAt, UpdatedAt: profile.UpdatedAt,
	}
}
