package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/kociumba/LethalModder/api"
	"github.com/kociumba/LethalModder/steam"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	err      error
	p        *tea.Program
	M        model

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

	// Need to make sure all of this is loaded before showing output
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		initMods()

	}()

	go func() {
		defer wg.Done()
		initLocalProfiles()
	}()

	err = spinner.New().Title("Loading mods...").Run()
	if err != nil {
		log.Fatalf("Error starting spinner: %v", err)
	}

	wg.Wait()

	// huh.NewSelect[int]().Run()

	// don't wanna render the 20k mods when testing
	if *dbg {
		os.Exit(0)
	}

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// This is fucking 182mb
// wtf were they smoking
func initMods() {
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
	// for i, listing := range packageListings {
	// 	ratingScore, _ := listing.GetRatingScore()
	// 	log.Debugf("%d: Name: %s, Owner: %s, Rating: %d, Link: %s",
	// 		i+1, listing.Name, listing.Owner, ratingScore, listing.PackageURL)
	// 	if i >= 9 { // Print only first 10 for brevity
	// 		break
	// 	}
	// }

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

	M = model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	M.list.Title = "Mods"

	p = tea.NewProgram(M, tea.WithAltScreen())
}

func initLocalProfiles() {
	steam, game, err := steam.FindSteam()
	if err != nil {
		log.Fatalf("Error finding steam: %v", err)
	}
	log.Debugf("Steam path: %s, Lethal Company path: %s\n", steam, game)

	// Download bepinex from url
	// api.Download("https://thunderstore.io/c/lethal-company/p/BepInEx/BepInExPack/", "bepinex.zip")

}
