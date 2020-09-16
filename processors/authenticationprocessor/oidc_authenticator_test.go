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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestOIDCAuthenticationSucceeded(t *testing.T) {
	// prepare
	oidcServer, err := newOIDCServer(logger)
	require.NoError(t, err)
	oidcServer.Start()
	defer oidcServer.Close()

	config := OIDC{
		IssuerURL: oidcServer.URL,
		Audience:  "unit-test",
	}

	p := newOIDCAuthenticator(logger, config)

	err = p.start(context.Background(), nil)
	require.NoError(t, err)
	defer p.shutdown(context.Background())

	payload, _ := json.Marshal(map[string]interface{}{
		"sub":  "jdoe@example.com",
		"name": "jdoe",
		"iss":  oidcServer.URL,
		"aud":  "unit-test",
		"exp":  time.Now().Add(time.Minute).Unix(),
	})
	token, err := oidcServer.token(payload)
	require.NoError(t, err)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token)))

	// test
	authenticated, err := p.authenticate(ctx)

	// verify
	assert.NoError(t, err)
	assert.True(t, authenticated)
}

func TestOIDCProviderForConfigWithTLS(t *testing.T) {
	// prepare the CA cert for the TLS handler
	cert := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * time.Second),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		SerialNumber: big.NewInt(9447457), // some number
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	x509Cert, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &priv.PublicKey, priv)
	require.NoError(t, err)

	caFile, err := ioutil.TempFile(os.TempDir(), "cert")
	require.NoError(t, err)
	defer os.Remove(caFile.Name())

	err = pem.Encode(caFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: x509Cert,
	})
	require.NoError(t, err)

	oidcServer, err := newOIDCServer(logger)
	require.NoError(t, err)
	defer oidcServer.Close()

	tlsCert := tls.Certificate{
		Certificate: [][]byte{x509Cert},
		PrivateKey:  priv,
	}
	oidcServer.TLS = &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	oidcServer.StartTLS()

	// prepare the processor configuration
	config := OIDC{
		IssuerURL:    oidcServer.URL,
		IssuerCAPath: caFile.Name(),
		Audience:     "unit-test",
	}

	// test
	provider, err := getProviderForConfig(config)

	// verify
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestOIDCLoadIssuerCAFromPath(t *testing.T) {
	// prepare
	cert := x509.Certificate{
		SerialNumber: big.NewInt(9447457), // some number
		IsCA:         true,
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	x509Cert, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &priv.PublicKey, priv)
	require.NoError(t, err)

	file, err := ioutil.TempFile(os.TempDir(), "cert")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	err = pem.Encode(file, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: x509Cert,
	})
	require.NoError(t, err)

	// test
	loaded, err := getIssuerCACertFromPath(file.Name())

	// verify
	assert.NoError(t, err)
	assert.Equal(t, cert.SerialNumber, loaded.SerialNumber)
}

func TestOIDCFailedToLoadIssuerCAFromPathEmptyCert(t *testing.T) {
	// prepare
	file, err := ioutil.TempFile(os.TempDir(), "cert")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	// test
	loaded, err := getIssuerCACertFromPath(file.Name()) // the file exists, but the contents isn't a cert

	// verify
	assert.Error(t, err)
	assert.Nil(t, loaded)
}

func TestOIDCFailedToLoadIssuerCAFromPathMissingFile(t *testing.T) {
	// test
	loaded, err := getIssuerCACertFromPath("some-non-existing-file")

	// verify
	assert.Error(t, err)
	assert.Nil(t, loaded)
}

func TestOIDCFailedToLoadIssuerCAFromPathInvalidContent(t *testing.T) {
	// prepare
	file, err := ioutil.TempFile(os.TempDir(), "cert")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	file.Write([]byte("foobar"))

	config := OIDC{
		IssuerCAPath: file.Name(),
	}

	// test
	provider, err := getProviderForConfig(config) // cross test with getIssuerCACertFromPath

	// verify
	assert.Error(t, err)
	assert.Nil(t, provider)
}

func TestOIDCInvalidAuthHeader(t *testing.T) {
	// prepare
	p := newOIDCAuthenticator(logger, OIDC{})

	// test
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "some-value"))
	authenticated, err := p.authenticate(ctx)

	// verify
	assert.Equal(t, errInvalidAuthenticationHeaderFormat, err)
	assert.False(t, authenticated)
}

