package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"oauth2-sso-demo/config"
	"oauth2-sso-demo/store"
)

// access_token 剩不到 60 秒就提前換新
const refreshThreshold = 60 * time.Second

// TokenRefresh 是一個 middleware，每個 request 進來時自動檢查 token 是否快過期
// 若快過期，在後端靜默刷新，使用者不會感知到
func TokenRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil || sid == "" {
			c.Next()
			return
		}

		sess, err := store.Store.Get(sid)
		if err != nil || sess == nil || sess.RefreshToken == "" {
			c.Next()
			return
		}

		// 讀 access_token 的過期時間
		exp, err := jwtExpiry(sess.AccessToken)
		if err != nil || time.Until(exp) > refreshThreshold {
			// 還沒到閾值，不需要刷新
			c.Next()
			return
		}

		// Token 快過期了，去 Keycloak 換新的
		newTokens, err := doRefresh(sess.RefreshToken)
		if err != nil {
			// 換失敗（refresh_token 也過期了），讓使用者重新登入
			c.Next()
			return
		}

		// 把新 token 存回 Redis
		sess.AccessToken = newTokens.AccessToken
		if newTokens.RefreshToken != "" {
			sess.RefreshToken = newTokens.RefreshToken
		}
		sess.IDToken = newTokens.IDToken
		store.Store.Set(sid, sess)

		c.Next()
	}
}

// RequireRole 檢查使用者是否擁有指定角色之一，沒有就回 403
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := c.Cookie("session_id")
		if err != nil || sid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
			c.Abort()
			return
		}

		sess, err := store.Store.Get(sid)
		if err != nil || sess == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
			c.Abort()
			return
		}

		// 只要符合其中一個 role 就放行
		for _, required := range roles {
			for _, userRole := range sess.User.Roles {
				if strings.EqualFold(userRole, required) {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":    "insufficient permissions",
			"required": roles,
		})
		c.Abort()
	}
}

// ── 內部工具 ──────────────────────────────────────────────────────────────────

// jwtExpiry 從 JWT payload 取出 exp（過期時間）
func jwtExpiry(tokenStr string) (time.Time, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid jwt")
	}
	decoded, err := base64.URLEncoding.DecodeString(padBase64(parts[1]))
	if err != nil {
		return time.Time{}, err
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return time.Time{}, err
	}
	return time.Unix(claims.Exp, 0), nil
}

type refreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// doRefresh 用 refresh_token 向 Keycloak 換取新的 access_token
func doRefresh(refreshToken string) (*refreshResp, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", config.Cfg.ClientID)
	form.Set("client_secret", config.Cfg.ClientSecret)
	form.Set("refresh_token", refreshToken)

	resp, err := http.PostForm(config.Cfg.KeycloakBase+"/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed: %s", body)
	}

	var result refreshResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func padBase64(s string) string {
	switch len(s) % 4 {
	case 2:
		return s + "=="
	case 3:
		return s + "="
	}
	return s
}
