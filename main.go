package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/williamsjokvist/cfn-tracker/cmd"
	"github.com/williamsjokvist/cfn-tracker/pkg/browser"
	"github.com/williamsjokvist/cfn-tracker/pkg/config"
	"github.com/williamsjokvist/cfn-tracker/pkg/model"
	"github.com/williamsjokvist/cfn-tracker/pkg/server"
	cfgDb "github.com/williamsjokvist/cfn-tracker/pkg/storage/config"
	"github.com/williamsjokvist/cfn-tracker/pkg/storage/sql"
	"github.com/williamsjokvist/cfn-tracker/pkg/storage/txt"
	"github.com/williamsjokvist/cfn-tracker/pkg/tracker/sf6/cfn"
	"github.com/williamsjokvist/cfn-tracker/pkg/tracker/t8/wavu"
	"github.com/williamsjokvist/cfn-tracker/pkg/update/github"
)

var (
	capIDEmail    string = ``
	capIDPassword string = ``
	appVersion    string = ``
	isProduction  string = ``
)

//go:embed all:gui/dist
var assets embed.FS

//go:embed gui/error/error.html
var errorTmpl []byte

var cfg config.BuildConfig
var logFile *os.File

func logToFile() {
	file, err := os.OpenFile(`cfn-tracker.log`, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	logFile = file
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.LstdFlags | log.Lshortfile)
}

func init() {
	if isProduction == `true` {
		logToFile()
	}

	err := godotenv.Load(`.env`)
	if err != nil {
		log.Println(fmt.Errorf(`missing .env file: %w`, err))
		cfg = config.BuildConfig{
			AppVersion:        appVersion,
			Headless:          isProduction == `true`,
			CapIDEmail:        capIDEmail,
			CapIDPassword:     capIDPassword,
			BrowserSourcePort: 4242,
		}
		return
	}
	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("process envconfig: %v", err)
	}
}

