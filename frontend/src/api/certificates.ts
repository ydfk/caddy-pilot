import { apiRequest } from "./client";

export type CertificateProfile = {
  id: string;
  name: string;
  certificate_type: "single" | "wildcard";
  subjects: string[];
  challenge_type: "http" | "dns";
  dns_provider_id?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
  issued_certificates: IssuedCertificate[];
};

export type IssuedCertificate = {
  subjects: string[];
  issued_at: string;
  expires_at: string;
  issuer: string;
  serial_number: string;
  status: "valid" | "expiring" | "expired";
};

export type CertificateProfilePayload = Omit<
  CertificateProfile,
  "id" | "created_at" | "updated_at" | "issued_certificates"
>;

export const listCertificates = () => apiRequest<CertificateProfile[]>("/api/certificates");
export const createCertificate = (payload: CertificateProfilePayload) =>
  apiRequest<CertificateProfile>("/api/certificates", {
    method: "POST",
    body: JSON.stringify(payload),
  });
export const updateCertificate = (id: string, payload: CertificateProfilePayload) =>
  apiRequest<CertificateProfile>(`/api/certificates/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
export const deleteCertificate = (id: string) =>
  apiRequest<void>(`/api/certificates/${id}`, { method: "DELETE" });
