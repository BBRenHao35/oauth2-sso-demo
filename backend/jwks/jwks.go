package jwks

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// 快取 Keycloak 的公鑰，每小時更新一次
type cache struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	certsURL  string
}

var c = &cache{keys: map[string]*rsa.PublicKey{}}

// Init 在程式啟動時設定 JWKS 端點 URL
func Init(certsURL string) {
	c.certsURL = certsURL
}

// GetKey 依 kid 取得對應的 RSA 公鑰，快取有效期 1 小時
func GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	fresh := time.Since(c.fetchedAt) < time.Hour
	c.mu.RUnlock()

	if ok && fresh {
		return key, nil
	}

	// 快取過期或找不到，重新去 Keycloak 拿
	if err := fetch(); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	key, ok = c.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}
	return key, nil
}

// fetch 從 Keycloak /certs 拿公鑰，解析成 rsa.PublicKey 存進快取
func fetch() error {
	resp, err := http.Get(c.certsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// JWKS 的格式
	var payload struct {
		Keys []struct {
			Kid string `json:"kid"` // Key ID
			Kty string `json:"kty"` // Key Type，我們只用 RSA
			N   string `json:"n"`   // RSA modulus（Base64URL）
			E   string `json:"e"`   // RSA exponent（Base64URL）
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, k := range payload.Keys {
		if k.Kty != "RSA" {
			continue
		}

		// 把 Base64URL 的 N 轉成 big.Int
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}

		// 把 Base64URL 的 E 轉成 int（通常是 65537）
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		var eBuf [4]byte
		copy(eBuf[4-len(eBytes):], eBytes)
		e := int(binary.BigEndian.Uint32(eBuf[:]))

		c.keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: e,
		}
	}

	c.fetchedAt = time.Now()
	return nil
}
