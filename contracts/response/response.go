// Package response defines the JSON structures received FROM the server.
// These mirror the server's response DTOs in
// gophkeeper/internal/controller/restapi/v1/response.
//
// The From* converter functions transform internal entity types into
// response DTOs, stripping fields that must not be exposed (e.g. user_id).
package response

import "github.com/Eanhain/gophkeeper-client/internal/entity"

// LoginPassword is a stored credential pair.
type LoginPassword struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Label    string `json:"label" db:"label"`
}

// TextSecret is a stored text note.
type TextSecret struct {
	Title string `json:"title" db:"title"`
	Body  string `json:"body" db:"body"`
}

// BinarySecret is a stored binary blob (base64-encoded).
type BinarySecret struct {
	Filename string `json:"filename" db:"filename"`
	MimeType string `json:"mime_type" db:"mime_type"`
	Data     string `json:"data" db:"data"`
}

// CardSecret is a stored bank card.
type CardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
	Pan        string `json:"pan" db:"pan"`
	ExpMonth   string `json:"exp_month" db:"exp_month"`
	ExpYear    string `json:"exp_year" db:"exp_year"`
	Brand      string `json:"brand" db:"brand"`
	Last4      string `json:"last4" db:"last4"`
}

// AllSecrets aggregates all secret types returned by get-all-secrets.
type AllSecrets struct {
	LoginPassword []LoginPassword `json:"login_password" db:"login_password"`
	TextSecret    []TextSecret    `json:"text_secret" db:"text_secret"`
	BinarySecret  []BinarySecret  `json:"binary_secret" db:"binary_secret"`
	CardSecret    []CardSecret    `json:"card_secret" db:"card_secret"`
}

// --- Entity → Response converters ---

// FromLoginPassword converts an entity to a response DTO.
func FromLoginPassword(value entity.LoginPassword) LoginPassword {
	return LoginPassword{
		Login:    value.Login,
		Password: value.Password,
		Label:    value.Label,
	}
}

// FromTextSecret converts an entity to a response DTO.
func FromTextSecret(value entity.TextSecret) TextSecret {
	return TextSecret{
		Title: value.Title,
		Body:  value.Body,
	}
}

// FromBinarySecret converts an entity to a response DTO.
func FromBinarySecret(value entity.BinarySecret) BinarySecret {
	return BinarySecret{
		Filename: value.Filename,
		MimeType: value.MimeType,
		Data:     value.Data,
	}
}

// FromCardSecret converts an entity to a response DTO.
func FromCardSecret(value entity.CardSecret) CardSecret {
	return CardSecret{
		Cardholder: value.Cardholder,
		Pan:        value.Pan,
		ExpMonth:   value.ExpMonth,
		ExpYear:    value.ExpYear,
		Brand:      value.Brand,
		Last4:      value.Last4,
	}
}

// FromLoginPasswords converts a slice of entities to response DTOs.
func FromLoginPasswords(values []entity.LoginPassword) []LoginPassword {
	result := make([]LoginPassword, 0, len(values))
	for _, value := range values {
		result = append(result, FromLoginPassword(value))
	}
	return result
}

// FromTextSecrets converts a slice of entities to response DTOs.
func FromTextSecrets(values []entity.TextSecret) []TextSecret {
	result := make([]TextSecret, 0, len(values))
	for _, value := range values {
		result = append(result, FromTextSecret(value))
	}
	return result
}

// FromBinarySecrets converts a slice of entities to response DTOs.
func FromBinarySecrets(values []entity.BinarySecret) []BinarySecret {
	result := make([]BinarySecret, 0, len(values))
	for _, value := range values {
		result = append(result, FromBinarySecret(value))
	}
	return result
}

// FromCardSecrets converts a slice of entities to response DTOs.
func FromCardSecrets(values []entity.CardSecret) []CardSecret {
	result := make([]CardSecret, 0, len(values))
	for _, value := range values {
		result = append(result, FromCardSecret(value))
	}
	return result
}

// FromAllSecrets converts the combined entity to a response DTO.
func FromAllSecrets(values entity.AllSecrets) AllSecrets {
	return AllSecrets{
		LoginPassword: FromLoginPasswords(values.LoginPassword),
		TextSecret:    FromTextSecrets(values.TextSecret),
		BinarySecret:  FromBinarySecrets(values.BinarySecret),
		CardSecret:    FromCardSecrets(values.CardSecret),
	}
}
