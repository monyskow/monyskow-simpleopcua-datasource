package models

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestParseSettings_BackfillClientCertConfigured(t *testing.T) {
	// Simulate a legacy datasource: cert bytes present in secureJsonData but
	// clientCertConfigured absent from jsonData (saved before the flag was introduced).
	jsonData, err := json.Marshal(map[string]interface{}{
		"endpoint":     "opc.tcp://localhost:4840",
		"securityMode": "Sign",
	})
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}

	settings := backend.DataSourceInstanceSettings{
		JSONData: jsonData,
		DecryptedSecureJSONData: map[string]string{
			"clientCert": "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----",
			"clientKey":  "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----",
		},
	}

	ds, err := ParseSettings(settings)
	if err != nil {
		t.Fatalf("ParseSettings: %v", err)
	}

	if !ds.ClientCertConfigured {
		t.Error("expected ClientCertConfigured=true for legacy datasource with cert bytes but no flag")
	}
}

func TestParseSettings_NoBackfillWhenKeyMissing(t *testing.T) {
	// Cert bytes present but key absent — backend skips cert loading in this case,
	// so the flag must not be set to avoid a false "configured" status.
	jsonData, err := json.Marshal(map[string]interface{}{
		"endpoint":     "opc.tcp://localhost:4840",
		"securityMode": "Sign",
	})
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}

	settings := backend.DataSourceInstanceSettings{
		JSONData: jsonData,
		DecryptedSecureJSONData: map[string]string{
			"clientCert": "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----",
			// clientKey intentionally absent
		},
	}

	ds, err := ParseSettings(settings)
	if err != nil {
		t.Fatalf("ParseSettings: %v", err)
	}

	if ds.ClientCertConfigured {
		t.Error("expected ClientCertConfigured=false when clientKey is missing (backend would skip cert loading)")
	}
}

func TestParseSettings_NoBackfillWhenNoCert(t *testing.T) {
	// Datasource without any cert — flag must remain false.
	jsonData, err := json.Marshal(map[string]interface{}{
		"endpoint": "opc.tcp://localhost:4840",
	})
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}

	settings := backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: map[string]string{},
	}

	ds, err := ParseSettings(settings)
	if err != nil {
		t.Fatalf("ParseSettings: %v", err)
	}

	if ds.ClientCertConfigured {
		t.Error("expected ClientCertConfigured=false when no cert bytes present")
	}
}

func TestParseSettings_FlagExplicitlySetInJsonData(t *testing.T) {
	// Datasource where the flag is already true in jsonData (normal post-fix case).
	jsonData, err := json.Marshal(map[string]interface{}{
		"endpoint":             "opc.tcp://localhost:4840",
		"clientCertConfigured": true,
	})
	if err != nil {
		t.Fatalf("marshal jsonData: %v", err)
	}

	// No cert bytes in secureJsonData (provisioned datasource scenario: bytes are in
	// Grafana's encrypted store, not returned to backend; flag comes from jsonData).
	settings := backend.DataSourceInstanceSettings{
		JSONData:                jsonData,
		DecryptedSecureJSONData: map[string]string{},
	}

	ds, err := ParseSettings(settings)
	if err != nil {
		t.Fatalf("ParseSettings: %v", err)
	}

	if !ds.ClientCertConfigured {
		t.Error("expected ClientCertConfigured=true when explicitly set in jsonData")
	}
}
