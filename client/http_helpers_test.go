// Copyright 2021 Harness Inc. All rights reserved.
// Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
// that can be found in the licenses directory at the root of this repository, also available at
// https://polyformproject.org/wp-content/uploads/2020/05/PolyForm-Free-Trial-1.0.0.txt.

package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name                string
		endpoint            string
		token               string
		accountID           string
		skipverify          bool
		additionalCertsDir  string
		base64MtlsCert      string
		base64MtlsKey       string
		wantClientCreated   bool
		wantEndpointTrimmed bool
	}{
		{
			name:                "basic client creation",
			endpoint:            "https://ti-service.example.com/",
			token:               "test-token",
			accountID:           "account123",
			skipverify:          false,
			wantClientCreated:   true,
			wantEndpointTrimmed: true,
		},
		{
			name:                "endpoint without trailing slash",
			endpoint:            "https://ti-service.example.com",
			token:               "test-token",
			accountID:           "account123",
			wantClientCreated:   true,
			wantEndpointTrimmed: true,
		},
		{
			name:              "client with skipverify",
			endpoint:          "https://ti-service.example.com",
			token:             "test-token",
			accountID:         "account123",
			skipverify:        true,
			wantClientCreated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewHTTPClient(
				tt.endpoint,
				tt.token,
				tt.accountID,
				"org123",
				"project123",
				"pipeline123",
				"build123",
				"stage123",
				"repo123",
				"sha123",
				"commit-link",
				tt.skipverify,
				tt.additionalCertsDir,
				tt.base64MtlsCert,
				tt.base64MtlsKey,
				"parent123",
			)

			if client == nil {
				t.Fatal("NewHTTPClient() returned nil")
			}

			if client.Endpoint != strings.TrimSuffix(tt.endpoint, "/") {
				t.Errorf("NewHTTPClient() endpoint = %v, want %v", client.Endpoint, strings.TrimSuffix(tt.endpoint, "/"))
			}

			if client.Token != tt.token {
				t.Errorf("NewHTTPClient() token = %v, want %v", client.Token, tt.token)
			}

			if client.AccountID != tt.accountID {
				t.Errorf("NewHTTPClient() accountID = %v, want %v", client.AccountID, tt.accountID)
			}

			if tt.skipverify && client.Client == nil {
				t.Error("NewHTTPClient() should create custom client when skipverify is true")
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "existing file",
			filename: tmpFile.Name(),
			want:     true,
		},
		{
			name:     "non-existent file",
			filename: "/nonexistent/file/path",
			want:     false,
		},
		{
			name:     "directory (should return false)",
			filename: tmpDir,
			want:     false,
		},
		{
			name:     "empty string",
			filename: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileExists(tt.filename)
			if got != tt.want {
				t.Errorf("fileExists(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLoadCertsFromBase64(t *testing.T) {
	// Generate a test certificate and key
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	// Encode to base64
	certBase64 := base64.StdEncoding.EncodeToString(certPEM)
	keyBase64 := base64.StdEncoding.EncodeToString(keyPEM)

	tests := []struct {
		name        string
		certBase64  string
		keyBase64   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid base64 cert and key",
			certBase64: certBase64,
			keyBase64:  keyBase64,
			wantErr:    false,
		},
		{
			name:        "invalid base64 cert",
			certBase64:  "invalid-base64",
			keyBase64:   keyBase64,
			wantErr:     true,
			errContains: "failed to decode base64 certificate",
		},
		{
			name:        "invalid base64 key",
			certBase64:  certBase64,
			keyBase64:   "invalid-base64",
			wantErr:     true,
			errContains: "failed to decode base64 key",
		},
		{
			name:        "invalid cert/key pair",
			certBase64:  base64.StdEncoding.EncodeToString([]byte("not a cert")),
			keyBase64:   base64.StdEncoding.EncodeToString([]byte("not a key")),
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadCertsFromBase64(tt.certBase64, tt.keyBase64)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadCertsFromBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Certificate == nil {
					t.Error("loadCertsFromBase64() returned empty certificate")
				}
			} else if tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("loadCertsFromBase64() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestLoadMTLSCertsFromFiles(t *testing.T) {
	// Generate test certificate and key files
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	// Create temporary files
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	defer os.Remove(certFile.Name())
	certFile.Write(certPEM)
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer os.Remove(keyFile.Name())
	keyFile.Write(keyPEM)
	keyFile.Close()

	tests := []struct {
		name     string
		certFile string
		keyFile  string
		want     bool
	}{
		{
			name:     "valid cert and key files",
			certFile: certFile.Name(),
			keyFile:  keyFile.Name(),
			want:     true,
		},
		{
			name:     "non-existent cert file",
			certFile: "/nonexistent/cert.pem",
			keyFile:  keyFile.Name(),
			want:     false,
		},
		{
			name:     "non-existent key file",
			certFile: certFile.Name(),
			keyFile:  "/nonexistent/key.pem",
			want:     false,
		},
		{
			name:     "both files non-existent",
			certFile: "/nonexistent/cert.pem",
			keyFile:  "/nonexistent/key.pem",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cert := loadMTLSCertsFromFiles(tt.certFile, tt.keyFile)
			if got != tt.want {
				t.Errorf("loadMTLSCertsFromFiles() = %v, want %v", got, tt.want)
			}
			if tt.want && cert.Certificate == nil {
				t.Error("loadMTLSCertsFromFiles() returned empty certificate when expected success")
			}
		})
	}
}

func TestLoadMTLSCerts(t *testing.T) {
	// Generate test certificate and key
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	certBase64 := base64.StdEncoding.EncodeToString(certPEM)
	keyBase64 := base64.StdEncoding.EncodeToString(keyPEM)

	// Create temporary files
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	defer os.Remove(certFile.Name())
	certFile.Write(certPEM)
	certFile.Close()

	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer os.Remove(keyFile.Name())
	keyFile.Write(keyPEM)
	keyFile.Close()

	tests := []struct {
		name           string
		base64Cert     string
		base64Key      string
		defaultCertFile string
		defaultKeyFile  string
		want           bool
	}{
		{
			name:           "load from base64",
			base64Cert:     certBase64,
			base64Key:      keyBase64,
			defaultCertFile: "/nonexistent",
			defaultKeyFile:  "/nonexistent",
			want:           true,
		},
		{
			name:           "fallback to files when base64 fails",
			base64Cert:     "invalid",
			base64Key:      "invalid",
			defaultCertFile: certFile.Name(),
			defaultKeyFile:  keyFile.Name(),
			want:           true,
		},
		{
			name:           "no certs available",
			base64Cert:     "",
			base64Key:      "",
			defaultCertFile: "/nonexistent",
			defaultKeyFile:  "/nonexistent",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cert := loadMTLSCerts(tt.base64Cert, tt.base64Key, tt.defaultCertFile, tt.defaultKeyFile)
			if got != tt.want {
				t.Errorf("loadMTLSCerts() = %v, want %v", got, tt.want)
			}
			if tt.want && cert.Certificate == nil {
				t.Error("loadMTLSCerts() returned empty certificate when expected success")
			}
		})
	}
}

func TestLoadRootCAs(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-certs-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid PEM certificate file
	certPEM := `-----BEGIN CERTIFICATE-----
MIICUTCCAfugAwIBAgIBADANBgkqhkiG9w0BAQQFADBXMQswCQYDVQQGEwJDTjEL
MAkGA1UECAwCQkoxCzAJBgNVBAcMAkJKMQswCQYDVQQKDAJDTjELMAkGA1UECwwC
QkoxCzAJBgNVBAMMAkNKMB4XDTIwMDEwMTAwMDAwMFoXDTIxMDEwMTAwMDAwMFow
VzELMAkGA1UEBhMCQ04xCzAJBgNVBAgMAkJKMQswCQYDVQQHDAJCSjELMAkGA1UE
CgwCQ04xCzAJBgNVBAsMAkJKMQswCQYDVQQDDAJDSjCBnzANBgkqhkiG9w0BAQEF
AAOBjQAwgYkCgYEAwIDAQABo4GJMIGGMB0GA1UdDgQWBBTgL3kqj+J2Y3K1vJ3K
1vJ3K1vJ3DAfBgNVHSMEGDAWgBTgL3kqj+J2Y3K1vJ3K1vJ3K1vJ3DAMBgNVHRME
BTADAQH/MAsGA1UdDwQEAwIBBjANBgkqhkiG9w0BAQQFAAOBgQCXgL3kqj+J2Y3K
1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3K1vJ3
-----END CERTIFICATE-----`

	certFile := filepath.Join(tmpDir, "cert.pem")
	err = os.WriteFile(certFile, []byte(certPEM), 0644)
	if err != nil {
		t.Fatalf("Failed to write cert file: %v", err)
	}

	// Create an invalid file
	invalidFile := filepath.Join(tmpDir, "invalid.txt")
	err = os.WriteFile(invalidFile, []byte("not a cert"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	tests := []struct {
		name             string
		additionalCertsDir string
		wantNil          bool
	}{
		{
			name:             "empty directory path",
			additionalCertsDir: "",
			wantNil:          true,
		},
		{
			name:             "directory with valid cert",
			additionalCertsDir: tmpDir,
			wantNil:          false,
		},
		{
			name:             "non-existent directory",
			additionalCertsDir: "/nonexistent/dir",
			wantNil:          false, // Returns empty pool, not nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loadRootCAs(tt.additionalCertsDir)
			if (got == nil) != tt.wantNil {
				t.Errorf("loadRootCAs() = %v, want nil = %v", got, tt.wantNil)
			}
		})
	}
}

func TestClientWithTLSConfig(t *testing.T) {
	// Generate test certificate
	certPEM, keyPEM, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	testCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("Failed to create test cert: %v", err)
	}

	// Create a test cert pool
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(certPEM)

	tests := []struct {
		name      string
		skipverify bool
		rootCAs    *x509.CertPool
		mtlsEnabled bool
		cert       tls.Certificate
		wantClient bool
	}{
		{
			name:       "skipverify enabled",
			skipverify: true,
			rootCAs:    nil,
			mtlsEnabled: false,
			wantClient: true,
		},
		{
			name:       "with root CAs",
			skipverify: false,
			rootCAs:    rootCAs,
			mtlsEnabled: false,
			wantClient: true,
		},
		{
			name:       "with mTLS",
			skipverify: false,
			rootCAs:    nil,
			mtlsEnabled: true,
			cert:       testCert,
			wantClient: true,
		},
		{
			name:       "with root CAs and mTLS",
			skipverify: false,
			rootCAs:    rootCAs,
			mtlsEnabled: true,
			cert:       testCert,
			wantClient: true,
		},
		{
			name:       "skipverify with root CAs (rootCAs ignored)",
			skipverify: true,
			rootCAs:    rootCAs,
			mtlsEnabled: false,
			wantClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clientWithTLSConfig(tt.skipverify, tt.rootCAs, tt.mtlsEnabled, tt.cert)
			if (got != nil) != tt.wantClient {
				t.Errorf("clientWithTLSConfig() = %v, want client = %v", got, tt.wantClient)
			}
			if got != nil {
				if got.Transport == nil {
					t.Error("clientWithTLSConfig() returned client without Transport")
				}
			}
		})
	}
}

func TestCreateBackoff(t *testing.T) {
	tests := []struct {
		name          string
		maxElapsedTime time.Duration
		wantNil       bool
	}{
		{
			name:          "zero duration",
			maxElapsedTime: 0,
			wantNil:       false,
		},
		{
			name:          "positive duration",
			maxElapsedTime: 5 * time.Minute,
			wantNil:       false,
		},
		{
			name:          "large duration",
			maxElapsedTime: 1 * time.Hour,
			wantNil:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createBackoff(tt.maxElapsedTime)
			if (got == nil) != tt.wantNil {
				t.Errorf("createBackoff() = %v, want nil = %v", got, tt.wantNil)
			}
			if got != nil && got.MaxElapsedTime != tt.maxElapsedTime {
				t.Errorf("createBackoff() MaxElapsedTime = %v, want %v", got.MaxElapsedTime, tt.maxElapsedTime)
			}
		})
	}
}

func TestCreateInfiniteBackoff(t *testing.T) {
	got := createInfiniteBackoff()
	if got == nil {
		t.Error("createInfiniteBackoff() returned nil")
	}
	if got.MaxElapsedTime != 0 {
		t.Errorf("createInfiniteBackoff() MaxElapsedTime = %v, want 0", got.MaxElapsedTime)
	}
}

// Helper functions

func generateTestCert() ([]byte, []byte, error) {
	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Encode private key using PKCS1 format (more compatible)
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
				indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

