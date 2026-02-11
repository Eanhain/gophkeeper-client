// Package tui implements the terminal user interface (TUI) for the
// GophKeeper client using the Bubble Tea framework (Elm architecture).
//
// Architecture:
//
//	Model.Init()   → initial command (blink cursor)
//	Model.Update() → process a message (key press, server response, error)
//	Model.View()   → render the current screen to a string
//
// Screens:
//
//	screenAuth – login / register form
//	screenMenu – main menu with available actions
//	screenForm – dynamic form for creating or deleting a secret
//	screenView – read-only table displaying all user secrets
//
// Navigation:
//
//	tab/shift-tab – switch between input fields
//	enter         – submit the current form or select a menu item
//	esc           – go back to the previous screen
//	ctrl+r        – toggle register / login on the auth screen
//	ctrl+c / q    – quit the application
package tui

import (
	"fmt"
	"strings"

	"github.com/Eanhain/gophkeeper-client/contracts/request"
	"github.com/Eanhain/gophkeeper-client/contracts/response"
	"github.com/Eanhain/gophkeeper-client/internal/usecase"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// screen identifies which view is currently displayed.
type screen int

const (
	screenAuth screen = iota // login/register form
	screenMenu               // main menu
	screenForm               // create/delete form
	screenView               // all-secrets table
)

// secretType identifies the secret kind that a form is targeting.
type secretType int

const (
	secretLoginPassword secretType = iota
	secretText
	secretBinary
	secretCard
	deleteLoginPassword
	deleteText
	deleteBinary
	deleteCard
)

// menuAction identifies the action chosen from the main menu.
type menuAction int

const (
	actionViewAll menuAction = iota
	actionAddLP
	actionAddText
	actionAddBinary
	actionAddCard
	actionDeleteLP
	actionDeleteText
	actionDeleteBinary
	actionDeleteCard
	actionResetCache
	actionExit
)

// menuItem describes a single entry in the main menu.
type menuItem struct {
	label    string
	action   menuAction
	needsNet bool // true = hidden in offline mode
}

// menuItems defines the labels and actions shown in the main menu.
var menuItems = []menuItem{
	{"View All Secrets", actionViewAll, false},
	{"Add Login/Password", actionAddLP, true},
	{"Add Text Secret", actionAddText, true},
	{"Add Binary Secret", actionAddBinary, true},
	{"Add Card Secret", actionAddCard, true},
	{"Delete Login/Password", actionDeleteLP, true},
	{"Delete Text Secret", actionDeleteText, true},
	{"Delete Binary Secret", actionDeleteBinary, true},
	{"Delete Card Secret", actionDeleteCard, true},
	{"Reset Cache", actionResetCache, false},
	{"Exit", actionExit, false},
}

// visibleMenuItems returns menu items available in the current mode.
// In offline mode, items requiring network (add/delete) are hidden.
func visibleMenu(offline bool) []menuItem {
	if !offline {
		return menuItems
	}
	var items []menuItem
	for _, it := range menuItems {
		if !it.needsNet {
			items = append(items, it)
		}
	}
	return items
}

// --- Bubble Tea messages ---

// tokenMsg is sent when login/register succeeds.
type tokenMsg struct{ token string }

// errMsg is sent when any async operation fails.
type errMsg struct{ err error }

// secretsMsg is sent when secrets are fetched from cache or server.
type secretsMsg struct{ secrets *response.AllSecrets }

// okMsg is sent on successful write/delete/cache-reset.
type okMsg struct{ message string }

// offlineOkMsg is sent when the cache has data for offline mode.
type offlineOkMsg struct{ secrets *response.AllSecrets }

// offlineFailMsg is sent when offline mode cannot be entered.
type offlineFailMsg struct{ reason string }

// --- lipgloss styles ---

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	menuPad       = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	secTitle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111")).MarginTop(1)
	secItem       = lipgloss.NewStyle().PaddingLeft(2)
)

// --- Model ---

