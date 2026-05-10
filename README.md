# oauth2-sso-demo

OAuth2 Authorization Code Flow SSO 實作練習，本機以 Keycloak 模擬正式授權伺服器的完整流程。

---

## 架構圖

```
Browser (localhost:5173)
    |
    | /api/* → proxy
    |
Vue3 Frontend (Vite dev server)
    |
    | HTTP
    |
Go Backend / Gin (localhost:8081)
    |          |
    |          +---> Redis (localhost:6379)   session storage
    |
    +---> Keycloak (localhost:8080)           OAuth2 / OIDC server
              |
              +--- Realm: demo
                      |
                      +--- Client: demo
                      +--- Roles: admin
                      +--- Users: testuser, testadmin
```

---

## Screenshots

### 前端

| 登入頁 | Keycloak 登入頁 |
|--------|----------------|
| ![登入頁](docs/screenshots/01_login.png) | ![Keycloak 登入](docs/screenshots/02_keycloak_login.png) |

| Dashboard（一般用戶） | Dashboard（Admin） |
|----------------------|-------------------|
| ![一般用戶](docs/screenshots/03_dashboard_user.png) | ![Admin](docs/screenshots/04_dashboard_admin.png) |

### Keycloak 設定

| Realm 設定 | Client 設定（Redirect URIs） |
|-----------|---------------------------|
| ![Realm](docs/screenshots/05_kc_realm.png) | ![Client](docs/screenshots/06_kc_client.png) |

| User 列表 | Role 列表 | testadmin Role Mapping |
|----------|----------|----------------------|
| ![Users](docs/screenshots/07_kc_users.png) | ![Roles](docs/screenshots/08_kc_roles.png) | ![Role Mapping](docs/screenshots/09_kc_role_mapping.png) |

---

## 使用工具

| 工具 | 用途 |
|------|------|
| **Keycloak 24** | 本機 OAuth2 / OIDC server，模擬正式環境的授權伺服器 |
| **Go + Gin** | 後端 HTTP server，處理 OAuth2 流程與 session 管理 |
| **go-redis** | 將 session 存入 Redis，支援多實例與自動 TTL 過期 |
| **Vue3 + Vite** | 前端，Vite proxy 解決跨域問題 |
| **Vue Router** | 前端路由，含 navigation guard 保護需登入頁面 |
| **Pinia** | 前端全域登入狀態管理 |
| **Docker Compose** | 一鍵啟動 Keycloak + Redis |

---

## 專案結構

```
oauth2-sso-demo/
├── docker-compose.yml
├── docs/
│   └── screenshots/
├── backend/
│   ├── main.go
│   ├── .env
│   ├── config/
│   │   └── config.go        讀取環境變數
│   ├── handler/
│   │   └── auth.go          /login /callback /logout /me /admin/data
│   ├── middleware/
│   │   └── auth.go          TokenRefresh、RequireRole
│   ├── jwks/
│   │   └── jwks.go          從 Keycloak 取公鑰並快取，用於 JWT 驗簽
│   └── store/
│       └── session.go       Redis session 存取
└── frontend/
    └── src/
        ├── stores/
        │   └── auth.js      Pinia，管理登入狀態
        ├── router/
        │   └── index.js     路由 + navigation guard
        └── views/
            ├── LoginView.vue
            └── DashboardView.vue
```

---

## 環境變數（backend/.env）

| 變數 | 說明 | 範例 |
|------|------|------|
| `KEYCLOAK_BASE` | Keycloak realm 的 OIDC base URL | `http://localhost:8080/realms/demo/protocol/openid-connect` |
| `CLIENT_ID` | Keycloak client ID | `demo` |
| `CLIENT_SECRET` | Keycloak client secret（從 Credentials 分頁取得） | `xxxxxx` |
| `REDIRECT_URI` | OAuth callback，指向後端 | `http://localhost:8081/api/auth/callback` |
| `POST_LOGOUT_URI` | 登出後跳回的網址 | `http://localhost:5173/` |
| `FRONTEND_URL` | 後端登入完成後導向前端的 base URL | `http://localhost:5173` |
| `REDIS_URL` | Redis 連線字串 | `redis://localhost:6379` |

---

## 啟動方式

### 1. 啟動 Keycloak + Redis

```bash
docker compose up -d
docker compose ps   # 確認兩個都是 healthy
```

---

### 2. Keycloak 初始設定（第一次需要）

開瀏覽器進 `http://localhost:8080`，帳密 `admin / admin`

```
建立 Realm：demo

建立 Client：
  Client ID: demo
  Client authentication: ON
  Valid redirect URIs: http://localhost:8081/api/auth/callback
  Post logout redirect URIs: http://localhost:5173/
  Web origins: http://localhost:5173

建立 Role：admin（Realm roles → Create role）

建立 User：
  testuser   → 不指派任何 role
  testadmin  → Role mapping 指派 admin role
  Credentials → Set password（Temporary: OFF）
```

---

### 3. 設定 backend/.env

複製 Keycloak `Clients → demo → Credentials → Client secret` 填入：

```bash
cp backend/.env.example backend/.env   # 若有 example 的話
# 填入 CLIENT_SECRET
```

