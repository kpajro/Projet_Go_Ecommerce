// Package widgets regroupe les composants d'interface réutilisables par les
// applications client et administrateur.
package widgets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pajro/pgo/internal/tui/theme"
)

// HealthMsg transporte le résultat de la vérification du serveur.
// C'est le message que renvoie la commande asynchrone checkHealth.
type HealthMsg struct {
	Status string
	Err    error
}

// Boot est l'écran de démarrage : il interroge /healthz et affiche le
// résultat. Il sert de base commune aux applications client et administrateur.
//
// Toute l'architecture repose sur le modèle de bubbletea :
//   - Init  : renvoie les commandes à lancer au démarrage.
//   - Update : traite les messages (clavier, réseau) et renvoie l'état à jour.
//   - View   : produit l'affichage à partir de l'état.
type Boot struct {
	Title   string
	APIURL  string
	theme   theme.Theme
	spin    spinner.Model
	loading bool
	status  string
	err     error
}

// NewBoot construit l'écran de démarrage.
func NewBoot(title, apiURL string) Boot {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	return Boot{
		Title:   title,
		APIURL:  apiURL,
		theme:   theme.New(),
		spin:    spin,
		loading: true,
	}
}

// Init lance le spinner et la vérification du serveur en parallèle.
// tea.Batch exécute plusieurs commandes sans attendre les unes les autres.
func (b Boot) Init() tea.Cmd {
	return tea.Batch(b.spin.Tick, b.checkHealth())
}

// Update traite les messages entrants.
func (b Boot) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return b, tea.Quit
		}

	case HealthMsg:
		b.loading = false
		b.status = msg.Status
		b.err = msg.Err

	case spinner.TickMsg:
		var cmd tea.Cmd
		b.spin, cmd = b.spin.Update(msg)
		return b, cmd
	}
	return b, nil
}

// View produit l'affichage de l'écran.
func (b Boot) View() string {
	header := b.theme.Title.Render(b.Title)

	var body string
	switch {
	case b.loading:
		body = fmt.Sprintf("%s %s", b.spin.View(), "Connexion au serveur…")
	case b.err != nil:
		body = b.theme.Error.Render("Serveur injoignable : " + b.err.Error())
	case b.status == "ok":
		body = b.theme.Success.Render("Serveur opérationnel")
	default:
		body = b.theme.Warning.Render("Statut du serveur : " + b.status)
	}

	help := b.theme.Help.Render("q / ctrl+c : quitter")

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", help)
}

// checkHealth interroge /healthz. En tant que tea.Cmd, elle s'exécute dans
// une goroutine : l'interface reste réactive pendant l'appel réseau.
func (b Boot) checkHealth() tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(b.APIURL + "/healthz")
		if err != nil {
			return HealthMsg{Err: err}
		}
		defer resp.Body.Close()

		var payload struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return HealthMsg{Err: err}
		}
		return HealthMsg{Status: payload.Status}
	}
}
