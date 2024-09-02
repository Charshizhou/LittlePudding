package auth

import (
	"crypto/tls"
	"fmt"
	"google.golang.org/grpc/credentials"
)

type Certificate struct {
	CertFile   string
	KeyFile    string
	ServerName string
}

func (c Certificate) GetTLSConfigForServer() (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(
		c.CertFile,
		c.KeyFile,
	)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read client ca cert: %s", err)
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    nil,
	}

	return tlsConfig, nil
}

func (c Certificate) GetTransportCredsForClient() (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(
		c.CertFile,
		c.KeyFile,
	)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert: %s", err)
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:         c.ServerName,
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: true,
		RootCAs:            nil,
	})

	return transportCreds, nil
}
