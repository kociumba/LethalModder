package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/kociumba/LethalModder/api"
	"github.com/kociumba/LethalModder/steam"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	err             error
	packageListings []api.PackageListing

	IsLethalCompanyInstalled = false

	dbg   = flag.Bool("dbg", false, "enable debug logging")
	print = flag.Bool("print", false, "print to stdout")
)

func main() {
	debug.SetMemoryLimit(10 * 1024 * 1024 * 1024)
	debug.SetMaxThreads(100)
	debug.SetMaxStack(2 * 1024 * 1024 * 1024)
	debug.SetGCPercent(80)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	// Create an instance of the app structure
	app := NewApp()

	// Init the mod data
	InitData()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "LethalModder",
		Width:  1600,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		CSSDragProperty: "--wails-draggable",
		CSSDragValue:    "drag",
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.Mica,
		},
	})

	if err != nil {
		log.Error("Error:", err.Error())
	}
}

func InitData() {
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

	// err = spinner.New().Title("Loading mods...").Run()
	// if err != nil {
	// 	log.Fatalf("Error starting spinner: %v", err)
	// }

	wg.Wait()

	// huh.NewSelect[int]().Run()

	// Uncomment to test if initialization works correctly
	// if *dbg {
	// 	os.Exit(0)
	// }

}

// This is fucking 182mb
// wtf were they smoking
//
// This is gonna have to init like smt like:
//
//	cfg.Mods{}
func initMods() {
	packageListings, err = api.GetMods()
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

	// Change this to a config struct or someting
	//
	// items := make([]list.Item, 0)
	// for _, listing := range packageListings {
	// 	isDeprecated, _ := listing.GetIsDeprecated()
	// 	if !isDeprecated {
	// 		rating, err := listing.GetRatingScore()
	// 		if err != nil {
	// 			log.Errorf("Error getting rating score: %v", err)
	// 		}
	// 		items = append(items, item{title: listing.Name, desc: "Owner: " + listing.Owner + " - " + fmt.Sprint(rating) + "☆"})
	// 	}
	// }
}

func initLocalProfiles() {
	steam, game, err := steam.FindSteam()
	if err != nil {
		// was crashing in CI/CD couse no steam installation was found
		log.Errorf("Error finding steam: %v", err)
	}
	log.Debugf("Steam path: %s, Lethal Company path: %s\n", steam, game)

	// gonna need this later
	_, err = os.Stat(game)
	if err != nil {
		IsLethalCompanyInstalled = true
	}

	// Download bepinex from url
	// api.Download("https://thunderstore.io/c/lethal-company/p/BepInEx/BepInExPack/", "bepinex.zip")

}
