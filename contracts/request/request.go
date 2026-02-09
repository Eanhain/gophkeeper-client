// Package request defines the JSON structures sent FROM the client TO the server.
// These match the server's request DTOs in gophkeeper/internal/controller/restapi/v1/request.
package request

// UserInput is sent to POST /api/user/login and /api/user/register.
type UserInput struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

// --- Create (POST) requests ---

// LoginPassword is sent to POST /api/user/secret/post-login-password.
type LoginPassword struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Label    string `json:"label" db:"label"`
}

// TextSecret is sent to POST /api/user/secret/post-text-secret.
type TextSecret struct {
	Title string `json:"title" db:"title"`
	Body  string `json:"body" db:"body"`
}

// BinarySecret is sent to POST /api/user/secret/post-binary-secret.
// Data must be a base64-encoded string.
type BinarySecret struct {
	Filename string `json:"filename" db:"filename"`
	MimeType string `json:"mime_type" db:"mime_type"`
	Data     string `json:"data" db:"data"`
}

// CardSecret is sent to POST /api/user/secret/post-card-secret.
type CardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
	Pan        string `json:"pan" db:"pan"`
	ExpMonth   string `json:"exp_month" db:"exp_month"`
	ExpYear    string `json:"exp_year" db:"exp_year"`
	Brand      string `json:"brand" db:"brand"`
	Last4      string `json:"last4" db:"last4"`
}

// Secret is a composite request containing all four secret types.
type Secret struct {
	Login  LoginPassword `json:"login" db:"login"`
	Text   TextSecret    `json:"text" db:"text"`
	Binary BinarySecret  `json:"binary" db:"binary"`
	Card   CardSecret    `json:"card" db:"card"`
}

// --- Delete requests ---

// DeleteLoginPassword is sent to DELETE /api/user/secret/delete-login-password.
type DeleteLoginPassword struct {
	Login string `json:"login" db:"login"`
}

// DeleteTextSecret is sent to DELETE /api/user/secret/delete-text-secret.
type DeleteTextSecret struct {
	Title string `json:"title" db:"title"`
}

// DeleteBinarySecret is sent to DELETE /api/user/secret/delete-binary-secret.
type DeleteBinarySecret struct {
	Filename string `json:"filename" db:"filename"`
}

// DeleteCardSecret is sent to DELETE /api/user/secret/delete-card-secret.
type DeleteCardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
}

// --- Get requests ---

// GetLoginPassword is sent to GET /api/user/secret/get-login-password.
type GetLoginPassword struct {
	Login string `json:"login" db:"login"`
}

// GetTextSecret is sent to GET /api/user/secret/get-text-secret.
type GetTextSecret struct {
	Title string `json:"title" db:"title"`
}

// GetBinarySecret is sent to GET /api/user/secret/get-binary-secret.
type GetBinarySecret struct {
	Filename string `json:"filename" db:"filename"`
}

// GetCardSecret is sent to GET /api/user/secret/get-card-secret.
type GetCardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
}
