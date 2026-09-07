package configuration

import "testing"

func TestDatabaseDefaultsUseLocalRootAuthentication(t *testing.T) {
	for _, key := range []string{"VALIO_DB_URL", "VALIO_DB_NAMESPACE", "VALIO_DB_DATABASE", "VALIO_DB_USER", "VALIO_DB_PASSWORD", "VALIO_DB_AUTH_LEVEL", "VALIO_TEST_DB_URL", "VALIO_TEST_DB_PASSWORD"} {
		t.Setenv(key, "")
	}

	config, err := Database()
	if err != nil {
		t.Fatal(err)
	}
	if config.Endpoint != DefaultDatabaseURL || config.Namespace != DefaultDatabaseNamespace || config.Database != DefaultDatabaseName || config.Username != DefaultDatabaseUser || config.Password != DefaultDatabasePassword || config.DatabaseAuth {
		t.Fatalf("unexpected default database configuration: %#v", config)
	}
}

func TestDatabasePrefersNormalOverridesOverTestOverrides(t *testing.T) {
	t.Setenv("VALIO_TEST_DB_URL", "http://localhost:19000")
	t.Setenv("VALIO_TEST_DB_PASSWORD", "test-password")
	t.Setenv("VALIO_DB_URL", "http://localhost:20000")
	t.Setenv("VALIO_DB_PASSWORD", "configured-password")
	t.Setenv("VALIO_DB_NAMESPACE", "configured_namespace")
	t.Setenv("VALIO_DB_DATABASE", "configured_database")
	t.Setenv("VALIO_DB_USER", "configured_user")
	t.Setenv("VALIO_DB_AUTH_LEVEL", "database")

	config, err := Database()
	if err != nil {
		t.Fatal(err)
	}
	if config.Endpoint != "http://localhost:20000" || config.Password != "configured-password" || config.Namespace != "configured_namespace" || config.Database != "configured_database" || config.Username != "configured_user" || !config.DatabaseAuth {
		t.Fatalf("normal overrides were not applied: %#v", config)
	}
}

func TestDatabaseIgnoresTestOnlyEnvironmentSettings(t *testing.T) {
	t.Setenv("VALIO_DB_URL", "")
	t.Setenv("VALIO_DB_PASSWORD", "")
	t.Setenv("VALIO_TEST_DB_URL", "http://localhost:19000")
	t.Setenv("VALIO_TEST_DB_PASSWORD", "test-password")

	config, err := Database()
	if err != nil {
		t.Fatal(err)
	}
	if config.Endpoint != DefaultDatabaseURL || config.Password != DefaultDatabasePassword {
		t.Fatalf("test-only settings changed runtime defaults: %#v", config)
	}
}

func TestApplicationTokenDefaultAndOverride(t *testing.T) {
	t.Setenv("VALIO_API_TOKEN", "")
	if got := ApplicationToken(); got != DefaultAPIToken {
		t.Fatalf("default token = %q, want %q", got, DefaultAPIToken)
	}
	t.Setenv("VALIO_API_TOKEN", "configured-token")
	if got := ApplicationToken(); got != "configured-token" {
		t.Fatalf("override token = %q", got)
	}
}
