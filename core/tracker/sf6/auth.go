package sf6

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

type AuthStatus struct {
	Progress int
	Err      error
}

func (s *AuthStatus) WithProgress(progress int) *AuthStatus {
	s.Progress = progress
	return s
}

func (s *AuthStatus) WithError(err error) *AuthStatus {
	s.Err = err
	return s
}

func (t *SF6Tracker) Authenticate(ctx context.Context, email string, password string, statChan chan AuthStatus) {
	status := &AuthStatus{Progress: 0, Err: nil}
	defer func() {
		if r := recover(); r != nil {
			log.Println(`Recovered from panic: `, r)
			statChan <- *status.WithError(fmt.Errorf(`panic: %v`, r))
		}
	}()

	if t.isAuthenticated || strings.Contains(t.Page.MustInfo().URL, `buckler`) {
		t.isAuthenticated = true
		return
	}

	log.Println(`Logging in`)
	t.Page.MustNavigate(`https://cid.capcom.com/ja/login/?guidedBy=web`).MustWaitLoad().MustWaitIdle()
	statChan <- *status.WithProgress(10)

	log.Print("Checking if already authed")
	if strings.Contains(t.Page.MustInfo().URL, `cid.capcom.com/ja/mypage`) {
		log.Print("User already authed")
		t.isAuthenticated = true
		statChan <- *status.WithProgress(100)
		return
	}
	log.Print("Not authed, continuing with auth process")

	// Bypass age check
	if strings.Contains(t.Page.MustInfo().URL, `agecheck`) {
		t.Page.MustElement(`#country`).MustSelect(COUNTRIES[rand.Intn(len(COUNTRIES))])
		t.Page.MustElement(`#birthYear`).MustSelect(strconv.Itoa(rand.Intn(1999-1970) + 1970))
		t.Page.MustElement(`#birthMonth`).MustSelect(strconv.Itoa(rand.Intn(12-1) + 1))
		t.Page.MustElement(`#birthDay`).MustSelect(strconv.Itoa(rand.Intn(28-1) + 1))
		t.Page.MustElement(`form button[type="submit"]`).MustClick()
		t.Page.MustWaitLoad().MustWaitRequestIdle()
	}
	statChan <- *status.WithProgress(30)

	// Submit form
	t.Page.MustElement(`input[name="email"]`).Input(email)
	t.Page.MustElement(`input[name="password"]`).Input(password)
	t.Page.MustElement(`button[type="submit"]`).MustClick()
	statChan <- *status.WithProgress(50)

	// Wait for redirection
	var secondsWaited time.Duration = 0
	for {
		// Break out if we are no longer on Auth0 (redirected to CFN)
		if !strings.Contains(t.Page.MustInfo().URL, `auth.cid.capcom.com`) {
			break
		}

		time.Sleep(time.Second)
		secondsWaited += time.Second
		log.Println(`Waiting for gateway to pass...`, secondsWaited)
	}
	statChan <- *status.WithProgress(65)

	t.Page.MustNavigate(`https://www.streetfighter.com/6/buckler/auth/loginep?redirect_url=/`)
	t.Page.MustWaitLoad().MustWaitRequestIdle()

	statChan <- *status.WithProgress(100)
	t.isAuthenticated = true
}