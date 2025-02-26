package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/williamsjokvist/cfn-tracker/pkg/config"
	"github.com/williamsjokvist/cfn-tracker/pkg/i18n"
	"github.com/williamsjokvist/cfn-tracker/pkg/model"
	"github.com/williamsjokvist/cfn-tracker/pkg/server"
	cfgDb "github.com/williamsjokvist/cfn-tracker/pkg/storage/config"
	"github.com/williamsjokvist/cfn-tracker/pkg/storage/sql"
	"github.com/williamsjokvist/cfn-tracker/pkg/storage/txt"
)

// The CommandHandler is the interface between the GUI and the core
type CommandHandler struct {
	hasUpdate bool

	sqlDb *sql.Storage
	cfgDb *cfgDb.Storage

	cfg *config.BuildConfig
}

func NewCommandHandler(hasUpdate bool, sqlDb *sql.Storage, cfgDb *cfgDb.Storage, txtDb *txt.Storage, cfg *config.BuildConfig) *CommandHandler {
	return &CommandHandler{
		sqlDb:        sqlDb,
		cfgDb:        cfgDb,
		hasUpdate: hasUpdate,
		cfg:          cfg,
	}
}

func (ch *CommandHandler) CheckForUpdate() (bool, error) {
	return ch.hasUpdate, nil
}

func (ch *CommandHandler) GetTranslation(locale string) (*model.Localization, error) {
	lng, err := i18n.GetTranslation(locale)
	if err != nil {
		return nil, model.WrapError(model.ErrGetTranslations, err)
	}
	return lng, nil
}

func (ch *CommandHandler) GetAppVersion() string {
	return ch.cfg.AppVersion
}

func (ch *CommandHandler) OpenResultsDirectory() error {
	switch runtime.GOOS {
	case "darwin":
		if err := exec.Command("Open", "./results").Run(); err != nil {
			return model.WrapError(model.ErrOpenResultsDirectory, err)
		}
	case "windows":
		if err := exec.Command("explorer.exe", ".\results").Run(); err != nil {
			return model.WrapError(model.ErrOpenResultsDirectory, err)
		}
	}
	return nil
}

func (ch *CommandHandler) GetSessions(userId, date string, limit uint8, offset uint16) ([]*model.Session, error) {
	sessions, err := ch.sqlDb.GetSessions(context.Background(), userId, date, limit, offset)
	if err != nil {
		return nil, model.WrapError(model.ErrGetSessions, err)
	}
	return sessions, nil
}

func (ch *CommandHandler) GetSessionsStatistics(userId string) (*model.SessionsStatistics, error) {
	sessionStatistics, err := ch.sqlDb.GetSessionsStatistics(context.Background(), userId)
	if err != nil {
		return nil, model.WrapError(model.ErrGetSessionStatistics, err)
	}
	return sessionStatistics, nil
}

func (ch *CommandHandler) GetMatches(sessionId uint16, userId string, limit uint8, offset uint16) ([]*model.Match, error) {
	matches, err := ch.sqlDb.GetMatches(context.Background(), sessionId, userId, limit, offset)
	if err != nil {
		return nil, model.WrapError(model.ErrGetMatches, err)
	}
	return matches, nil
}

func (ch *CommandHandler) GetUsers() ([]*model.User, error) {
	users, err := ch.sqlDb.GetUsers(context.Background())
	if err != nil {
		return nil, model.WrapError(model.ErrGetUser, err)
	}
	return users, nil
}

func (ch *CommandHandler) GetThemes() ([]model.Theme, error) {
	// get internal themes
	internalThemes := server.GetInternalThemes()

	// get custom themes
	files, err := os.ReadDir("themes")
	if err != nil {
		return internalThemes, nil
	}
	customThemes := make([]model.Theme, 0, len(files))
	for _, file := range files {
		fileName := file.Name()

		if !strings.Contains(fileName, ".css") {
			continue
		}
		css, err := os.ReadFile(fmt.Sprintf("themes/%s", fileName))
		if err != nil {
			return nil, model.WrapError(model.ErrReadThemeCSS, err)
		}
		name := strings.Split(fileName, ".css")[0]

		customThemes = append(customThemes, model.Theme{
			Name: name,
			CSS:  string(css),
		})
	}

	combinedThemes := append(customThemes, internalThemes...)
	return combinedThemes, nil
}

func (ch *CommandHandler) GetSupportedLanguages() ([]string, error) {
	return i18n.GetSupportedLanguages()
}

func (ch *CommandHandler) SaveLocale(locale string) error {
	runtimeCfg, err := ch.cfgDb.GetRuntimeConfig()
	if err != nil {
		return model.WrapError(model.ErrGetGUIConfig, err)
	}
	runtimeCfg.GUI.Locale = locale
	if err := ch.cfgDb.SaveRuntimeConfig(runtimeCfg); err != nil {
		return model.WrapError(model.ErrSaveLocale, err)
	}
	return nil
}

func (ch *CommandHandler) GetGuiConfig() (*model.GUIConfig, error) {
	runtimeCfg, err := ch.cfgDb.GetRuntimeConfig()
	if err != nil {
		return nil, model.WrapError(model.ErrGetGUIConfig, err)
	}
	return &runtimeCfg.GUI, nil
}

func (ch *CommandHandler) SaveSidebarMinimized(sidebar bool) error {
	runtimeCfg, err := ch.cfgDb.GetRuntimeConfig()
	if err != nil {
		return model.WrapError(model.ErrGetGUIConfig, err)
	}
	runtimeCfg.GUI.SideBar = sidebar
	if err := ch.cfgDb.SaveRuntimeConfig(runtimeCfg); err != nil {
		return model.WrapError(model.ErrSaveSidebar, err)
	}
	return nil
}

func (ch *CommandHandler) SaveTheme(theme model.ThemeName) error {
	runtimeCfg, err := ch.cfgDb.GetRuntimeConfig()
	if err != nil {
		return model.WrapError(model.ErrGetGUIConfig, err)
	}
	runtimeCfg.GUI.Theme = theme
	if err := ch.cfgDb.SaveRuntimeConfig(runtimeCfg); err != nil {
		return model.WrapError(model.ErrSaveTheme, err)
	}
	return nil
}

func (ch *CommandHandler) GetFGCTrackerErrorModelUnused() *model.FGCTrackerError {
	return nil
}
