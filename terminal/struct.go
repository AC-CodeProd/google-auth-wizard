package terminal

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Terminal struct {
	listHeight        int
	titleStyle        lipgloss.Style
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
	paginationStyle   lipgloss.Style
	helpStyle         lipgloss.Style
	quitTextStyle     lipgloss.Style
	model             *model
}

type Option func(*Terminal)

type Item struct {
	Title       string
	Description string
	Value       string
	IsHeader    bool
	Children    []Item
}

type itemDelegate struct {
	model *model
}

type ViewState int

const (
	ViewServices ViewState = iota
	ViewScopes
	ViewConfirm
)

type model struct {
	list             list.Model
	choice           []string
	hasBeenValidated bool
	quitting         bool
	selectedItems    map[int]bool
	terminal         *Terminal
	viewState        ViewState
	currentService   string
	serviceItems     []Item
	scopeItems       []Item
	breadcrumb       []string
}