func main() {
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	// If we have a previous instance running, wait for it to close before proceeding
	if i := lo.IndexOf(os.Args, "--auto-update"); i != -1 {
		if len(os.Args) < i+2 {
			panic(`missing pid argument for --auto-update`)
		}
		prevInstancePidStr := os.Args[i+1]
		prevInstancePid, err := strconv.Atoi(prevInstancePidStr)
		if err != nil {
			panic(fmt.Sprintf(`failed to convert pid to int: %v`, err))
		}

		for i := 0; i < 10; i++ {

			// On Unix systems, FindProcess always succeeds and returns a Process
			// for the given pid, regardless of whether the process exists. To test whether
			// the process actually exists, see whether p.Signal(syscall.Signal(0)) reports
			// an error.
			p, err := os.FindProcess(prevInstancePid)
			if err != nil {
				slog.Warn(fmt.Sprintf(`failed (err received) to find previous instance process, it's probably shut down...: %v'`, err))
				break
			}
			if p == nil {
				slog.Info(`failed to find previous instance process, it's probably shut down...'`)
				break
			}

			err = p.Signal(syscall.Signal(0))
			if err != nil {
				// The process is not running
				break
			}

			slog.Info(`waiting for previous instance to close...`)
			time.Sleep(1 * time.Second)

		}
	}

	var appBrowser *browser.Browser
	closeWithError := func(err error) {
		if appBrowser != nil {
			appBrowser.Page.Browser().Close()
		}
		log.Println("close with error", err)
		if err := wails.Run(&options.App{
			Title:                    `CFN Tracker - Error`,
			Width:                    400,
			Height:                   148,
			DisableResize:            true,
			Frameless:                true,
			EnableDefaultContextMenu: false,
			AssetServer: &assetserver.Options{
				Middleware: assetserver.ChainMiddleware(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						var b bytes.Buffer
						tmpl := template.Must(template.New("errorPage").Parse(string(errorTmpl)))
						params := struct {
							Error string
						}{
							Error: err.Error(),
						}
						if err := tmpl.Execute(&b, params); err != nil {
							log.Println("failed to write error template: ", err)
						}
						_, err := w.Write(b.Bytes())
						if err != nil {
							log.Println("failed to write error page: ", err)
						}
					})
				}),
			},
		}); err != nil {
			log.Println("failed to launch error app")
		}
	}
	appBrowser, err := browser.NewBrowser(cfg.Headless)
	if err != nil {
		closeWithError(fmt.Errorf("failed to launch browser: %w", err))
		return
	}
	sqlDb, err := sql.NewStorage(false)
	if err != nil {
		closeWithError(fmt.Errorf("failed to initalize database: %w", err))
		return
	}
	noSqlDb, err := cfgDb.NewStorage()
	if err != nil {
		closeWithError(fmt.Errorf("failed to initalize app config: %w", err))
		return
	}
	txtDb, err := txt.NewStorage()
	if err != nil {
		closeWithError(fmt.Errorf("failed to initalize text store: %w", err))
		return
	}

	browserSrcMatchChan := make(chan model.Match, 1)

	cmdHandler := cmd.NewCommandHandler(
		github.NewClient(),
		sqlDb,
		noSqlDb,
		txtDb,
		&cfg,
	)
	trackingHandler := cmd.NewTrackingHandler(
		wavu.NewClient(),
		cfn.NewClient(appBrowser),
		sqlDb,
		noSqlDb,
		txtDb,
		&cfg,
		browserSrcMatchChan,
	)

	browserSrcServer := server.NewBrowserSourceServer(browserSrcMatchChan)

	var onSecondInstanceLaunch func(secondInstanceData options.SecondInstanceData)
	err = wails.Run(&options.App{
		Title:              fmt.Sprintf(`CFN Tracker v%s`, appVersion),
		Assets:             assets,
		Width:              920,
		Height:             450,
		MinWidth:           800,
		MinHeight:          450,
		DisableResize:      true,
		Fullscreen:         false,
		Frameless:          true,
		StartHidden:        false,
		HideWindowOnClose:  false,
		LogLevel:           logger.WARNING,
		LogLevelProduction: logger.ERROR,
		ErrorFormatter:     model.FormatError,
		BackgroundColour:   options.NewRGBA(0, 0, 0, 1),
		CSSDragProperty:    `--draggable`,
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			Theme:                             windows.Theme(windows.Dark),
			DisableFramelessWindowDecorations: true,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
				Title:   fmt.Sprintf(`CFN Tracker v%s`, appVersion),
				Message: fmt.Sprintf(`CFN Tracker version %s © 2023 William Sjökvist <william.sjokvist@gmail.com>`, appVersion),
			},
		},
		OnDomReady: func(ctx context.Context) {
			onSecondInstanceLaunch = func(secondInstanceData options.SecondInstanceData) {
				log.Println("user opened second instance", strings.Join(secondInstanceData.Args, ","))
				log.Println("user opened second from", secondInstanceData.WorkingDirectory)
				runtime.WindowUnminimise(ctx)
				runtime.Show(ctx)
				go runtime.EventsEmit(ctx, "launchArgs", secondInstanceData.Args)
			}

			trackingHandler.SetEventEmitter(func(eventName string, optionalData ...interface{}) {
				log.Println("[FE EVENT]", eventName, optionalData)
				runtime.EventsEmit(ctx, eventName, optionalData...)
			})
		},
		OnStartup: func(ctx context.Context) {
			go browserSrcServer.Start(ctx, &cfg)
		},
		OnShutdown: func(_ context.Context) {
			appBrowser.Page.Browser().Close()
		},
		OnBeforeClose: func(_ context.Context) (prevent bool) {
			appBrowser.Page.Browser().Close()
			return false
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "d0ef6612-49f7-437a-9ffc-2076ec9e37db",
			OnSecondInstanceLaunch: onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			cmdHandler,
			trackingHandler,
		},
		EnumBind: []interface{}{
			model.AllThemes,
			model.AllGameTypes,
			model.AllErrorKeys,
		},
	})
	if err != nil {
		closeWithError(fmt.Errorf("failed to launch: %w", err))
	}
}
