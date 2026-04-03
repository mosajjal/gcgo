package compute

import (
	"context"
	"errors"
	"fmt"
	"os"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

func (c *gcpClient) ListSSLCertificates(ctx context.Context, project string) ([]*SSLCertificate, error) {
	it := c.sslCertificates.List(ctx, &computepb.ListSslCertificatesRequest{
		Project: project,
	})
	var out []*SSLCertificate
	for {
		cert, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list ssl certificates: %w", err)
		}
		out = append(out, sslCertFromProto(cert))
	}
	return out, nil
}

func (c *gcpClient) GetSSLCertificate(ctx context.Context, project, name string) (*SSLCertificate, error) {
	cert, err := c.sslCertificates.Get(ctx, &computepb.GetSslCertificateRequest{
		Project:        project,
		SslCertificate: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get ssl certificate %s: %w", name, err)
	}
	return sslCertFromProto(cert), nil
}

func (c *gcpClient) CreateSSLCertificate(ctx context.Context, project string, req *CreateSSLCertificateRequest) error {
	res := &computepb.SslCertificate{
		Name:        &req.Name,
		Description: strPtrOrNil(req.Description),
	}

	if len(req.Domains) > 0 {
		// managed certificate
		certType := "MANAGED"
		res.Type = &certType
		res.Managed = &computepb.SslCertificateManagedSslCertificate{
			Domains: req.Domains,
		}
	} else {
		// self-managed certificate
		certType := "SELF_MANAGED"
		res.Type = &certType
		certPEM, err := os.ReadFile(req.CertFile)
		if err != nil {
			return fmt.Errorf("read cert file %s: %w", req.CertFile, err)
		}
		keyPEM, err := os.ReadFile(req.KeyFile)
		if err != nil {
			return fmt.Errorf("read key file %s: %w", req.KeyFile, err)
		}
		certStr := string(certPEM)
		keyStr := string(keyPEM)
		res.SelfManaged = &computepb.SslCertificateSelfManagedSslCertificate{
			Certificate: &certStr,
			PrivateKey:  &keyStr,
		}
	}

	op, err := c.sslCertificates.Insert(ctx, &computepb.InsertSslCertificateRequest{
		Project:               project,
		SslCertificateResource: res,
	})
	if err != nil {
		return fmt.Errorf("create ssl certificate %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteSSLCertificate(ctx context.Context, project, name string) error {
	op, err := c.sslCertificates.Delete(ctx, &computepb.DeleteSslCertificateRequest{
		Project:        project,
		SslCertificate: name,
	})
	if err != nil {
		return fmt.Errorf("delete ssl certificate %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func sslCertFromProto(cert *computepb.SslCertificate) *SSLCertificate {
	out := &SSLCertificate{
		Name:        cert.GetName(),
		Type:        cert.GetType(),
		Description: cert.GetDescription(),
	}
	if m := cert.GetManaged(); m != nil {
		out.Domains = m.GetDomains()
		out.Status = m.GetStatus()
	}
	return out
}
