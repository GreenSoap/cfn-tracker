package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type MatchHistory struct {
	lp           int
	lpGain       int
	wins         int
	totalWins    int
	totalLosses  int
	totalMatches int
	losses       int
	winrate      int
}

var matchHistory = MatchHistory{
	lp:           0,
	lpGain:       0,
	wins:         0,
	losses:       0,
	totalWins:    0,
	totalLosses:  0,
	totalMatches: 0,
	winrate:      0,
}

var (
	isTracking    = false
	isInitialized = false
	pageInstance  *rod.Page
)

func LogMatchHistory() {
	fmt.Println("["+time.Now().Format(`15:04`)+"] LP:", matchHistory.lp, "/ Gain:", matchHistory.lpGain, "/ Wins:", matchHistory.wins, "/ Losses:", matchHistory.losses, "/ Winrate:", matchHistory.winrate, `%`)
}

func Login(profile string, page *rod.Page, steamUsername string, steamPassword string) (int, *rod.Page) {
	fmt.Println("Logging in")
	page.MustNavigate(`https://game.capcom.com/cfn/sfv/consent/steam`).MustWaitLoad()

	// Accepting CFN terms
	wait := page.MustWaitLoad().MustWaitRequestIdle()
	page.MustElement(`input[type="submit"]`).MustClick()
	wait()
	fmt.Println("Accepted CFN terms")

	// If CFN already opened
	url := page.MustInfo().URL
	if url != `https://game.capcom.com/cfn/sfv/` {
		page.WaitElementsMoreThan(`#loginForm`, 0)
	}

	isSteamOpen, _, _ := page.Has(`#loginForm`)

	if isSteamOpen {
		fmt.Println("Passing the gateway")
		if page.MustInfo().URL == `https://game.capcom.com/cfn/sfv/` {
			return LoginError.returnCode, nil
		}

		usernameElement, _ := page.Element(`#loginForm input[name="username"]`)
		passwordElement, _ := page.Element(`#loginForm input[name="password"]`)
		buttonElement, e := page.Element(`input#imageLogin`)

		if e != nil {
			return LoginError.returnCode, nil
		}

		usernameElement.Input(steamUsername)
		passwordElement.Input(steamPassword)
		buttonElement.Click(proto.InputMouseButtonLeft, 2)

		var secondsWaited time.Duration = 0
		for {
			body := page.MustElement(`body`)
			errorElement, _ := body.Element(`#error_display`)
			if errorElement != nil {
				errorText, e := errorElement.Text()

				if e != nil || len(errorText) > 0 {
					return CaptchaError.returnCode, nil
				}
			}

			time.Sleep(time.Second)
			secondsWaited += time.Second
			if !strings.Contains(page.MustInfo().URL, `steam`) {
				// Gateway passed
				break
			}
		}
	}

	fmt.Println("Gateway passed")

	return 1, page
}

func RefreshData(profile string, page *rod.Page) {
	isFirstFetch := matchHistory.lp == 0
	if !isFirstFetch && page.MustInfo().URL != `https://game.capcom.com/cfn/sfv/profile/`+profile {
		return
	}

	if !isFirstFetch {
		fmt.Println("Reloading page")
		page.Reload()
	}

	// Read from DOM
	totalMatchesEl, e := page.Element(`.battleNumber>.total>dd`)
	totalWinsEl, e := page.Element(`.battleNumber>.win>dd`)
	totalLossesEl, e := page.Element(`.battleNumber>.lose>dd`)
	lpEl, e := page.Element(`.leagueInfo>dl:last-child>dd`)

	if e != nil {
		LogError(ParseError)
		return
	}

	// Convert to ints
	newLp, _ := strconv.Atoi(strings.TrimSuffix(lpEl.MustText(), `LP`))
	totalWins, _ := strconv.Atoi(totalWinsEl.MustText())
	totalLosses, _ := strconv.Atoi(totalLossesEl.MustText())
	totalMatches, e := strconv.Atoi(totalMatchesEl.MustText())

	if e != nil {
		LogError(ParseError)
		return
	}

	hasNewMatch := totalMatches != matchHistory.totalMatches

	// Return if no new data
	if !(isFirstFetch || hasNewMatch) {
		return
	}

	// Matches have been played since first fetch
	if hasNewMatch && !isFirstFetch {
		matchHistory.wins = matchHistory.wins + int(math.Abs(float64(matchHistory.totalWins)-float64(totalWins)))
		matchHistory.losses = matchHistory.losses + int(math.Abs(float64(matchHistory.totalLosses)-float64(totalLosses)))
		matchHistory.lpGain = matchHistory.lpGain + (newLp - matchHistory.lp)
		matchHistory.winrate = int((float64(matchHistory.wins) / float64(matchHistory.wins+matchHistory.losses)) * 100)
	}

	matchHistory.totalWins = totalWins
	matchHistory.totalLosses = totalLosses
	matchHistory.totalMatches = totalMatches
	matchHistory.lp = newLp

	SaveMatchHistory(matchHistory)
	LogMatchHistory()
}

func SetupBrowser() *rod.Page {
	fmt.Println("Setting up browser")
	u := launcher.New().Leakless(false).Headless(true).MustLaunch()
	page := rod.New().ControlURL(u).MustConnect().MustPage("")
	router := page.HijackRequests()

	// Block all images, stylesheets, fonts and unessential scripts
	router.MustAdd("*", func(ctx *rod.Hijack) {
		if ctx.Request.Type() == proto.NetworkResourceTypeImage ||
			ctx.Request.Type() == proto.NetworkResourceTypeStylesheet ||
			ctx.Request.Type() == proto.NetworkResourceTypeFont {
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		// Only check for scripts on non-steam requests
		if !strings.Contains(ctx.Request.URL().Hostname(), `steam`) &&
			ctx.Request.Type() == proto.NetworkResourceTypeScript {
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()
	return page
}

func Initialize() int {
	if isInitialized {
		return 1
	}

	page := SetupBrowser()
	pageInstance = page
	loginStatus, page := Login(profile, page, steamUsername, steamPassword)
	isInitialized = (loginStatus == 1)
	if loginStatus == LoginError.returnCode {
		LogError(LoginError)
	} else if loginStatus == ProfileError.returnCode {
		LogError(ProfileError)
	} else if loginStatus == CaptchaError.returnCode {
		LogError(CaptchaError)
	}
	pageInstance = page
	return loginStatus
}

func StartTracking(profile string) {
	if isInitialized == false {
		return
	}

	fmt.Println("Loading profile")
	pageInstance.MustNavigate(`https://game.capcom.com/cfn/sfv/profile/` + profile).MustWaitLoad()
	fmt.Println("Profile loaded")
	isTracking = true
	time.Sleep(3 * time.Second)
	for {
		if isTracking == false {
			fmt.Println("Stopped tracking")
			break
		}
		RefreshData(profile, pageInstance)
		time.Sleep(30 * time.Second)
	}
}
