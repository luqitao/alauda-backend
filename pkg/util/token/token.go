package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

var serviceAccountIssuers []string = []string{
	"kubernetes/serviceaccount",
	"https://kubernetes.default.svc.cluster.local",
}

type JWEToken struct {
	Issuer  string `json:"iss"`
	Subject string `json:"sub"`
	// Audience has two types: string and []slice, default use string, in tStack use []string.
	// see http://jira.alauda.cn/browse/AIT-2908
	Audience      interface{} `json:"aud"`
	Expiry        int         `json:"exp"`
	IssuedAt      int         `json:"iat"`
	Nonce         string      `json:"nonce"`
	Email         string      `json:"email"`
	EmailVerified bool        `json:"email_verified"`
	Name          string      `json:"name"`
	Groups        []string    `json:"groups"`
	Ext           jwtTokenExt `json:"ext"`
	MetadataName  string
	Sid           string `json:"sid,omitempty"`
}

func (t *JWEToken) IsServiceAccount() bool {
	for _, iss := range serviceAccountIssuers {
		if t.Issuer == iss {
			return true
		}
	}
	return false
}

type jwtTokenExt struct {
	IsAdmin bool   `json:"is_admin"`
	ConnID  string `json:"conn_id"`
}

func ParseJWTFromHeader(request *http.Request) (*JWEToken, error) {
	rawToken, err := ParseRawToken(request)
	if err != nil {
		return nil, err
	}
	return ParseJWT(rawToken)
}

func ParseRawToken(request *http.Request) (string, error) {
	var rawToken string
	authorization := request.Header.Get("Authorization")
	if strings.TrimSpace(authorization) == "" {
		return "", apiErrors.NewUnauthorized("Token required")
	}
	switch {
	case strings.HasPrefix(strings.TrimSpace(authorization), "Bearer"):
		rawToken = strings.TrimPrefix(strings.TrimSpace(authorization), "Bearer")
	case strings.HasPrefix(strings.TrimSpace(authorization), "bearer"):
		rawToken = strings.TrimPrefix(strings.TrimSpace(authorization), "bearer")
	}
	return rawToken, nil
}

func ParseJWT(rawToken string) (*JWEToken, error) {
	var (
		token JWEToken
	)

	if rawToken == "" {
		return nil, apiErrors.NewUnauthorized("Token required")
	}
	parts := strings.Split(rawToken, ".")
	if len(parts) < 2 {
		return nil, apiErrors.NewUnauthorized(fmt.Sprintf("oidc: malformed jwt, expected 3 parts got %d", len(parts)))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, apiErrors.NewUnauthorized(fmt.Sprintf("oidc: malformed jwt payload: %v", err))
	}
	if err := json.Unmarshal(payload, &token); err != nil {
		fmt.Println(err)
		return nil, apiErrors.NewUnauthorized(err.Error())
	}

	return &token, nil
}