---

### 4. 啟動 Go 後端

```bash
cd backend
go run main.go
# Listening and serving HTTP on :8081
```

---

### 5. 啟動 Vue3 前端

```bash
cd frontend
npm install
npm run dev
# http://localhost:5173
```

---

## API

| Method | 路徑 | 說明 | 保護 |
|--------|------|------|------|
| `GET` | `/api/auth/login` | 產生 state，redirect 到 Keycloak 登入頁 | — |
| `GET` | `/api/auth/callback` | 接收 code，換 token，建立 session，redirect 前端 | — |
| `GET` | `/api/auth/logout` | 清除 session，redirect Keycloak 全域登出 | — |
| `GET` | `/api/auth/me` | 回傳目前登入的使用者資訊 | session |
| `GET` | `/api/admin/data` | 管理員專屬資料 | session + admin role |

**`/api/auth/me` 回傳範例**

```json
{
  "sub": "使用者唯一 ID",
  "name": "Test User",
  "username": "testuser",
  "email": "test@demo.com",
  "roles": ["offline_access", "admin", "default-roles-demo", "uma_authorization"]
}
```

---

## 核心概念

**為什麼 callback 直接打後端，不經過前端？**

Keycloak 登入完後，code 直接送到 Go 後端（`/api/auth/callback`），後端換完 token 才把瀏覽器導到前端。這樣 `client_secret` 完全不會暴露給瀏覽器，前端也不需要處理 OAuth 邏輯。

**State 的作用**

`/api/auth/login` 產生隨機 state 存進 Redis，callback 回來時驗證是否一致。防止 CSRF 攻擊：確保這個 callback 是由你的 login 發起的，不是別人偽造的。

**id_token 和 access_token 的分工**

兩個 token 都是 JWT，但用途不同：
- `id_token`：帶使用者身份資訊（name、email、sub），供後端識別「你是誰」
- `access_token`：帶授權資訊（`realm_access.roles`），供後端判斷「你能做什麼」

**JWT 簽章驗證（JWKS）**

後端收到 id_token 後，從 Keycloak 的 `/certs` 端點取得 RSA 公鑰，驗證 JWT 簽章是否合法。公鑰快取 1 小時，避免每次都打 Keycloak。

**Token Refresh（靜默換新）**

每個 request 進來時，後端 middleware 檢查 access_token 是否快過期（剩不到 60 秒）。快過期就用 refresh_token 向 Keycloak 換新 token 並更新 Redis，使用者完全感知不到。

**Session 為什麼存 Redis 不存記憶體？**

記憶體 session 在 server 重啟或多實例部署時會消失。Redis 有 TTL 自動過期（30 分鐘），且多台 server 共享同一份 session。

**Vite proxy 的作用**

前端呼叫 `/api/auth/me`，Vite dev server 把這個 request 轉發到 `localhost:8081`，避免瀏覽器的跨域（CORS）限制，同時 cookie 也能正確帶上。

---

## 踩過的坑

**1. Keycloak healthcheck 打錯 port**

Keycloak 24 的 `/health/ready` 是在 management port 9000，不是 8080。但 `start-dev` 模式根本不開 9000，最後改用 bash 的 `/dev/tcp` 直接打 HTTP request 到 8080 來做 healthcheck。

**2. `docker restart` 不會套用新設定**

更新 `docker-compose.yml` 後用 `restart` 無效，必須 `docker compose down && docker compose up -d` 才會重建 container 套用新設定。

**3. `CMD-SHELL` 用 `/bin/sh`，不是 bash**

Docker healthcheck 的 `CMD-SHELL` 預設走 `/bin/sh`，但 `/dev/tcp` 是 bash 專屬語法。改成 `["CMD", "bash", "-c", "..."]` 才能正常執行。

**4. Volume 加入時機**

第一次啟動沒有 volume，Keycloak 設定存在 container 內部。後來加了 volume 重建 container，新的空 volume 蓋掉原本的設定，導致 realm 消失。需要在加 volume 後重新設定一次，之後才會持久化。

**5. Redirect URI 要指向後端**

原本 `REDIRECT_URI` 設成前端的 `localhost:5173/callback`，導致 Keycloak 把 code 送到前端，但前端沒跑起來就拒絕連線。改成直接指向後端 `localhost:8081/api/auth/callback`，由後端接收 code 並處理。

**6. roles 在 access_token，不在 id_token**

`realm_access.roles` 只存在 access_token，id_token 不帶 roles。一開始只解析 id_token 導致 roles 永遠是空的，需要另外 parse access_token 來取 roles。

**7. Role 比對要 case-insensitive**

Keycloak role 名稱是 `admin`（小寫），但程式碼寫的是 `"Admin"`，導致前端判斷顯示異常、後端 API 永遠 403。前端用 `.toLowerCase()` 比對，後端改用 `strings.EqualFold()` 解決。

---

## 關於這個專案

為了理解 OAuth2 Authorization Code Flow 實際運作而建的練習專案。本機以 Keycloak 模擬正式環境的授權伺服器，換掉 `.env` 的 URL 即可對接正式環境。

開發過程以 [Claude Code](https://claude.ai/code) 輔助。
