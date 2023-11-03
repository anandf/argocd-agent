package issuer

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jannfis/argocd-agent/test/fake/certs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func signedTokenWithClaims(method jwt.SigningMethod, key interface{}, claims jwt.Claims) (string, error) {
	tok := jwt.NewWithClaims(method, claims)
	return tok.SignedString(key)
}

func Test_Issuer(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var tok string
	t.Run("Issue an access token", func(t *testing.T) {
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		tok, err = i.IssueAccessToken("agent", 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("Validate access token", func(t *testing.T) {
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		require.NoError(t, err)
		sub, err := c.GetSubject()
		require.NoError(t, err)
		assert.Equal(t, "agent", sub)
		c, err = i.ValidateRefreshToken(tok)
		require.Error(t, err)
		require.Nil(t, c)
	})
	t.Run("Issue a refresh token", func(t *testing.T) {
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		tok, err = i.IssueRefreshToken("agent", 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("Validate refresh token", func(t *testing.T) {
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateRefreshToken(tok)
		require.NoError(t, err)
		sub, err := c.GetSubject()
		require.NoError(t, err)
		assert.Equal(t, "agent", sub)
		c, err = i.ValidateAccessToken(tok)
		require.Error(t, err)
		require.Nil(t, c)
	})
	t.Run("JWT signed by another issuer", func(t *testing.T) {
		i1, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		i2, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		tok, err := i2.IssueAccessToken("agent", 5*time.Second)
		require.NoError(t, err)
		c, err := i1.ValidateAccessToken(tok)
		assert.ErrorContains(t, err, jwt.ErrSignatureInvalid.Error())
		assert.Nil(t, c)
	})

	t.Run("JWT signed with forbidden none method", func(t *testing.T) {
		tok, err := signedTokenWithClaims(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.RegisteredClaims{
			Issuer:    "server",
			Subject:   "agent",
			Audience:  jwt.ClaimStrings{"server"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		require.NoError(t, err)
		require.NotNil(t, tok)
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		assert.ErrorContains(t, err, jwt.ErrSignatureInvalid.Error())
		assert.Nil(t, c)
	})

	t.Run("JWT with invalid audience", func(t *testing.T) {
		tok, err := signedTokenWithClaims(jwt.SigningMethodRS512, key, jwt.RegisteredClaims{
			Issuer:    "server",
			Subject:   "agent",
			Audience:  jwt.ClaimStrings{"agent"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		require.NoError(t, err)
		require.NotNil(t, tok)
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		assert.ErrorContains(t, err, jwt.ErrTokenInvalidAudience.Error())
		assert.Nil(t, c)
	})

	t.Run("JWT with invalid issuer", func(t *testing.T) {
		tok, err := signedTokenWithClaims(jwt.SigningMethodRS512, key, jwt.RegisteredClaims{
			Issuer:    "agent",
			Subject:   "agent",
			Audience:  jwt.ClaimStrings{"server"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		require.NoError(t, err)
		require.NotNil(t, tok)
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		assert.ErrorContains(t, err, jwt.ErrTokenInvalidIssuer.Error())
		assert.Nil(t, c)
	})

	t.Run("Expired JWT", func(t *testing.T) {
		tok, err := signedTokenWithClaims(jwt.SigningMethodRS512, key, jwt.RegisteredClaims{
			Issuer:    "server",
			Subject:   "agent",
			Audience:  jwt.ClaimStrings{"server"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-5 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		require.NoError(t, err)
		require.NotNil(t, tok)
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		assert.ErrorContains(t, err, jwt.ErrTokenExpired.Error())
		assert.Nil(t, c)
	})

	t.Run("JWT not yet valid", func(t *testing.T) {
		tok, err := signedTokenWithClaims(jwt.SigningMethodRS512, key, jwt.RegisteredClaims{
			Issuer:    "server",
			Subject:   "agent",
			Audience:  jwt.ClaimStrings{"server"},
			NotBefore: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		require.NoError(t, err)
		require.NotNil(t, tok)
		i, err := NewIssuer("server", WithRSAPrivateKey(key))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		assert.ErrorContains(t, err, jwt.ErrTokenNotValidYet.Error())
		assert.Nil(t, c)
	})

	t.Run("Issue a JWT with key from file", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := path.Join(tempDir, "somekey.key")
		certs.WriteRSAPrivateKey(t, keyPath)
		i, err := NewIssuer("server", WithRSAPrivateKeyFromFile(keyPath))
		require.NoError(t, err)
		tok, err = i.IssueAccessToken("agent", 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("Create an issuer with public key from file", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := path.Join(tempDir, "somekey.key")
		certs.WriteRSAPublicKey(t, keyPath)
		i, err := NewIssuer("server", WithRSAPublicKeyFromFile(keyPath))
		require.NoError(t, err)
		require.NotNil(t, i)
	})

	t.Run("Validate tokens with public key", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := path.Join(tempDir, "somekey.key")
		key := certs.WriteRSAPrivateKey(t, keyPath)
		i, err := NewIssuer("server", WithRSAPrivateKeyFromFile(keyPath))
		require.NoError(t, err)
		require.NotNil(t, i)
		tok, err := i.IssueAccessToken("agent", 5*time.Second)
		require.NoError(t, err)
		i, err = NewIssuer("server", WithRSAPublicKey(&key.PublicKey))
		require.NoError(t, err)
		c, err := i.ValidateAccessToken(tok)
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("Issuer using RSA key from non-existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := path.Join(tempDir, "somekey.key")
		_, err := NewIssuer("server", WithRSAPrivateKeyFromFile(keyPath))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("Issuer using RSA key from file that has no PEM data", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := path.Join(tempDir, "somekey.key")
		os.WriteFile(keyPath, []byte("this is not a key"), 0600)
		_, err := NewIssuer("server", WithRSAPrivateKeyFromFile(keyPath))
		require.ErrorContains(t, err, "no valid PEM")
	})

	t.Run("Issuer using RSA key from file that has PEM without private key", func(t *testing.T) {
		tempDir := t.TempDir()
		basePath := path.Join(tempDir, "cert")
		certs.WriteSelfSignedCert(t, basePath, x509.Certificate{SerialNumber: big.NewInt(1)})
		_, err := NewIssuer("server", WithRSAPrivateKeyFromFile(basePath+".crt"))
		require.ErrorContains(t, err, "no RSA private key")
	})

}
