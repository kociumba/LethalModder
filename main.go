package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/kociumba/LethalModder/api"
	"github.com/kociumba/LethalModder/profiles"
	"github.com/kociumba/LethalModder/steam"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed frontend/dist/*
var assets embed.FS

var (
	err             error
	packageListings []api.PackageListing

	IsLethalCompanyInstalled = false
	steamPath                string
	gamePath                 string

	SelectProfile profiles.Profile

	packageMap     map[string]api.PackageListing
	downloadedMods map[string]bool
	globalMu       sync.RWMutex

	done = make(chan bool)

	dbg   = flag.Bool("dbg", false, "enable debug logging")
	print = flag.Bool("print", false, "print to stdout")

	// wailsv3 app
	app = application.New(application.Options{
		Name:        "LethalModder",
		Description: "A Lethal Company only alternative to the incredibly laggy r2modman mod manager.",
		// Icon:        []byte{},
		Services: []application.Service{
			application.NewService(&DataService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		PanicHandler: func(r any) {
			log.Info("Panic occurred", "error", r)
		},
	})

	// splash screen
	splashScreen = app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Name:             "splashScreen",
		Title:            "LethalModder",
		Width:            600,
		Height:           300,
		HTML:             splashScreenHTML,
		AlwaysOnTop:      true,
		URL:              "/splash",
		DisableResize:    true,
		Frameless:        true,
		Centered:         true,
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.RGBA{Red: 0, Green: 0, Blue: 0, Alpha: 0},
		Windows: application.WindowsWindow{
			DisableIcon:                       true,
			DisableFramelessWindowDecorations: true,
			DisableMenu:                       true,
			HiddenOnTaskbar:                   true,
		},
		ShouldClose: func(window *application.WebviewWindow) bool {
			return true
		},
		IgnoreMouseEvents: true,
	})

	// main window
	LethalModder = app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Name:                    "mainApp",
		Title:                   "LethalModder",
		Width:                   1280,
		Height:                  720,
		URL:                     "/",
		Centered:                true,
		BackgroundType:          application.BackgroundTypeTranslucent,
		BackgroundColour:        application.RGBA{Red: 19, Green: 23, Blue: 31, Alpha: 0},
		FullscreenButtonEnabled: true,
		ZoomControlEnabled:      true,
		Frameless:               true,
		Windows: application.WindowsWindow{
			Theme:                             application.Dark,
			ResizeDebounceMS:                  100,
			BackdropType:                      application.Acrylic,
			DisableFramelessWindowDecorations: true,
		},
		ShouldClose: func(window *application.WebviewWindow) bool {
			return true
		},
		// DevToolsEnabled:            false,
		// DefaultContextMenuDisabled: false,
	})
)

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.

	// Tells the frontend when data is actaully loaded
	// Needed with more than one window
	go func() {
		for {
			time.Sleep(50 * time.Millisecond)
			if len(packageListings) != 0 {
				app.Events.Emit(&application.WailsEvent{
					Name: "totalItems",
					Data: len(packageListings),
				})
				break
			}
		}
	}()

	go ManageWindows()

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}

// Make this shit work couse for some reason it's fucked
func ManageWindows() {
	splashScreen.Show()
	LethalModder.Hide()

	go func() {
		InitData()
	}()

	if <-done {
		// something isn't loading
		// time.Sleep(1 * time.Second)
		// log.Info(packageListings[0:2])

		// Needed when centering the splash screen
		if *dbg {
			splashScreen.OpenDevTools()
			return
		}
		splashScreen.Close()
		LethalModder.Show()
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

	wg.Wait()

	InitializeGlobalMaps(packageListings)

	// Uncomment to only test initialization
	// if *dbg {
	// 	os.Exit(0)
	// }
}

// This is fucking 182mb
// wtf were they smoking
//
// Implementation is in ./app.go
// That stuff works of of arrays and data structs
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

	// packageMap = make(map[string]api.PackageListing, len(packageListings))
	// for _, listing := range packageListings {
	// 	packageMap[listing.Name] = listing
	// }

	done <- true
}

func initLocalProfiles() {
	steamPath, gamePath, err = steam.FindSteam()
	if err != nil {
		// was crashing in CI/CD couse no steam installation was found
		log.Errorf("Error finding steam: %v", err)
	}
	log.Infof("Steam path: %s, Lethal Company path: %s\n", steamPath, gamePath)

	// gonna need this later
	_, err = os.Stat(gamePath)
	if err == nil {
		IsLethalCompanyInstalled = true
	}

	log.Info("Steam check complete: ", "IsLethalCompanyInstalled", IsLethalCompanyInstalled)

	// idk this event doesn't get picked up
	app.Events.Emit(&application.WailsEvent{
		Name: "lethalCompanyWarning",
		Data: IsLethalCompanyInstalled,
	})

	profiles.GetLocalData()

	// Download bepinex from url
	// api.Download("https://thunderstore.io/c/lethal-company/p/BepInEx/BepInExPack/", "bepinex.zip")

}
