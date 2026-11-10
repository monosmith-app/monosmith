package scaffold

import "testing"

func TestIsSecretName(t *testing.T) {
	cases := map[string]bool{
		".env": true, "app/.env.local": true, ".env.production": true, ".env.example": false,
		"keys/server.pem": true, "id_rsa": true, "id_rsa.pub": true, "credentials.json": true,
		".npmrc": true, "README.md": false, "src/main.go": false, "environment.ts": false,
	}
	for name, want := range cases {
		if got := IsSecretName(name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}
