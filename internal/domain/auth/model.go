package auth

import "time"

type RefreshToken struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	TokenHash  string    `db:"token_hash"`
	DeviceInfo *string   `db:"device_info"`
	IPAddress  *string   `db:"ip_address"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
	Revoked    bool      `db:"revoked"`
}

type UserVerificationToken struct {
	ID         int64     `db:"id"`
	UserAuthID int64     `db:"user_auth_id"`
	TokenHash  string    `db:"token_hash"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
	LastSendAt time.Time `db:"last_send_at"`
}
