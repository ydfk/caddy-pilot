package certificate

import (
	"time"

	model "go-fiber-starter/internal/model/certificate"
	"go-fiber-starter/internal/service"

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
	ID                 uuid.UUID                   `json:"id"`
	Name               string                      `json:"name"`
	CertificateType    string                      `json:"certificate_type"`
	Subjects           []string                    `json:"subjects"`
	ChallengeType      string                      `json:"challenge_type"`
	DNSProviderID      *uuid.UUID                  `json:"dns_provider_id,omitempty"`
	Enabled            bool                        `json:"enabled"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at"`
	IssuedCertificates []IssuedCertificateResponse `json:"issued_certificates"`
	IssuanceState      string                      `json:"issuance_state"`
	IssuanceMessage    string                      `json:"issuance_message"`
	LastError          string                      `json:"last_error,omitempty"`
	UsageCount         int                         `json:"usage_count"`
}

type IssuedCertificateResponse struct {
	Subjects     []string  `json:"subjects"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
}

type CertificateProfileOutput struct{ Body CertificateProfileResponse }
type CertificateProfileListOutput struct{ Body []CertificateProfileResponse }

func responseFromModel(profile model.CertificateProfile, subjects []string, issued []service.IssuedCertificate, runtime service.CertificateIssuanceStatus, usageCount int) CertificateProfileResponse {
	response := CertificateProfileResponse{
		ID: profile.Id, Name: profile.Name, CertificateType: profile.CertificateType,
		Subjects: subjects, ChallengeType: profile.ChallengeType, DNSProviderID: profile.DNSProviderID,
		Enabled: profile.Enabled, CreatedAt: profile.CreatedAt, UpdatedAt: profile.UpdatedAt,
		IssuedCertificates: make([]IssuedCertificateResponse, 0, len(issued)),
		IssuanceState:      runtime.State, IssuanceMessage: runtime.Message,
		LastError: runtime.LastError, UsageCount: usageCount,
	}
	for _, certificate := range issued {
		response.IssuedCertificates = append(response.IssuedCertificates, IssuedCertificateResponse{
			Subjects: certificate.Subjects, IssuedAt: certificate.IssuedAt, ExpiresAt: certificate.ExpiresAt,
			Issuer: certificate.Issuer, SerialNumber: certificate.SerialNumber, Status: certificate.Status,
		})
	}
	return response
}
