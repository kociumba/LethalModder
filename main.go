package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/kociumba/LethalModder/api"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	err      error
	p        *tea.Program
	finished = make(chan struct{})

	dbg   = flag.Bool("dbg", false, "enable debug logging")
	print = flag.Bool("print", false, "print to stdout")
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func main() {
	flag.Parse()
	log.SetReportCaller(true)
	if *dbg {
		log.SetLevel(log.DebugLevel)
	}

	// This is fucking 182mb
	// wtf were they smoking
	go func() {
		packageListings, err := api.GetMods()
		if err != nil {
			log.Fatalf("Error getting mods: %v", err)
		}

		if *print {
			for _, listing := range packageListings {
				fmt.Println(listing)
			}
			os.Exit(0)
		}

		log.Debugf("Successfully parsed %d package listings\n", len(packageListings))
		for i, listing := range packageListings {
			ratingScore, _ := listing.GetRatingScore()
			isPinned, _ := listing.GetIsPinned()
			categories, _ := listing.GetCategories()
			log.Debugf("%d: Name: %s, Owner: %s, Rating: %d, IsPinned: %v, Categories: %v\n",
				i+1, listing.Name, listing.Owner, ratingScore, isPinned, categories)
			if i >= 9 { // Print only first 10 for brevity
				break
			}
		}

		items := make([]list.Item, 0)
		for _, listing := range packageListings {
			isDeprecated, _ := listing.GetIsDeprecated()
			if !isDeprecated {
				rating, err := listing.GetRatingScore()
				if err != nil {
					log.Errorf("Error getting rating score: %v", err)
				}
				items = append(items, item{title: listing.Name, desc: "Owner: " + listing.Owner + " - " + fmt.Sprint(rating) + "☆"})
			}
		}

		m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
		m.list.Title = "Mods"

		p = tea.NewProgram(m, tea.WithAltScreen())

		// Program is ready
		// Should stop

		finished <- struct{}{}
	}()

	err = spinner.New().Title("Loading mods...").Run()
	if err != nil {
		log.Fatalf("Error starting spinner: %v", err)
	}

	<-finished

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
