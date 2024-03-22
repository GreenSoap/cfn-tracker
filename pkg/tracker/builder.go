package tracker

import (
	"context"
	"net/http"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/williamsjokvist/cfn-tracker/pkg/browser"
	"github.com/williamsjokvist/cfn-tracker/pkg/config"
	"github.com/williamsjokvist/cfn-tracker/pkg/errorsx"
	"github.com/williamsjokvist/cfn-tracker/pkg/storage/sql"
	"github.com/williamsjokvist/cfn-tracker/pkg/storage/txt"
	"github.com/williamsjokvist/cfn-tracker/pkg/tracker/sf6"
	"github.com/williamsjokvist/cfn-tracker/pkg/tracker/sfv"
)

type GameTracker interface {
	Start(ctx context.Context, cfn string, restore bool, refreshInterval time.Duration) error
	Stop()
}

type GameType uint8

const (
	GameTypeUndefined GameType = iota
	GameTypeSFV
	GameTypeSF6
)

func (s GameType) String() string {
	switch s {
	case GameTypeSFV:
		return `sfv`
	case GameTypeSF6:
		return `sf6`
	case GameTypeUndefined:
		return `undefined`
	}
	return `unknown`
}

// Make a SF6Tracker and expose it as a GameTracker
func MakeSF6Tracker(ctx context.Context, cfg *config.Config, browser *browser.Browser, sqlDb *sql.Storage, txtDb *txt.Storage) (GameTracker, error) {
	sf6Tracker := sf6.NewSF6Tracker(browser, sqlDb, txtDb)

	authChan := make(chan sf6.AuthStatus)
	go sf6Tracker.Authenticate(ctx, cfg.CapIDEmail, cfg.CapIDPassword, authChan)
	for status := range authChan {
		if status.Err != nil {
			return nil, errorsx.NewFormattedError(http.StatusUnauthorized, status.Err)
		}
		runtime.EventsEmit(ctx, "auth-progress", status.Progress)

		if status.Progress >= 100 {
			close(authChan)
			break
		}
	}

	var gt GameTracker = sf6Tracker
	return gt, nil
}

// Make a SFVTracker and expose it as a GameTracker
func MakeSFVTracker(ctx context.Context, cfg *config.Config, browser *browser.Browser) (GameTracker, error) {
	sfvTracker := sfv.NewSFVTracker(browser)

	authChan := make(chan sfv.AuthStatus)
	go sfvTracker.Authenticate(ctx, cfg.SteamUsername, cfg.SteamPassword, authChan)
	for status := range authChan {
		if status.Err != nil {
			return nil, errorsx.NewFormattedError(http.StatusUnauthorized, status.Err)
		}
		runtime.EventsEmit(ctx, "auth-progress", status.Progress)

		if status.Progress >= 100 {
			close(authChan)
			break
		}
	}

	var gt GameTracker = sfvTracker
	return gt, nil
}