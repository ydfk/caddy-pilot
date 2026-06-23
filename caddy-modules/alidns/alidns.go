package alidns

import (
	"context"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	libdnsalidns "github.com/libdns/alidns"
	"github.com/libdns/libdns"
	"go.uber.org/zap"
)

type Provider struct {
	*libdnsalidns.Provider
	providerID string
	logger     *zap.Logger
}

func init() {
	caddy.RegisterModule(Provider{})
}

func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.alidns",
		New: func() caddy.Module { return &Provider{Provider: new(libdnsalidns.Provider)} },
	}
}

func (p *Provider) Provision(ctx caddy.Context) error {
	replacer := caddy.NewReplacer()
	p.providerID = providerIDFromEnvPlaceholder(p.AccessKeyID)
	p.AccessKeyID = replacer.ReplaceAll(p.AccessKeyID, "")
	p.AccessKeySecret = replacer.ReplaceAll(p.AccessKeySecret, "")
	p.SecurityToken = replacer.ReplaceAll(p.SecurityToken, "")
	p.logger = ctx.Logger(p).Named("dns_provider_audit")
	return nil
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.audit(ctx, "append", zone, records, p.Provider.AppendRecords)
}

func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.audit(ctx, "set", zone, records, p.Provider.SetRecords)
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.audit(ctx, "delete", zone, records, p.Provider.DeleteRecords)
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	started := time.Now()
	records, err := p.Provider.GetRecords(ctx, zone)
	p.log("get", zone, nil, records, started, err)
	return records, err
}

func (p *Provider) audit(
	ctx context.Context,
	operation string,
	zone string,
	records []libdns.Record,
	call func(context.Context, string, []libdns.Record) ([]libdns.Record, error),
) ([]libdns.Record, error) {
	started := time.Now()
	result, err := call(ctx, zone, records)
	p.log(operation, zone, records, result, started, err)
	return result, err
}

func (p *Provider) log(operation, zone string, input, result []libdns.Record, started time.Time, err error) {
	if p.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.String("provider", "alidns"),
		zap.String("provider_id", p.providerID),
		zap.String("zone", zone),
		zap.String("operation", operation),
		zap.Any("records", auditRecords(input)),
		zap.Any("result", auditRecords(result)),
		zap.Duration("duration", time.Since(started)),
	}
	if err != nil {
		p.logger.Error("dns_provider_call", append(fields, zap.Error(err))...)
		return
	}
	p.logger.Info("dns_provider_call", fields...)
}

type auditRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   string `json:"ttl,omitempty"`
}

func auditRecords(records []libdns.Record) []auditRecord {
	result := make([]auditRecord, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		result = append(result, auditRecord{Name: rr.Name, Type: rr.Type, Value: rr.Data, TTL: rr.TTL.String()})
	}
	return result
}

func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			if err := p.unmarshalOption(d); err != nil {
				return err
			}
		}
	}
	if p.AccessKeyID == "" || p.AccessKeySecret == "" {
		return d.Err("AccessKeyID or AccessKeySecret is empty")
	}
	return nil
}

func (p *Provider) unmarshalOption(d *caddyfile.Dispenser) error {
	directive := d.Val()
	if !d.NextArg() {
		return d.ArgErr()
	}
	value := d.Val()
	if d.NextArg() {
		return d.ArgErr()
	}
	switch directive {
	case "access_key_id":
		p.AccessKeyID = value
	case "access_key_secret":
		p.AccessKeySecret = value
	case "region_id":
		p.RegionID = value
	case "security_token":
		p.SecurityToken = value
	default:
		return d.Errf("unrecognized subdirective '%s'", directive)
	}
	return nil
}

func providerIDFromEnvPlaceholder(value string) string {
	const prefix = "{env.CADDYPILOT_DNS_"
	const suffix = "_ACCESS_KEY_ID}"
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, suffix) {
		return ""
	}
	id := strings.TrimSuffix(strings.TrimPrefix(value, prefix), suffix)
	return strings.ToLower(strings.ReplaceAll(id, "_", "-"))
}

var (
	_ caddy.Provisioner     = (*Provider)(nil)
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
