package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

type Config struct {
	Username, PasswordHash, Salt string
	Secure                       bool
	Origins                      []string
}
type failure struct {
	count int
	at    time.Time
}
type Auth struct {
	db       *sql.DB
	config   Config
	mu       sync.Mutex
	failures map[string]failure
}
type Error struct {
	Code   string
	Status int
}

func (e *Error) Error() string { return e.Code }

func New(db *sql.DB, config Config) (*Auth, error) {
	hash, err := hex.DecodeString(config.PasswordHash)
	if err != nil || len(hash) != 32 {
		return nil, &Error{"AUTH_NOT_CONFIGURED", 503}
	}
	salt, err := hex.DecodeString(config.Salt)
	if err != nil || len(salt) != 16 || config.Username == "" {
		return nil, &Error{"AUTH_NOT_CONFIGURED", 503}
	}
	return &Auth{db: db, config: config, failures: map[string]failure{}}, nil
}
func (a *Auth) Login(username, password, address string) (string, error) {
	if username == "" || password == "" || len(password) > 1024 {
		return "", &Error{"INVALID_CREDENTIALS", 401}
	}
	a.mu.Lock()
	entry := a.failures[address]
	if time.Since(entry.at) > 15*time.Minute {
		entry = failure{}
	}
	blocked := entry.count >= 6
	a.mu.Unlock()
	if blocked {
		return "", &Error{"LOGIN_RATE_LIMITED", 429}
	}
	salt, _ := hex.DecodeString(a.config.Salt)
	expected, _ := hex.DecodeString(a.config.PasswordHash)
	digest := pbkdf2.Key([]byte(password), salt, 260000, 32, sha256.New)
	valid := subtle.ConstantTimeCompare(digest, expected) & subtle.ConstantTimeCompare([]byte(username), []byte(a.config.Username))
	a.mu.Lock()
	defer a.mu.Unlock()
	if valid != 1 {
		if len(a.failures) > 1024 {
			for key := range a.failures {
				delete(a.failures, key)
				break
			}
		}
		entry.count++
		entry.at = time.Now()
		a.failures[address] = entry
		return "", &Error{"INVALID_CREDENTIALS", 401}
	}
	delete(a.failures, address)
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	_, _ = a.db.Exec("DELETE FROM sessions WHERE expires < ?", time.Now().Unix())
	_, err := a.db.Exec("INSERT INTO sessions(token_hash,expires) VALUES(?,?)", hashToken(token), time.Now().Add(12*time.Hour).Unix())
	return token, err
}
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (a *Auth) Authenticated(r *http.Request) bool {
	cookie, err := r.Cookie("devgraph_session")
	if err != nil {
		return false
	}
	var expiry int64
	err = a.db.QueryRow("SELECT expires FROM sessions WHERE token_hash=?", hashToken(cookie.Value)).Scan(&expiry)
	return err == nil && expiry > time.Now().Unix()
}
func (a *Auth) Logout(r *http.Request) {
	if cookie, err := r.Cookie("devgraph_session"); err == nil {
		_, _ = a.db.Exec("DELETE FROM sessions WHERE token_hash=?", hashToken(cookie.Value))
	}
}
func (a *Auth) Cookie(r *http.Request, token string, clear bool) *http.Cookie {
	path := "/"
	if r.Header.Get("X-Forwarded-Prefix") == "/graph/" {
		path = "/graph/"
	}
	maxAge := 43200
	if clear {
		maxAge = -1
	}
	return &http.Cookie{Name: "devgraph_session", Value: token, Path: path, MaxAge: maxAge, HttpOnly: true, Secure: a.config.Secure, SameSite: http.SameSiteStrictMode}
}
func (a *Auth) AllowedMutation(r *http.Request) bool {
	if r.Header.Get("X-DevGraph-Request") != "1" {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range a.config.Origins {
		if origin == strings.TrimRight(allowed, "/") {
			return true
		}
	}
	return false
}
