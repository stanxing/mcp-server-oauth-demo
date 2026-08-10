package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

type signedTokenVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

func newAccessTokenVerifier(ctx context.Context, cfg config) (auth.TokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover OAUTH_ISSUER_URL %q: %w", cfg.IssuerURL, err)
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID:             cfg.ResourceURL,
		SupportedSigningAlgs: cfg.SigningAlgs,
	})
	return accessTokenVerifier(verifier), nil
}

func accessTokenVerifier(verifier signedTokenVerifier) auth.TokenVerifier {
	return func(ctx context.Context, rawToken string, _ *http.Request) (*auth.TokenInfo, error) {
		fp := tokenFingerprint(rawToken)

		token, err := verifier.Verify(ctx, rawToken)
		if err != nil {
			log.Printf("auth: token=%s rejected: signature/claims verification failed: %v", fp, err)
			return nil, fmt.Errorf("%w: access token verification failed", auth.ErrInvalidToken)
		}
		if token.Subject == "" {
			log.Printf("auth: token=%s rejected: missing subject claim (issuer=%s)", fp, token.Issuer)
			return nil, fmt.Errorf("%w: access token has no subject", auth.ErrInvalidToken)
		}

		var claims accessTokenClaims
		if err := token.Claims(&claims); err != nil {
			log.Printf("auth: token=%s rejected: failed to decode claims: %v", fp, err)
			return nil, fmt.Errorf("%w: access token claims are invalid", auth.ErrInvalidToken)
		}

		extra := map[string]any{}
		if claims.Email != "" {
			extra["email"] = claims.Email
		}
		if claims.PreferredUsername != "" {
			extra["preferred_username"] = claims.PreferredUsername
		}

		log.Printf("auth: token=%s accepted issuer=%s subject=%s scopes=%v expiry=%s email=%s",
			fp, token.Issuer, token.Subject, claims.Scope, token.Expiry.Format(time.RFC3339), claims.Email)

		return &auth.TokenInfo{
			Scopes:     claims.Scope,
			Expiration: token.Expiry,
			UserID:     token.Subject,
			Extra:      extra,
		}, nil
	}
}

// tokenFingerprint returns a short, non-reversible identifier derived from a
// raw bearer token so auth log lines can be correlated across requests
// without ever writing the credential itself to the log.
func tokenFingerprint(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:6])
}

type accessTokenClaims struct {
	Scope             scopeClaim `json:"scope"`
	Email             string     `json:"email"`
	PreferredUsername string     `json:"preferred_username"`
}

type scopeClaim []string

func (s *scopeClaim) UnmarshalJSON(data []byte) error {
	var scopeString string
	if err := json.Unmarshal(data, &scopeString); err == nil {
		*s = strings.Fields(scopeString)
		return nil
	}

	var scopes []string
	if err := json.Unmarshal(data, &scopes); err != nil {
		return fmt.Errorf("scope must be a string or an array of strings")
	}
	*s = scopes
	return nil
}
