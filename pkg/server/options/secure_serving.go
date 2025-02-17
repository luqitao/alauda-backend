/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	"gomod.alauda.cn/alauda-backend/pkg/certwatcher"
	"gomod.alauda.cn/alauda-backend/pkg/server"
	"gomod.alauda.cn/log"

	"github.com/spf13/pflag"
)

var versions = map[string]uint16{
	"VersionTLS10": tls.VersionTLS10,
	"VersionTLS11": tls.VersionTLS11,
	"VersionTLS12": tls.VersionTLS12,
	"VersionTLS13": tls.VersionTLS13,
}

// SecureServingOptions are for creating an unauthenticated, authorized, secure port.
type SecureServingOptions struct {
	BindAddress net.IP
	// BindPort is ignored when Listener is set, will serve https even with 0.
	BindPort int
	// BindNetwork is the type of network to bind to - defaults to "tcp", accepts "tcp",
	// "tcp4", and "tcp6".
	BindNetwork string
	// Required set to true means that BindPort cannot be zero.
	Required bool

	// Listener is the secure server network listener.
	// either Listener or BindAddress/BindPort/BindNetwork is set,
	// if Listener is set, use it and omit BindAddress/BindPort/BindNetwork.
	Listener net.Listener

	// CertFile is the server certificate file. Defaults to TLS.crt.
	CertFile string

	// KeyFile is the server key file. Defaults to TLS.key.
	KeyFile string

	// ClientCAFile is the CA certificate name which server used to verify remote(client)'s certificate.
	// Defaults to "", which means server does not verify client's certificate.
	ClientCAFile string

	// TLSVersion is the minimum version of TLS supported. Accepts
	// "", "1.0", "1.1", "1.2" and "1.3" only ("" is equivalent to "1.0" for backwards compatibility)
	TLSMinVersion string
}

// NewSecureServingOptions construct serving options to Server
func NewSecureServingOptions() *SecureServingOptions {
	return &SecureServingOptions{
		BindAddress:   net.IPv4zero,
		BindPort:      443,
		BindNetwork:   "tcp",
		TLSMinVersion: "VersionTLS10",
		Required:      false,
	}
}

// AddFlags adds flags related to insecure serving to the specified FlagSet.
func (s *SecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}

	fs.IPVar(&s.BindAddress, "secure-bind-address", s.BindAddress, ""+
		"The IP address on which to listen for the --secure-port port. The "+
		"associated interface(s) must be reachable by the rest of the cluster, and by CLI/web "+
		"clients. If blank or an unspecified address (0.0.0.0 or ::), all interfaces will be used.")

	desc := "The port on which to serve HTTPS with authentication and authorization."
	if s.Required {
		desc += " It cannot be switched off with 0."
	} else {
		desc += " If 0, don't serve HTTPS at all."
	}
	fs.IntVar(&s.BindPort, "secure-port", s.BindPort, desc)

	fs.StringVar(&s.CertFile, "tls-cert-file", s.CertFile, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert). If HTTPS serving is enabled, --tls-cert-file and "+
		"--tls-private-key-file should be provided.")

	fs.StringVar(&s.KeyFile, "tls-private-key-file", s.KeyFile,
		"File containing the default x509 private key matching --tls-cert-file.")

	fs.StringVar(&s.ClientCAFile, "client-ca-file", s.ClientCAFile,
		"File containing the client certificate. matching --tls-cert-file. If set, the server will "+
			"verify client's certificate for any request.")

	tlsPossibleVersions := []string{"VersionTLS10", "VersionTLS11", "VersionTLS12", "VersionTLS13"}
	fs.StringVar(&s.TLSMinVersion, "tls-min-version", s.TLSMinVersion,
		"Minimum TLS version supported. "+
			"Possible values: "+strings.Join(tlsPossibleVersions, ", "))
}

// ApplyFlags apply flags
func (s *SecureServingOptions) ApplyFlags() []error {
	var errs []error

	// Already apply flags in AddFlags
	return errs
}

// ApplyToServer adds SecureServingOptions to the insecureserverinfo and kube-controller manager configuration.
func (s *SecureServingOptions) ApplyToServer(srv server.Server) (err error) {
	if s == nil {
		return nil
	}
	if s.BindPort <= 0 {
		return nil
	}

	if s.Listener == nil {
		var err error

		cw, err := certwatcher.New(s.CertFile, s.KeyFile)
		if err != nil {
			return err
		}
		go func() {
			if err := cw.Start(context.Background()); err != nil {
				log.Error("certificate watcher error, " + err.Error())
			}
		}()

		cfg := &tls.Config{ //nolint:gosec
			NextProtos:     []string{"h2"},
			GetCertificate: cw.GetCertificate,
			MinVersion:     versions[s.TLSMinVersion],
		}

		// load CA to verify client certificate
		if s.ClientCAFile != "" {
			certPool := x509.NewCertPool()
			clientCABytes, err := ioutil.ReadFile(s.ClientCAFile)
			if err != nil {
				return fmt.Errorf("failed to read client CA cert: %v", err)
			}

			ok := certPool.AppendCertsFromPEM(clientCABytes)
			if !ok {
				return fmt.Errorf("failed to append client CA cert to CA pool")
			}

			cfg.ClientCAs = certPool
			cfg.ClientAuth = tls.RequireAndVerifyClientCert
		}

		s.Listener, err = tls.Listen(s.BindNetwork, net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort)), cfg)
		if err != nil {
			return fmt.Errorf("failed to create listener: %v", err)
		}
	}

	if setter := srv.(server.ListenerSetter); setter != nil {
		setter.SetListener(s.Listener, s.BindPort)
	}

	return nil
}
