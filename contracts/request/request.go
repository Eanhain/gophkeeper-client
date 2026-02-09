package request

// Аутентификация пользователя (прием на сервер)
// POST /api/user/login.
type UserInput struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

type LoginPassword struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	Label    string `json:"label" db:"label"`
}

type TextSecret struct {
	Title string `json:"title" db:"title"`
	Body  string `json:"body" db:"body"`
}

type BinarySecret struct {
	Filename string `json:"filename" db:"filename"`
	MimeType string `json:"mime_type" db:"mime_type"`
	Data     string `json:"data" db:"data"`
}

type CardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
	Pan        string `json:"pan" db:"pan"`
	ExpMonth   string `json:"exp_month" db:"exp_month"`
	ExpYear    string `json:"exp_year" db:"exp_year"`
	Brand      string `json:"brand" db:"brand"`
	Last4      string `json:"last4" db:"last4"`
}

type Secret struct {
	Login  LoginPassword `json:"login" db:"login"`
	Text   TextSecret    `json:"text" db:"text"`
	Binary BinarySecret  `json:"binary" db:"binary"`
	Card   CardSecret    `json:"card" db:"card"`
}

// DELETE /api/user/login.

type DeleteLoginPassword struct {
	Login string `json:"login" db:"login"`
}

type DeleteTextSecret struct {
	Title string `json:"title" db:"title"`
}

type DeleteBinarySecret struct {
	Filename string `json:"filename" db:"filename"`
}

type DeleteCardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
}

// GET /api/user/login.
type GetLoginPassword struct {
	Login string `json:"login" db:"login"`
}

type GetTextSecret struct {
	Title string `json:"title" db:"title"`
}

type GetBinarySecret struct {
	Filename string `json:"filename" db:"filename"`
}

type GetCardSecret struct {
	Cardholder string `json:"cardholder" db:"cardholder"`
}
