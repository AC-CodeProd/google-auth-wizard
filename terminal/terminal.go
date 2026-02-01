package terminal

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func New(opts ...Option) *Terminal {
	var (
		defaultListHeight        = 14
		defaultTitleStyle        = lipgloss.NewStyle().MarginLeft(2)
		defaultItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
		defaultSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
		defaultPaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
		defaultHelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
		defaultQuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	)

	t := &Terminal{
		listHeight:        defaultListHeight,
		titleStyle:        defaultTitleStyle,
		itemStyle:         defaultItemStyle,
		selectedItemStyle: defaultSelectedItemStyle,
		paginationStyle:   defaultPaginationStyle,
		helpStyle:         defaultHelpStyle,
		quitTextStyle:     defaultQuitTextStyle,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Terminal) Run(title string, items []Item) ([]string, error) {
	serviceItems := make([]Item, 0)
	serviceMap := make(map[string][]Item)

	for _, item := range items {
		if item.IsHeader {
			serviceItems = append(serviceItems, item)
			serviceMap[item.Title] = item.Children
		}
	}

	listItems := make([]list.Item, len(serviceItems))
	for i, item := range serviceItems {
		listItems[i] = item
	}

	const defaultWidth = 80

	m := &model{
		choice:           make([]string, 0),
		selectedItems:    make(map[int]bool),
		terminal:         t,
		viewState:        ViewServices,
		serviceItems:     serviceItems,
		breadcrumb:       []string{title},
		hasBeenValidated: false,
	}

	t.model = m

	delegate := itemDelegate{
		model: m,
	}

	l := list.New(listItems, delegate, defaultWidth, t.listHeight)
	l.Title = "Select Google APIs"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = t.titleStyle
	l.Styles.PaginationStyle = t.paginationStyle
	l.Styles.HelpStyle = t.helpStyle

	m.list = l

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	finalModel := result.(*model)
	return finalModel.choice, nil
}

func (t *Terminal) HasBeenValidated() bool {
	return t.model.hasBeenValidated
}

func (i Item) FilterValue() string { return i.Title }

func (d itemDelegate) Height() int                               { return 2 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	selected := index == m.Index()

	if i.IsHeader {
		str := i.Title
		if len(i.Children) > 0 {
			str += fmt.Sprintf(" (%d scopes)", len(i.Children))
		}
		if selected {
			_, _ = fmt.Fprint(w, d.model.terminal.selectedItemStyle.Render("> "+str))
		} else {
			_, _ = fmt.Fprint(w, d.model.terminal.titleStyle.Render(str))
		}
	} else {
		var s strings.Builder
		if selected {
			s.WriteString("> ")
		} else {
			s.WriteString("")
		}

		isInChoices := false
		for _, choice := range d.model.choice {
			if choice == i.Value {
				isInChoices = true
				break
			}
		}

		if isInChoices {
			s.WriteString("(•) ")
		} else {
			if i.Value != "confirm" {
				s.WriteString("( ) ")
			}
		}

		s.WriteString(i.Title)

		if len(i.Description) > 60 {
			s.WriteString(fmt.Sprintf("\n      %s", i.Description[:60]+"..."))
		} else if i.Description != "" {
			s.WriteString(fmt.Sprintf("\n      %s", i.Description))
		}

		if selected {
			_, _ = fmt.Fprint(w, d.model.terminal.selectedItemStyle.Render(s.String()))
		} else {
			_, _ = fmt.Fprint(w, d.model.terminal.itemStyle.Render(s.String()))
		}
	}
}
func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			m.list.ResetFilter()

			switch m.viewState {
			case ViewScopes:
				m.viewState = ViewServices
				// Safe breadcrumb manipulation
				if len(m.breadcrumb) > 1 {
					m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
				}

				listItems := make([]list.Item, len(m.serviceItems))
				for i, item := range m.serviceItems {
					listItems[i] = item
				}
				m.list.SetItems(listItems)
				if len(m.breadcrumb) > 0 {
					m.list.Title = m.breadcrumb[len(m.breadcrumb)-1]
				}

			case ViewConfirm:
				m.viewState = ViewScopes
				if len(m.breadcrumb) > 1 {
					m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
				}

				listItems := make([]list.Item, len(m.scopeItems))
				for i, item := range m.scopeItems {
					listItems[i] = item
				}
				m.list.SetItems(listItems)
				if len(m.breadcrumb) > 0 {
					m.list.Title = m.breadcrumb[len(m.breadcrumb)-1]
				}
			}
			return m, nil

		case "tab":
			if m.viewState == ViewServices {
				i, ok := m.list.SelectedItem().(Item)
				if ok && i.IsHeader && len(i.Children) > 0 {
					m.list.ResetFilter()

					m.viewState = ViewScopes
					m.currentService = i.Title
					m.scopeItems = i.Children
					m.breadcrumb = append(m.breadcrumb, i.Title)

					listItems := make([]list.Item, len(i.Children))
					for idx, child := range i.Children {
						listItems[idx] = child
					}
					m.list.SetItems(listItems)
					m.list.Title = i.Title
					m.list.ResetSelected()

					m.selectedItems = make(map[int]bool)
				}
			}
			return m, nil

		case "enter":
			switch m.viewState {
			case ViewServices:
				if len(m.choice) > 0 {
					m.list.ResetFilter()
					m.viewState = ViewConfirm
					m.breadcrumb = append(m.breadcrumb, "Confirm Selection")

					confirmItems := make([]list.Item, len(m.choice)+1)
					for idx, choice := range m.choice {
						confirmItems[idx] = Item{
							Title:       choice,
							Description: "Selected scope",
							Value:       choice,
							IsHeader:    false,
						}
					}
					confirmItems[len(m.choice)] = Item{
						Title:       "✓ Confirm Selection",
						Description: "Press Enter to confirm",
						Value:       "confirm",
						IsHeader:    false,
					}

					listItems := make([]list.Item, len(confirmItems))
					copy(listItems, confirmItems)

					m.list.SetItems(listItems)
					m.list.Title = "Confirm Selection"
					m.list.ResetSelected()
				}

			case ViewScopes:
				if len(m.choice) > 0 {
					m.list.ResetFilter()

					m.viewState = ViewConfirm
					m.breadcrumb = append(m.breadcrumb, "Confirm Selection")

					confirmItems := make([]list.Item, len(m.choice)+1)
					for idx, choice := range m.choice {
						confirmItems[idx] = Item{
							Title:       choice,
							Description: "Selected scope",
							Value:       choice,
							IsHeader:    false,
						}
					}
					confirmItems[len(m.choice)] = Item{
						Title:       "✓ Confirm Selection",
						Description: "Press Enter to confirm",
						Value:       "confirm",
						IsHeader:    false,
					}

					listItems := make([]list.Item, len(confirmItems))
					copy(listItems, confirmItems)
					m.list.SetItems(listItems)
					m.list.Title = "Confirm Selection"
					m.list.ResetSelected()
				}

			case ViewConfirm:
				i, ok := m.list.SelectedItem().(Item)
				if ok && i.Value == "confirm" {
					m.quitting = true
					m.hasBeenValidated = true
					return m, tea.Quit
				}
			}
			return m, nil

		case " ":
			if m.viewState == ViewScopes {
				i, ok := m.list.SelectedItem().(Item)
				if ok && !i.IsHeader && i.Value != "" {
					if m.isSelected(i.Value) {
						for idx, choice := range m.choice {
							if choice == i.Value {
								m.choice = append(m.choice[:idx], m.choice[idx+1:]...)
								break
							}
						}
					} else {
						m.choice = append(m.choice, i.Value)
					}
				}
			}
			return m, nil

		case "ctrl+l":
			m.list.ResetFilter()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return m.terminal.quitTextStyle.Render("Selected scopes saved!")
	}

	breadcrumbStr := strings.Join(m.breadcrumb, " > ")

	var status string
	switch m.viewState {
	case ViewServices:
		if len(m.choice) > 0 {
			status = fmt.Sprintf("Selected: %d scopes | Tab to enter service | Enter to confirm | Type to filter | q to quit", len(m.choice))
		} else {
			status = "Tab to enter service | Type to filter | q to quit"
		}
	case ViewScopes:
		status = fmt.Sprintf("Selected: %d scopes | Space to select/deselect | Enter to confirm | Type to filter | Ctrl+L to clear filter | Esc to go back | q to quit", len(m.choice))
	case ViewConfirm:
		status = "Enter to confirm | Esc to go back | q to quit"
	}

	if m.list.FilterState() == list.Filtering {
		status = "Filtering... | Esc to cancel | Enter to apply"
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s\n",
		m.terminal.titleStyle.Render(breadcrumbStr),
		m.list.View(),
		status)
}

