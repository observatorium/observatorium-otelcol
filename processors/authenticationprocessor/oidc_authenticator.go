// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authenticationprocessor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/collector/component"
)

type subject struct{}
type groups struct{}
type oidcAuthenticator struct {
	config   OIDC
	logger   *zap.Logger
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

var (
	_ authenticator = (*oidcAuthenticator)(nil)

	// Subject is the context key holding the Subject from this request
	Subject = subject{}

	// Groups is the context key holding the groups the subject from this request belongs to
	Groups = groups{}

	errInvalidAuthenticationHeaderFormat = errors.New("invalid authorization header format")
	errFailedToObtainClaimsFromToken     = errors.New("failed to get the subject from the token issued by the OIDC provider")
	errClaimNotFound                     = errors.New("username claim from the OIDC configuration not found on the token returned by the OIDC provider")
	errUsernameNotString                 = errors.New("the username returned by the OIDC provider isn't a regular string")
	errGroupsClaimNotFound               = errors.New("groups claim from the OIDC configuration not found on the token returned by the OIDC provider")
)

func newOIDCAuthenticator(logger *zap.Logger, config OIDC) *oidcAuthenticator {
	return &oidcAuthenticator{
		logger: logger.With(zap.String("authenticator", "oidc")),
		config: config,
	}
}

func (o *oidcAuthenticator) authenticate(ctx context.Context) (bool, error) {
	// TODO: check also if we can get the HTTP headers, in case the gRPC one yields no token
	md, _ := metadata.FromIncomingContext(ctx)
	headers := md.Get("authorization")
	o.logger.Debug("received trace")

	authenticated := false
	for _, header := range headers {
		parts := strings.Split(header, " ")

		if len(parts) != 2 {
			return false, errInvalidAuthenticationHeaderFormat
		}

		idToken, err := o.verifier.Verify(ctx, parts[1])
		if err != nil {
			o.logger.Debug("failed to verify token", zap.Error(err))
			return false, err
		}

		claims := map[string]interface{}{}
		if err = idToken.Claims(&claims); err != nil {
			// currently, this isn't a valid condition, the Verify call a few lines above
			// will already attempt to parse the payload as a json and set it as the claims
			// for the token. As we are using a map to hold the claims, there's no way to fail
			// to read the claims. It could fail if we were using a custom struct. Instead of
			// swalling the error, it's better to make this future-proof, in case the underlying
			// code changes
			return false, errFailedToObtainClaimsFromToken
		}

		subject, err := getSubjectFromClaims(claims, o.config.UsernameClaim, idToken.Subject)
		if err != nil {
			return false, err
		}
		ctx = context.WithValue(ctx, Subject, subject)

		groups, err := getGroupsFromClaims(claims, o.config.GroupsClaim)
		if err != nil {
			return false, err
		}
		ctx = context.WithValue(ctx, Groups, groups)

		o.logger.Debug("authentication succeeded for batch")
		authenticated = true
	}

	return authenticated, nil
}

func (o *oidcAuthenticator) start(context.Context, component.Host) error {
	provider, err := getProviderForConfig(o.config)
	if err != nil {
		return err
	}
	o.provider = provider

	o.verifier = o.provider.Verifier(&oidc.Config{
		ClientID: o.config.Audience,
	})

	return nil
}

func (o *oidcAuthenticator) shutdown(context.Context) error {
	return nil
}

func getSubjectFromClaims(claims map[string]interface{}, usernameClaim string, fallback string) (string, error) {
	if len(usernameClaim) > 0 {
		username, found := claims[usernameClaim]
		if !found {
			return "", errClaimNotFound
		}

		sUsername, ok := username.(string)
		if !ok {
			return "", errUsernameNotString
		}

		return sUsername, nil
	}

	return fallback, nil
}

func getGroupsFromClaims(claims map[string]interface{}, groupsClaim string) ([]string, error) {
	if len(groupsClaim) > 0 {
		var groups []string
		rawGroup, ok := claims[groupsClaim]
		if !ok {
			return nil, errGroupsClaimNotFound
		}
		switch v := rawGroup.(type) {
		case string:
			groups = append(groups, v)
		case []string:
			groups = v
		case []interface{}:
			groups = make([]string, 0, len(v))
			for i := range v {
				groups = append(groups, fmt.Sprintf("%v", v[i]))
			}
		}

		return groups, nil
	}

	return []string{}, nil
}

func getProviderForConfig(config OIDC) (*oidc.Provider, error) {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cert, err := getIssuerCACertFromPath(config.IssuerCAPath)
	if err != nil {
		return nil, err // the errors from this path have enough context already
	}

	if cert != nil {
		t.TLSClientConfig = &tls.Config{
			RootCAs: x509.NewCertPool(),
		}
		t.TLSClientConfig.RootCAs.AddCert(cert)
	}

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: t,
	}
	oidcContext := oidc.ClientContext(context.Background(), client)
	return oidc.NewProvider(oidcContext, config.IssuerURL)
}

func getIssuerCACertFromPath(path string) (*x509.Certificate, error) {
	if path == "" {
		return nil, nil
	}

	rawCA, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read the CA file %q: %w", path, err)
	}

	if len(rawCA) == 0 {
		return nil, fmt.Errorf("could not read the CA file %q: empty file", path)
	}

	block, _ := pem.Decode(rawCA)
	if block == nil {
		return nil, fmt.Errorf("cannot decode the contents of the CA file %q: %w", path, err)
	}

	return x509.ParseCertificate(block.Bytes)
}
