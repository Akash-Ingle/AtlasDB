package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const apiKeyPrefix = "atlas_"

type APIKey struct {
	KeyID      uuid.UUID  `json:"key_id"`
	UserID     uuid.UUID  `json:"user_id"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type APIKeyWithSecret struct {
	APIKey
	RawKey string `json:"raw_key"`
}

type APIKeyStore struct {
	pool *pgxpool.Pool
}

func NewAPIKeyStore(pool *pgxpool.Pool) *APIKeyStore {
	return &APIKeyStore{pool: pool}
}

func (s *APIKeyStore) Create(ctx context.Context, userID uuid.UUID, name string, scopes []string, expiresAt *time.Time) (*APIKeyWithSecret, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	rawKey := apiKeyPrefix + hex.EncodeToString(rawBytes)
	prefix := rawKey[:len(apiKeyPrefix)+8]

	hash, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash key: %w", err)
	}

	keyID := uuid.New()
	now := time.Now()

	_, err = s.pool.Exec(ctx, `
		INSERT INTO api_keys (key_id, user_id, key_hash, key_prefix, name, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, keyID, userID, string(hash), prefix, name, scopes, expiresAt, now)
	if err != nil {
		return nil, fmt.Errorf("store api key: %w", err)
	}

	return &APIKeyWithSecret{
		APIKey: APIKey{
			KeyID:     keyID,
			UserID:    userID,
			KeyPrefix: prefix,
			Name:      name,
			Scopes:    scopes,
			ExpiresAt: expiresAt,
			CreatedAt: now,
		},
		RawKey: rawKey,
	}, nil
}

func (s *APIKeyStore) Validate(ctx context.Context, rawKey string) (*APIKey, error) {
	if len(rawKey) < len(apiKeyPrefix)+8 {
		return nil, fmt.Errorf("invalid api key format")
	}

	prefix := rawKey[:len(apiKeyPrefix)+8]

	rows, err := s.pool.Query(ctx, `
		SELECT key_id, user_id, key_hash, key_prefix, name, scopes, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE key_prefix = $1
	`, prefix)
	if err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ak APIKey
		var keyHash string
		err := rows.Scan(
			&ak.KeyID, &ak.UserID, &keyHash, &ak.KeyPrefix,
			&ak.Name, &ak.Scopes, &ak.LastUsedAt, &ak.ExpiresAt, &ak.CreatedAt,
		)
		if err != nil {
			continue
		}

		if bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(rawKey)) == nil {
			if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
				return nil, fmt.Errorf("api key expired")
			}

			// Update last_used_at (fire and forget)
			go func() {
				s.pool.Exec(context.Background(),
					"UPDATE api_keys SET last_used_at = $1 WHERE key_id = $2",
					time.Now(), ak.KeyID)
			}()

			return &ak, nil
		}
	}

	return nil, fmt.Errorf("invalid api key")
}

func (s *APIKeyStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]APIKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT key_id, user_id, key_prefix, name, scopes, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var ak APIKey
		err := rows.Scan(
			&ak.KeyID, &ak.UserID, &ak.KeyPrefix,
			&ak.Name, &ak.Scopes, &ak.LastUsedAt, &ak.ExpiresAt, &ak.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, ak)
	}

	return keys, nil
}

func (s *APIKeyStore) Delete(ctx context.Context, keyID, userID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		"DELETE FROM api_keys WHERE key_id = $1 AND user_id = $2",
		keyID, userID)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
