package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"oauth2-sso-demo/config"
)

const sessionTTL = 30 * time.Minute

// Session 代表一個登入狀態
type Session struct {
	State        string   `json:"state"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	IDToken      string   `json:"id_token"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	Sub      string   `json:"sub"`
	Name     string   `json:"name"`
	Username string   `json:"preferred_username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// SessionStore 負責所有 session 的存取
type SessionStore struct {
	client *redis.Client
}

var Store *SessionStore

// Init 在程式啟動時呼叫一次，建立 Redis 連線
func Init() {
	opts, err := redis.ParseURL(config.Cfg.RedisURL)
	if err != nil {
		panic("Redis URL 格式錯誤: " + err.Error())
	}

	client := redis.NewClient(opts)

	// 測試連線是否正常
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic("無法連接 Redis: " + err.Error())
	}

	Store = &SessionStore{client: client}
}

func key(sessionID string) string {
	return fmt.Sprintf("sess:%s", sessionID)
}

// Set 存入 session，TTL 到期後 Redis 自動刪除
func (s *SessionStore) Set(sessionID string, sess *Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.client.Set(context.Background(), key(sessionID), data, sessionTTL).Err()
}

// Get 取出 session，找不到時回傳 nil
func (s *SessionStore) Get(sessionID string) (*Session, error) {
	data, err := s.client.Get(context.Background(), key(sessionID)).Bytes()
	if err == redis.Nil {
		return nil, nil // session 不存在（不算 error）
	}
	if err != nil {
		return nil, err
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// Delete 刪除 session（登出時用）
func (s *SessionStore) Delete(sessionID string) error {
	return s.client.Del(context.Background(), key(sessionID)).Err()
}