// Model is the Bubble Tea model. It holds the entire UI state:
// current screen, input fields, menu cursor position, loaded secrets,
// status messages, and a reference to the business-logic UseCase.
type Model struct {
	screen screen
	uc     *usecase.UseCase

	// Auth screen
	authInputs []textinput.Model
	authFocus  int
	register   bool // true = register mode, false = login mode

	// Menu screen
	menuCursor int

	// Form screen (create or delete secret)
	formInputs []textinput.Model
	formFocus  int
	formTitle  string
	formType   secretType

	// View screen
	secrets *response.AllSecrets

	// Status bar
	message   string
	messageOk bool
	loading   bool

	// Offline mode (read-only, from cache)
	offline bool
}

// New creates the initial TUI model starting on the auth screen.
func New(uc *usecase.UseCase) Model {
	m := Model{screen: screenAuth, uc: uc}
	m.initAuth()
	return m
}

// initAuth prepares the login and password text inputs for the auth screen.
func (m *Model) initAuth() {
	login := textinput.New()
	login.Placeholder = "login"
	login.CharLimit = 64
	login.Focus()

	password := textinput.New()
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword
	password.CharLimit = 64

	m.authInputs = []textinput.Model{login, password}
	m.authFocus = 0
}

// Init returns the initial command (cursor blink).
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// --- Async commands (run in a goroutine by Bubble Tea) ---

// loginCmd sends login credentials to the server.
func (m Model) loginCmd() tea.Cmd {
	login, pass, uc := m.authInputs[0].Value(), m.authInputs[1].Value(), m.uc
	return func() tea.Msg {
		token, err := uc.Login(login, pass)
		if err != nil {
			return errMsg{err}
		}
		return tokenMsg{token}
	}
}

// registerCmd registers a new account and logs in.
func (m Model) registerCmd() tea.Cmd {
	login, pass, uc := m.authInputs[0].Value(), m.authInputs[1].Value(), m.uc
	return func() tea.Msg {
		token, err := uc.Register(login, pass)
		if err != nil {
			return errMsg{err}
		}
		return tokenMsg{token}
	}
}

// fetchSecretsCmd fetches all secrets (cache-first strategy).
func (m Model) fetchSecretsCmd() tea.Cmd {
	uc := m.uc
	return func() tea.Msg {
		secrets, err := uc.GetAllSecrets()
		if err != nil {
			return errMsg{err}
		}
		return secretsMsg{secrets}
	}
}

// submitFormCmd sends the filled-in form data to the server.
func (m Model) submitFormCmd() tea.Cmd {
	ft := m.formType
	vals := make([]string, len(m.formInputs))
	for i, inp := range m.formInputs {
		vals[i] = inp.Value()
	}
	uc := m.uc
	return func() tea.Msg {
		var err error
		switch ft {
		case secretLoginPassword:
			err = uc.AddLoginPassword(request.LoginPassword{
				Login: vals[0], Password: vals[1], Label: vals[2],
			})
		case secretText:
			err = uc.AddTextSecret(request.TextSecret{
				Title: vals[0], Body: vals[1],
			})
		case secretBinary:
			err = uc.AddBinarySecret(request.BinarySecret{
				Filename: vals[0], MimeType: vals[1], Data: vals[2],
			})
		case secretCard:
			err = uc.AddCardSecret(request.CardSecret{
				Cardholder: vals[0], Pan: vals[1],
				ExpMonth: vals[2], ExpYear: vals[3],
				Brand: vals[4], Last4: vals[5],
			})
		case deleteLoginPassword:
			err = uc.DeleteLoginPassword(vals[0])
		case deleteText:
			err = uc.DeleteTextSecret(vals[0])
		case deleteBinary:
			err = uc.DeleteBinarySecret(vals[0])
		case deleteCard:
			err = uc.DeleteCardSecret(vals[0])
		}
		if err != nil {
			return errMsg{err}
		}
		return okMsg{"Done"}
	}
}

