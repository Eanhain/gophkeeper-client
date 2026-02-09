package entity

// LoginPassword represents a stored login-password pair.
type LoginPassword struct {
	UserID   int    `json:"user_id" db:"user_id"`
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Label    string `json:"label" db:"label"`
}

// TextSecret represents a stored text note.
type TextSecret struct {
	UserID int    `json:"user_id" db:"user_id"`
	Title  string `json:"title" db:"title"`
	Body   string `json:"body" db:"body"`
}

// BinarySecret represents a stored binary blob.
type BinarySecret struct {
	UserID   int    `json:"user_id" db:"user_id"`
	Filename string `json:"filename" db:"filename"`
	MimeType string `json:"mime_type" db:"mime_type"`
	Data     string `json:"data" db:"data"`
}

// CardSecret represents a stored bank card.
type CardSecret struct {
	UserID     int    `json:"user_id" db:"user_id"`
	Cardholder string `json:"cardholder" db:"cardholder"`
	Pan        string `json:"pan" db:"pan"`
	ExpMonth   string `json:"exp_month" db:"exp_month"`
	ExpYear    string `json:"exp_year" db:"exp_year"`
	Brand      string `json:"brand" db:"brand"`
	Last4      string `json:"last4" db:"last4"`
}

// AllSecrets is a composite type aggregating all secret types for a user.
type AllSecrets struct {
	LoginPassword []LoginPassword `json:"login_password" db:"login_password"`
	TextSecret    []TextSecret    `json:"text_secret" db:"text_secret"`
	BinarySecret  []BinarySecret  `json:"binary_secret" db:"binary_secret"`
	CardSecret    []CardSecret    `json:"card_secret" db:"card_secret"`
}
