// Package entity defines the client-side domain models.
// These are local copies of the server entities, used for internal
// transformations (e.g. response → entity → cache).
package entity

// User represents a user record as stored in the server database.
type User struct {
	Login string `json:"login" db:"username"`
	Hash  string `json:"hash" db:"password_hash"`
}

// UserInput contains the raw credentials entered by the user.
type UserInput struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}
