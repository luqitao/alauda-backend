/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"gomod.alauda.cn/alauda-backend/pkg/decorator"
	"gomod.alauda.cn/alauda-backend/pkg/server"
)

const (
	flagBindAddress = "insecure-bind-address"
	flagBindPort    = "insecure-port"
	flagHealthCheck = "health-check"
	flagPathPrefix  = "path-prefix"

	configBindAddress = "server.insecure_bind_address"
	configBindPort    = "server.insecure_port"
	configHealthCheck = "server.health_check"
	configPathPrefix  = "server.path_prefix"
)

// InsecureServingOptions are for creating an unauthenticated, unauthorized, insecure port.
type InsecureServingOptions struct {
	BindAddress net.IP
	BindPort    int
	HealthCheck bool
	PathPrefix  string

	// Listener is the secure server network listener.
	// either Listener or BindAddress/BindPort/BindNetwork is set,
	// if Listener is set, use it and omit BindAddress/BindPort/BindNetwork.
	Listener net.Listener

	// ListenFunc can be overridden to create a custom listener, e.g. for mocking in tests.
	// It defaults to options.CreateListener.
	ListenFunc func(network, addr string) (net.Listener, int, error)
}

// NewInsecureServingOptions construct serving options to Server
func NewInsecureServingOptions() *InsecureServingOptions {
	return &InsecureServingOptions{
		BindAddress: net.IPv4zero,
		BindPort:    8080,
		HealthCheck: true,
	}
}

// Validate ensures that the insecure port values within the range of the port.
func (s *InsecureServingOptions) Validate() []error {
	if s == nil {
		return nil
	}

	errors := []error{}

	if s.BindPort < 0 || s.BindPort > 65335 {
		errors = append(errors, fmt.Errorf("insecure port %v must be between 0 and 65335, inclusive. 0 for turning off insecure (HTTP) port", s.BindPort))
	}

	return errors
}

// AddFlags adds flags related to insecure serving to the specified FlagSet.
func (s *InsecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	if s == nil {
		return
	}

	fs.IP(flagBindAddress, s.BindAddress, ""+
		"The IP address on which to serve the --insecure-port (set to 0.0.0.0 for all IPv4 interfaces and :: for all IPv6 interfaces).")
	_ = viper.BindPFlag(configBindAddress, fs.Lookup(flagBindAddress))

	fs.Int(flagBindPort, s.BindPort, ""+
		"The port on which to serve unsecured, unauthenticated access.")
	_ = viper.BindPFlag(configBindPort, fs.Lookup(flagBindPort))

	fs.Bool(flagHealthCheck, s.HealthCheck, ""+
		"Enables health check endpoint /healthz on server.")
	_ = viper.BindPFlag(configHealthCheck, fs.Lookup(flagHealthCheck))

	fs.String(flagPathPrefix, s.PathPrefix, ""+
		"Sets a path prefix as instruction for other services")
	_ = viper.BindPFlag(configPathPrefix, fs.Lookup(flagPathPrefix))
}

// ApplyFlags apply flags
func (s *InsecureServingOptions) ApplyFlags() []error {
	var errs []error

	s.BindAddress = net.ParseIP(viper.GetString(configBindAddress))
	s.BindPort = viper.GetInt(configBindPort)
	s.HealthCheck = viper.GetBool(configHealthCheck)
	s.PathPrefix = viper.GetString(configPathPrefix)

	return errs
}

// ApplyToServer adds InsecureServingOptions to the insecureserverinfo and kube-controller manager configuration.
// Note: the double pointer allows to set the *DeprecatedInsecureServingInfo to nil without referencing the struct hosting this pointer.
func (s *InsecureServingOptions) ApplyToServer(svr server.Server) error {
	if s == nil {
		return nil
	}
	if s.BindPort <= 0 {
		return nil
	}

	if s.Listener == nil {
		var err error
		listen := CreateListener
		if s.ListenFunc != nil {
			listen = s.ListenFunc
		}
		addr := net.JoinHostPort(s.BindAddress.String(), fmt.Sprintf("%d", s.BindPort))
		s.Listener, s.BindPort, err = listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to create listener: %v", err)
		}
	}

	if setter := svr.(server.ListenerSetter); setter != nil {
		setter.SetListener(s.Listener, s.BindPort)
	}
	if s.HealthCheck {
		svr.Container().Handle("/healthz", healthz{})
		svr.Container().Handle("/_ping/", healthz{})
		svr.Container().Handle("/_ping", healthz{})
	}
	if s.PathPrefix != "" {
		svr.Container().Router(decorator.NewRewriteRouter(s.PathPrefix, svr.L()))
	}
	return nil
}

// CreateListener create a new listener
func CreateListener(network, addr string) (net.Listener, int, error) {
	if len(network) == 0 {
		network = "tcp"
	}
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen on %v: %v", addr, err)
	}

	// get port
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		ln.Close()
		return nil, 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}

	return ln, tcpAddr.Port, nil
}

type healthz struct{}

func (healthz) ServeHTTP(rw http.ResponseWriter, httpRequest *http.Request) {
	io.WriteString(rw, "ok")
}
