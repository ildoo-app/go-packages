package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// StaticKeyProvider verifies tokens with a fixed RSA public key.
type StaticKeyProvider struct {
	key *rsa.PublicKey
	kid string
}

func NewStaticKeyProvider(kid string, key *rsa.PublicKey) *StaticKeyProvider {
	return &StaticKeyProvider{key: key, kid: kid}
}

func (p *StaticKeyProvider) KeyFunc(token *jwt.Token) (any, error) {
	if token.Method != jwt.SigningMethodRS256 {
		return nil, fmt.Errorf("unexpected signing method")
	}
	return p.key, nil
}

// JWKSKeyProvider fetches and caches JWKS from a remote URL.
type JWKSKeyProvider struct {
	url        string
	httpClient *http.Client
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	fetchedAt  time.Time
	ttl        time.Duration
}

func NewJWKSKeyProvider(jwksURL string) *JWKSKeyProvider {
	return &JWKSKeyProvider{
		url: jwksURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		keys: make(map[string]*rsa.PublicKey),
		ttl:  5 * time.Minute,
	}
}

func (p *JWKSKeyProvider) KeyFunc(token *jwt.Token) (any, error) {
	if token.Method != jwt.SigningMethodRS256 {
		return nil, fmt.Errorf("unexpected signing method")
	}
	kid, _ := token.Header["kid"].(string)
	key, err := p.keyFor(kid)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (p *JWKSKeyProvider) keyFor(kid string) (*rsa.PublicKey, error) {
	p.mu.RLock()
	if key, ok := p.keys[kid]; ok && time.Since(p.fetchedAt) < p.ttl {
		p.mu.RUnlock()
		return key, nil
	}
	p.mu.RUnlock()

	if err := p.refresh(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	key, ok := p.keys[kid]
	if !ok {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}
	return key, nil
}

func (p *JWKSKeyProvider) refresh() error {
	resp, err := p.httpClient.Get(p.url)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch jwks: status %d", resp.StatusCode)
	}

	var payload struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(payload.Keys))
	for _, key := range payload.Keys {
		if key.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}
		keys[key.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
	}

	p.mu.Lock()
	p.keys = keys
	p.fetchedAt = time.Now()
	p.mu.Unlock()
	return nil
}

type accessClaims struct {
	SessionID  string `json:"sid"`
	MerchantID string `json:"tid,omitempty"`
	jwt.RegisteredClaims
}

// JWTVerifier verifies RS256 access tokens using the provided key func.
type JWTVerifier struct {
	keyFunc  jwt.Keyfunc
	issuer   string
	audience string
}

func NewJWTVerifier(keyFunc jwt.Keyfunc, issuer, audience string) *JWTVerifier {
	return &JWTVerifier{keyFunc: keyFunc, issuer: issuer, audience: audience}
}

func (v *JWTVerifier) VerifyAccessToken(_ context.Context, rawToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &accessClaims{}, v.keyFunc,
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid claims")
	}
	aud := ""
	if len(claims.Audience) > 0 {
		aud = claims.Audience[0]
	}
	return &Claims{
		Subject:    claims.Subject,
		SessionID:  claims.SessionID,
		MerchantID: claims.MerchantID,
		Audience:   aud,
	}, nil
}
