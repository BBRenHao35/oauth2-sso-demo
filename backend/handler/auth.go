package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"oauth2-sso-demo/config"
	"oauth2-sso-demo/store"
)

// ── 工具：產生隨機字串 ────────────────────────────────────────────────────────

func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getSessionID(c *gin.Context) (string, bool) {
	sid, err := c.Cookie("session_id")
	if err != nil || sid == "" {
		return "", false
	}
	return sid, true
}

// ── /api/auth/login ───────────────────────────────────────────────────────────

func Login(c *gin.Context) {
	state := randomString(16)
	sid := randomString(32)

	// 把 state 存進 Redis（此時還沒有 user 資訊，只存 state 用來驗證）
	if err := store.Store.Set(sid, &store.Session{State: state}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	c.SetCookie("session_id", sid, 600, "/", "", false, true)

	authURL := fmt.Sprintf(
		"%s/auth?response_type=code&client_id=%s&scope=openid&state=%s&redirect_uri=%s",
		config.Cfg.KeycloakBase,
		config.Cfg.ClientID,
		url.QueryEscape(state),
		url.QueryEscape(config.Cfg.RedirectURI),
	)

	c.Redirect(http.StatusFound, authURL)
}

// ── /api/auth/callback ────────────────────────────────────────────────────────

func Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	sid, ok := getSessionID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing session"})
		return
	}

	// 從 Redis 取出 session，驗證 state
	sess, err := store.Store.Get(sid)
	if err != nil || sess == nil || sess.State != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// 用 code 換 token
	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	// 解析 id_token 取得使用者資訊
	user, err := parseIDToken(tokenResp.IDToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse id_token"})
		return
	}

	// 把完整 session 存回 Redis（覆蓋原本只有 state 的那筆）
	if err := store.Store.Set(sid, &store.Session{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		User:         user,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	c.Redirect(http.StatusFound, config.Cfg.FrontendURL+"/dashboard")
}

// ── /api/auth/logout ──────────────────────────────────────────────────────────

func Logout(c *gin.Context) {
	idToken := ""

	if sid, ok := getSessionID(c); ok {
		if sess, _ := store.Store.Get(sid); sess != nil {
			idToken = sess.IDToken
		}
		store.Store.Delete(sid)
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)

	logoutURL := fmt.Sprintf(
		"%s/logout?id_token_hint=%s&post_logout_redirect_uri=%s",
		config.Cfg.KeycloakBase,
		url.QueryEscape(idToken),
		url.QueryEscape(config.Cfg.PostLogoutURI),
	)
	c.Redirect(http.StatusFound, logoutURL)
}

// ── /api/auth/me ──────────────────────────────────────────────────────────────

func Me(c *gin.Context) {
	sid, ok := getSessionID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
		return
	}

	sess, err := store.Store.Get(sid)
	if err != nil || sess == nil || sess.User.Sub == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sub":      sess.User.Sub,
		"name":     sess.User.Name,
		"username": sess.User.Username,
		"email":    sess.User.Email,
		"roles":    sess.User.Roles,
	})
}

// ── 內部工具函數 ──────────────────────────────────────────────────────────────

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func exchangeCodeForToken(code string) (*tokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", config.Cfg.ClientID)
	form.Set("client_secret", config.Cfg.ClientSecret)
	form.Set("redirect_uri", config.Cfg.RedirectURI)
	form.Set("code", code)

	resp, err := http.PostForm(config.Cfg.KeycloakBase+"/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("keycloak error: %s", body)
	}

	var result tokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func parseIDToken(idToken string) (store.UserInfo, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return store.UserInfo{}, fmt.Errorf("invalid JWT format")
	}

	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return store.UserInfo{}, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return store.UserInfo{}, err
	}

	user := store.UserInfo{
		Sub:      getString(raw, "sub"),
		Name:     getString(raw, "name"),
		Username: getString(raw, "preferred_username"),
		Email:    getString(raw, "email"),
	}

	if ra, ok := raw["realm_access"].(map[string]interface{}); ok {
		if rolesRaw, ok := ra["roles"].([]interface{}); ok {
			for _, r := range rolesRaw {
				if s, ok := r.(string); ok {
					user.Roles = append(user.Roles, s)
				}
			}
		}
	}

	return user, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