// tryOfflineCmd checks whether the local cache contains data.
// If yes, sends offlineOkMsg; if wrong key, sends offlineFailMsg with
// an appropriate message; otherwise reports empty cache.
func (m Model) tryOfflineCmd() tea.Cmd {
	uc := m.uc
	return func() tea.Msg {
		if uc.IsWrongKey() {
			return offlineFailMsg{reason: "Wrong CRYPTO_KEY — cannot decrypt local cache"}
		}
		if cached := uc.GetCachedSecrets(); cached != nil {
			return offlineOkMsg{cached}
		}
		return offlineFailMsg{reason: "Cache is empty — offline mode unavailable"}
	}
}

// resetCacheCmd clears the local encrypted cache.
func (m Model) resetCacheCmd() tea.Cmd {
	uc := m.uc
	return func() tea.Msg {
		uc.ResetCache()
		return okMsg{"Cache cleared"}
	}
}

// --- Update (message handler) ---

// Update processes Bubble Tea messages and returns an updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case errMsg:
		m.message = msg.err.Error()
		m.messageOk = false
		m.loading = false
		return m, nil
	case tokenMsg:
		m.uc.SetToken(msg.token)
		m.screen = screenMenu
		m.message = "Authenticated"
		m.messageOk = true
		m.loading = false
		return m, nil
	case secretsMsg:
		m.secrets = msg.secrets
		m.screen = screenView
		m.loading = false
		return m, nil
	case okMsg:
		m.message = msg.message
		m.messageOk = true
		m.screen = screenMenu
		m.loading = false
		return m, nil
	case offlineOkMsg:
		m.offline = true
		m.screen = screenMenu
		m.menuCursor = 0
		m.message = "Offline mode (read-only, from cache)"
		m.messageOk = true
		m.loading = false
		return m, nil
	case offlineFailMsg:
		m.message = msg.reason
		m.messageOk = false
		m.loading = false
		return m, nil
	}

	switch m.screen {
	case screenAuth:
		return m.updateAuth(msg)
	case screenMenu:
		return m.updateMenu(msg)
	case screenForm:
		return m.updateForm(msg)
	case screenView:
		return m.updateView(msg)
	}
	return m, nil
}

