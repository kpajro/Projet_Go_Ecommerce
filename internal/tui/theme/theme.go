// Package theme centralise tous les styles lipgloss de l'application.
//
// Règle du projet : aucun écran ne code une couleur en dur. Une couleur ou
// une bordure se définit ici, puis se réutilise via les champs de Theme.
package theme

import "github.com/charmbracelet/lipgloss"

// Palette : couleurs adaptatives (une valeur pour le fond clair, une pour le
// fond sombre), ce qui évite l'inversion de lisibilité dans un terminal.
var (
	Primary = lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#7AA2F7"}
	Success = lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#9ECE6A"}
	Warning = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#E0AF68"}
	Danger  = lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#F7768E"}
	Muted   = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#A9B1D6"}
	Accent  = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#BB9AF7"}
)

// Theme regroupe les styles réutilisables par les écrans bubbletea.
type Theme struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Label    lipgloss.Style
	Value    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Error    lipgloss.Style
	Help     lipgloss.Style
	Box      lipgloss.Style
}

// New construit le thème par défaut de l application's.
func New() Theme {
	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(Primary).Padding(0, 1),
		Subtitle: lipgloss.NewStyle().Foreground(Muted),
		Label:    lipgloss.NewStyle().Bold(true).Foreground(Accent),
		Value:    lipgloss.NewStyle(),
		Success:  lipgloss.NewStyle().Foreground(Success),
		Warning:  lipgloss.NewStyle().Foreground(Warning),
		Error:    lipgloss.NewStyle().Foreground(Danger),
		Help:     lipgloss.NewStyle().Foreground(Muted),
		Box:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Primary).Padding(0, 1),
	}
}
