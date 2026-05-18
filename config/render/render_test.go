package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderEnv(t *testing.T) {
	t.Setenv("FOO", "bar")

	rendered, err := testRenderer(t).Render(`value: "${env:FOO}"`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != `value: "bar"` {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestRenderMultipleEnv(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("BAZ", "qux")

	rendered, err := testRenderer(t).Render(`${env:FOO}-${env:BAZ}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "bar-qux" {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestMissingEnvReturnsError(t *testing.T) {
	_, err := testRenderer(t).Render(`${env:MISSING_ENV_FOR_TEST}`)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing value for env:MISSING_ENV_FOR_TEST") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDuplicateMissingEnvReturnsErrorOnce(t *testing.T) {
	_, err := testRenderer(t).Render(`${env:MISSING_ENV_FOR_TEST} ${env:MISSING_ENV_FOR_TEST}`)
	if err == nil {
		t.Fatal("expected error")
	}
	if count := strings.Count(err.Error(), "missing value for env:MISSING_ENV_FOR_TEST"); count != 1 {
		t.Fatalf("expected one missing error, got %d: %v", count, err)
	}
}

func TestDollarFormsNotProcessed(t *testing.T) {
	raw := `$VAR ${VAR} abc$def`

	rendered, err := testRenderer(t).Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != raw {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestEscapedDollarPlaceholder(t *testing.T) {
	t.Setenv("FOO", "bar")

	rendered, err := testRenderer(t).Render(`$${env:FOO}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != `${env:FOO}` {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestEscapedBackslashPlaceholder(t *testing.T) {
	t.Setenv("FOO", "bar")

	rendered, err := testRenderer(t).Render(`\${env:FOO}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != `${env:FOO}` {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestUnknownSourceReturnsError(t *testing.T) {
	_, err := testRenderer(t).Render(`${vault:secret}`)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unsupported placeholder source "vault"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmptyKeyReturnsError(t *testing.T) {
	_, err := testRenderer(t).Render(`${env:}`)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid placeholder ${env:}") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmptySourceReturnsError(t *testing.T) {
	_, err := testRenderer(t).Render(`${:FOO}`)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid placeholder ${:FOO}") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderOnePassOnly(t *testing.T) {
	t.Setenv("FOO", `${env:BAR}`)
	t.Setenv("BAR", "bar")

	rendered, err := testRenderer(t).Render(`${env:FOO}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != `${env:BAR}` {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestEnvValuePreservesDollar(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", `abc$def`)

	rendered, err := testRenderer(t).Render(`${env:MYSQL_PASSWORD}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != `abc$def` {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestRenderFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	rendered, err := testRenderer(t).Render(`${file:` + path + `}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "secret" {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestRenderFileTrimsTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(path, []byte("secret\r\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	rendered, err := testRenderer(t).Render(`${file:` + path + `}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "secret" {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestRenderDockerSecret(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "mysql_password"), []byte("secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	rendered, err := NewRenderer(NewDockerSecretResolver(root)).Render(`${docker-secret:mysql_password}`)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != "secret" {
		t.Fatalf("unexpected rendered value: %q", rendered)
	}
}

func TestDockerSecretInvalidName(t *testing.T) {
	cases := []string{
		`${docker-secret:../secret}`,
		`${docker-secret:dir/secret}`,
		`${docker-secret:dir\secret}`,
	}

	for _, tc := range cases {
		t.Run(
			tc, func(t *testing.T) {
				_, err := NewRenderer(NewDockerSecretResolver(t.TempDir())).Render(tc)
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), "invalid docker secret name") {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		)
	}
}

func testRenderer(t *testing.T) *Renderer {
	t.Helper()
	return NewRenderer(
		EnvResolver{},
		FileResolver{},
		NewDockerSecretResolver(t.TempDir()),
	)
}