// updateAuth handles key presses on the login/register screen.
func (m Model) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyTab, tea.KeyShiftTab:
			m.authFocus = (m.authFocus + 1) % len(m.authInputs)
			for i := range m.authInputs {
				if i == m.authFocus {
					m.authInputs[i].Focus()
				} else {
					m.authInputs[i].Blur()
				}
			}
			return m, nil
		case tea.KeyEnter:
			if m.authInputs[0].Value() == "" || m.authInputs[1].Value() == "" {
				m.message = "Login and password required"
				m.messageOk = false
				return m, nil
			}
			m.loading = true
			m.message = ""
			if m.register {
				return m, m.registerCmd()
			}
			return m, m.loginCmd()
		case tea.KeyCtrlR:
			m.register = !m.register
			return m, nil
		case tea.KeyCtrlO:
			m.loading = true
			m.message = ""
			return m, m.tryOfflineCmd()
		}
	}

	var cmds []tea.Cmd
	for i := range m.authInputs {
		var cmd tea.Cmd
		m.authInputs[i], cmd = m.authInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// updateMenu handles key presses on the main menu.
func (m Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := visibleMenu(m.offline)
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyUp:
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case tea.KeyDown:
			if m.menuCursor < len(items)-1 {
				m.menuCursor++
			}
		case tea.KeyEnter:
			return m.handleMenuAction(items[m.menuCursor].action)
		}
		if key.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// handleMenuAction dispatches the selected menu item to the
// corresponding screen or command.
func (m Model) handleMenuAction(action menuAction) (tea.Model, tea.Cmd) {
	m.message = ""
	switch action {
	case actionViewAll:
		m.loading = true
		return m, m.fetchSecretsCmd()
	case actionAddLP:
		m.setupForm("Add Login/Password", secretLoginPassword,
			[]string{"Login", "Password", "Label"})
	case actionAddText:
		m.setupForm("Add Text Secret", secretText,
			[]string{"Title", "Body"})
	case actionAddBinary:
		m.setupForm("Add Binary Secret", secretBinary,
			[]string{"Filename", "MIME Type", "Data (base64)"})
	case actionAddCard:
		m.setupForm("Add Card Secret", secretCard,
			[]string{"Cardholder", "PAN", "Exp Month", "Exp Year", "Brand", "Last4"})
	case actionDeleteLP:
		m.setupForm("Delete Login/Password", deleteLoginPassword,
			[]string{"Login to delete"})
	case actionDeleteText:
		m.setupForm("Delete Text Secret", deleteText,
			[]string{"Title to delete"})
	case actionDeleteBinary:
		m.setupForm("Delete Binary Secret", deleteBinary,
			[]string{"Filename to delete"})
	case actionDeleteCard:
		m.setupForm("Delete Card Secret", deleteCard,
			[]string{"Cardholder to delete"})
	case actionResetCache:
		return m, m.resetCacheCmd()
	case actionExit:
		return m, tea.Quit
	}
	return m, textinput.Blink
}

// setupForm initialises the form screen with the given title,
// secret type and placeholder labels for each input field.
func (m *Model) setupForm(title string, ft secretType, placeholders []string) {
	m.formTitle = title
	m.formType = ft
	m.formFocus = 0
	m.formInputs = make([]textinput.Model, len(placeholders))
	for i, ph := range placeholders {
		ti := textinput.New()
		ti.Placeholder = ph
		ti.CharLimit = 256
		if i == 0 {
			ti.Focus()
		}
		m.formInputs[i] = ti
	}
	m.screen = screenForm
}

// updateForm handles key presses on the create/delete form.
func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyEsc:
			m.screen = screenMenu
			return m, nil
		case tea.KeyTab:
			m.formFocus = (m.formFocus + 1) % len(m.formInputs)
			for i := range m.formInputs {
				if i == m.formFocus {
					m.formInputs[i].Focus()
				} else {
					m.formInputs[i].Blur()
				}
			}
			return m, nil
		case tea.KeyShiftTab:
			m.formFocus = (m.formFocus - 1 + len(m.formInputs)) % len(m.formInputs)
			for i := range m.formInputs {
				if i == m.formFocus {
					m.formInputs[i].Focus()
				} else {
					m.formInputs[i].Blur()
				}
			}
			return m, nil
		case tea.KeyEnter:
			for _, inp := range m.formInputs {
				if inp.Value() == "" {
					m.message = "All fields are required"
					m.messageOk = false
					return m, nil
				}
			}
			m.loading = true
			m.message = ""
			return m, m.submitFormCmd()
		}
	}

	var cmds []tea.Cmd
	for i := range m.formInputs {
		var cmd tea.Cmd
		m.formInputs[i], cmd = m.formInputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// updateView handles key presses on the read-only secrets view.
func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.Type == tea.KeyEsc {
			m.screen = screenMenu
			return m, nil
		}
	}
	return m, nil
}

// --- View (render) ---

// View renders the current screen to a string.
func (m Model) View() string {
	if m.loading {
		return "\n  Loading...\n"
	}

	var s strings.Builder

	switch m.screen {
	case screenAuth:
		s.WriteString(m.viewAuth())
	case screenMenu:
		s.WriteString(m.viewMenu())
	case screenForm:
		s.WriteString(m.viewForm())
	case screenView:
		s.WriteString(m.viewSecrets())
	}

	if m.message != "" {
		s.WriteString("\n")
		if m.messageOk {
			s.WriteString(okStyle.Render("  + " + m.message))
		} else {
			s.WriteString(errStyle.Render("  ! " + m.message))
		}
		s.WriteString("\n")
	}

	return s.String()
}