func TestOIDCNotAuthenticated(t *testing.T) {
	// prepare
	p := newOIDCAuthenticator(logger, OIDC{})

	// test
	authenticated, err := p.authenticate(context.Background())

	// verify
	assert.NoError(t, err)
	assert.False(t, authenticated)
}

func TestProviderNotReacheable(t *testing.T) {
	// prepare
	config := OIDC{
		IssuerURL: "http://example.com/",
	}
	p := newOIDCAuthenticator(logger, config)

	// test
	err := p.start(context.Background(), nil)

	// verify
	assert.Error(t, err)
}

func TestFailedToVerifyToken(t *testing.T) {
	// prepare
	oidcServer, err := newOIDCServer(logger)
	require.NoError(t, err)
	oidcServer.Start()
	defer oidcServer.Close()

	config := OIDC{
		IssuerURL: oidcServer.URL,
		Audience:  "unit-test",
	}

	p := newOIDCAuthenticator(logger, config)

	err = p.start(context.Background(), nil)
	require.NoError(t, err)
	defer p.shutdown(context.Background())

	// test
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer some-token"))
	authenticated, err := p.authenticate(ctx)

	// verify
	assert.Error(t, err)
	assert.False(t, authenticated)
}

func TestFailedToGetGroupsClaimFromToken(t *testing.T) {
	// prepare
	oidcServer, err := newOIDCServer(logger)
	require.NoError(t, err)
	oidcServer.Start()
	defer oidcServer.Close()

	for _, tt := range []struct {
		casename      string
		config        OIDC
		expectedError error
	}{
		{
			"groupsClaimNonExisting",
			OIDC{
				IssuerURL:   oidcServer.URL,
				Audience:    "unit-test",
				GroupsClaim: "non-existing-claim",
			},
			errGroupsClaimNotFound,
		},
		{
			"usernameClaimNonExisting",
			OIDC{
				IssuerURL:     oidcServer.URL,
				Audience:      "unit-test",
				UsernameClaim: "non-existing-claim",
			},
			errClaimNotFound,
		},
		{
			"usernameNotString",
			OIDC{
				IssuerURL:     oidcServer.URL,
				Audience:      "unit-test",
				UsernameClaim: "some-non-string-field",
			},
			errUsernameNotString,
		},
	} {
		t.Run(tt.casename, func(t *testing.T) {
			p := newOIDCAuthenticator(logger, tt.config)

			err = p.start(context.Background(), nil)
			require.NoError(t, err)
			defer p.shutdown(context.Background())

			payload, _ := json.Marshal(map[string]interface{}{
				"iss":                   oidcServer.URL,
				"some-non-string-field": 123,
				"aud":                   "unit-test",
				"exp":                   time.Now().Add(time.Minute).Unix(),
			})
			token, err := oidcServer.token(payload)
			require.NoError(t, err)
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token)))

			// test
			authenticated, err := p.authenticate(ctx)

			// verify
			assert.Equal(t, tt.expectedError, err)
			assert.False(t, authenticated)
		})
	}
}

func TestSubjectFromClaims(t *testing.T) {
	// prepare
	claims := map[string]interface{}{
		"username": "jdoe",
	}

	// test
	username, err := getSubjectFromClaims(claims, "username", "")

	// verify
	assert.NoError(t, err)
	assert.Equal(t, "jdoe", username)
}

func TestSubjectFallback(t *testing.T) {
	// prepare
	claims := map[string]interface{}{
		"sub": "jdoe",
	}

	// test
	username, err := getSubjectFromClaims(claims, "", "jdoe")

	// verify
	assert.NoError(t, err)
	assert.Equal(t, "jdoe", username)
}

func TestGroupsFromClaim(t *testing.T) {
	// prepare
	for _, tt := range []struct {
		casename string
		input    interface{}
		expected []string
	}{
		{
			"single-string",
			"department-1",
			[]string{"department-1"},
		},
		{
			"multiple-strings",
			[]string{"department-1", "department-2"},
			[]string{"department-1", "department-2"},
		},
		{
			"multiple-things",
			[]interface{}{"department-1", 123},
			[]string{"department-1", "123"},
		},
	} {
		t.Run(tt.casename, func(t *testing.T) {
			claims := map[string]interface{}{
				"sub":         "jdoe",
				"memberships": tt.input,
			}

			// test
			groups, err := getGroupsFromClaims(claims, "memberships")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, groups)
		})
	}
}