func (m *model) isSelected(value string) bool {
	for _, choice := range m.choice {
		if choice == value {
			return true
		}
	}
	return false
}

func DefaultTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("205"))
}

func DefaultSelectedItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Bold(true)
}

func Color(color string) lipgloss.Color {
	return lipgloss.Color(color)
}

func WithListHeight(listHeight int) Option {
	return func(e *Terminal) {
		e.listHeight = listHeight
	}
}

func WithTitleStyle(titleStyle lipgloss.Style) Option {
	return func(e *Terminal) {
		e.titleStyle = titleStyle
	}
}

func WithItemStyle(itemStyle lipgloss.Style) Option {
	return func(e *Terminal) {
		e.itemStyle = itemStyle
	}
}

func WithSelectedItemStyle(selectedItemStyle lipgloss.Style) Option {
	return func(e *Terminal) {
		e.selectedItemStyle = selectedItemStyle
	}
}

func WithPaginationStyle(paginationStyle lipgloss.Style) Option {
	return func(e *Terminal) {
		e.paginationStyle = paginationStyle
	}
}

func WithHelpStyle(helpStyle lipgloss.Style) Option {
	return func(e *Terminal) {
		e.helpStyle = helpStyle
	}
}

func WithQuitTextStyle(quitTextStyle lipgloss.Style) Option {
	return func(e *Terminal) {
		e.quitTextStyle = quitTextStyle
	}
}
