package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

func (s *graphQLServer) Signup(ctx context.Context, username string, password string) (*User, error) {
	if password == "" {
		recordVedhSignup("validation_error")
		return nil, errs.New("must provide a password")
	}
	hashed, err := hashPassword(password)
	if err != nil {
		recordVedhSignup("error")
		return nil, errs.Wrap(err)
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	stmt := `
	INSERT INTO "users" (uuid, username, password, last_login_at, last_active_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING uuid, username;
	`
	result, err := s.db.Query(stmt, id, username, hashed, now, now)
	if err != nil {
		if strings.Contains(err.Error(), "username_unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate key value") {
			recordVedhSignup("conflict")
			return nil, errs.New("That username is already taken. Try another one.")
		}
		recordVedhSignup("error")
		return nil, errs.Wrap(err)
	}
	defer result.Close()
	user := &User{}
	for result.Next() {
		if err := result.Scan(&user.ID, &user.Username); err != nil {
			recordVedhSignup("error")
			return nil, errs.Wrap(err)
		}
	}

	t, err := newAuthToken(user, time.Hour*24)
	if err != nil {
		recordVedhSignup("error")
		return nil, errs.Wrap(err)
	}
	user.Token = &t
	recordVedhSignup("success")
	if err := refreshVedhUserActivityMetrics(s.db); err != nil {
		s.loggerFor(ctx).Warn("failed to refresh active-user metrics after signup", "err", err)
	}

	return user, nil
}

func (s *graphQLServer) Login(ctx context.Context, username string, password string) (*User, error) {
	if password == "" {
		recordVedhLoginAttempt("validation_error")
		return nil, errs.New("must provide a password for authentication")
	}
	if username == "" {
		recordVedhLoginAttempt("validation_error")
		return nil, errs.New("must provide a username for authentication")
	}

	// TECHDEBT: enforce uniqueness as a constraint on username in the DB
	q := `SELECT "uuid", "username", "password" FROM "users" WHERE username=$1;`
	rows, err := s.db.Query(q, username)
	if err != nil {
		if err == sql.ErrNoRows {
			recordVedhLoginAttempt("invalid_credentials")
			return nil, fmt.Errorf("user not found")
		}
		recordVedhLoginAttempt("error")
		return nil, fmt.Errorf("failed to find rows: %w", err)
	}

	defer rows.Close()
	user := &User{}
	var hash string
	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.Username, &hash); err != nil {
			s.loggerFor(ctx).Error("failed to scan user at login", "err", err, "username", username)
			recordVedhLoginAttempt("error")
			return nil, errs.Wrap(err)
		}
	}

	// check password validity, return if invalid
	valid := checkPasswordHash(password, hash)
	if !valid {
		recordVedhLoginAttempt("invalid_credentials")
		return nil, errs.New("failed to authenticate")
	}

	// we're valid, so generate a new token and assign it to the user
	now := time.Now().UTC()
	if _, err := s.db.Exec(`UPDATE "users" SET last_login_at = $2, last_active_at = $2 WHERE uuid = $1`, user.ID, now); err != nil {
		s.loggerFor(ctx).Warn("failed to persist user login activity", "err", err, "user_id", user.ID)
	}
	t, err := newAuthToken(user, time.Hour*24)
	if err != nil {
		recordVedhLoginAttempt("error")
		return nil, errs.Wrap(err)
	}

	// set password to blank so we don't return the sensitive material
	user.Password = nil

	// TECHDEBT: replace the old redis cache with a LRU in-mem cache
	// s.lru.Set(user.Username, t, time.Duration(time.Hour*24*14))

	user.Token = &t
	recordVedhLoginAttempt("success")
	if err := refreshVedhUserActivityMetrics(s.db); err != nil {
		s.loggerFor(ctx).Warn("failed to refresh active-user metrics after login", "err", err)
	}

	return user, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *graphQLServer) Users(ctx context.Context, id *string) ([]string, error) {
	if !isPublicQuery("users") {
		if _, err := requireAuth(ctx); err != nil {
			return nil, err
		}
	}
	limit := 10000
	offset := 0
	rows, err := s.db.Query(`select * from users limit $1 offset $2;`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	users := []string{}
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, s)
	}
	return users, nil
}