// viewAuth renders the login/register screen.
func (m Model) viewAuth() string {
	var s strings.Builder
	mode := "Login"
	if m.register {
		mode = "Register"
	}
	s.WriteString(titleStyle.Render(fmt.Sprintf("  GophKeeper  -  %s", mode)))
	s.WriteString("\n\n")
	for i, inp := range m.authInputs {
		if i == m.authFocus {
			s.WriteString(selectedStyle.Render("  > "))
		} else {
			s.WriteString("    ")
		}
		s.WriteString(inp.View())
		s.WriteString("\n")
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  enter: submit | tab: switch field | ctrl+r: toggle register/login | ctrl+o: offline | ctrl+c: quit"))
	return s.String()
}

// viewMenu renders the main menu with a cursor indicator.
func (m Model) viewMenu() string {
	items := visibleMenu(m.offline)
	var s strings.Builder
	title := "  GophKeeper  -  Menu"
	if m.offline {
		title = "  GophKeeper  -  Menu (offline, read-only)"
	}
	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")
	for i, item := range items {
		cursor := "  "
		style := normalStyle
		if i == m.menuCursor {
			cursor = "> "
			style = selectedStyle
		}
		s.WriteString(menuPad.Render(style.Render(cursor + item.label)))
		s.WriteString("\n")
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  up/down: navigate | enter: select | q: quit"))
	return s.String()
}

// viewForm renders the dynamic create/delete form.
func (m Model) viewForm() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render(fmt.Sprintf("  %s", m.formTitle)))
	s.WriteString("\n\n")
	for i, inp := range m.formInputs {
		if i == m.formFocus {
			s.WriteString(selectedStyle.Render("  > "))
		} else {
			s.WriteString("    ")
		}
		s.WriteString(inp.View())
		s.WriteString("\n")
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  enter: submit | tab/shift+tab: switch field | esc: back"))
	return s.String()
}

// viewSecrets renders the read-only list of all user secrets.
func (m Model) viewSecrets() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("  All Secrets"))
	s.WriteString("\n")

	if m.secrets == nil {
		s.WriteString("\n  No secrets found\n")
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("  esc: back"))
		return s.String()
	}

	empty := true

	if len(m.secrets.LoginPassword) > 0 {
		empty = false
		s.WriteString(secTitle.Render("  Login/Passwords"))
		s.WriteString("\n")
		for _, lp := range m.secrets.LoginPassword {
			s.WriteString(secItem.Render(fmt.Sprintf(
				"login: %s  |  password: %s  |  label: %s",
				lp.Login, lp.Password, lp.Label)))
			s.WriteString("\n")
		}
	}

	if len(m.secrets.TextSecret) > 0 {
		empty = false
		s.WriteString(secTitle.Render("  Text Secrets"))
		s.WriteString("\n")
		for _, ts := range m.secrets.TextSecret {
			s.WriteString(secItem.Render(fmt.Sprintf(
				"title: %s  |  body: %s",
				ts.Title, ts.Body)))
			s.WriteString("\n")
		}
	}

	if len(m.secrets.BinarySecret) > 0 {
		empty = false
		s.WriteString(secTitle.Render("  Binary Secrets"))
		s.WriteString("\n")
		for _, bs := range m.secrets.BinarySecret {
			s.WriteString(secItem.Render(fmt.Sprintf(
				"file: %s  |  type: %s",
				bs.Filename, bs.MimeType)))
			s.WriteString("\n")
		}
	}

	if len(m.secrets.CardSecret) > 0 {
		empty = false
		s.WriteString(secTitle.Render("  Card Secrets"))
		s.WriteString("\n")
		for _, cs := range m.secrets.CardSecret {
			s.WriteString(secItem.Render(fmt.Sprintf(
				"holder: %s  |  pan: %s  |  exp: %s/%s  |  brand: %s",
				cs.Cardholder, cs.Pan, cs.ExpMonth, cs.ExpYear, cs.Brand)))
			s.WriteString("\n")
		}
	}

	if empty {
		s.WriteString("\n  No secrets stored yet\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  esc: back"))
	return s.String()
}
