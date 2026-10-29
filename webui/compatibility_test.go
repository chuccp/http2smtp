package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/http2smtp/compatibility"
)

func TestCompatibilityPageShowsBackupAndResetInstructions(t *testing.T) {
	page := NewCompatibilityPage(&compatibility.IncompatibleError{
		Code:            compatibility.ProblemLegacyData,
		ConfigPath:      `C:\app\config.ini`,
		SQLitePath:      `C:\app\d-mail.db`,
		StorageRoot:     `C:\app\storage`,
		DatabaseType:    "sqlite",
		RequiredVersion: compatibility.CurrentDataVersion,
	})
	server := templateTestServer(t, page.render)
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.Header.Set("Accept-Language", "en")
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Upgrade blocked",
		"This update includes major changes, so a smooth upgrade is not possible",
		"Back up before resetting",
		`C:\app\config.ini`,
		`C:\app\d-mail.db`,
		`C:\app\storage`,
		"Reset after the backup is complete",
		"Do not delete it",
		"deletes old configuration",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("compatibility page does not contain %q:\n%s", want, body)
		}
	}
}

func TestCompatibilityPageShowsMySQLBackupStep(t *testing.T) {
	page := NewCompatibilityPage(&compatibility.IncompatibleError{
		Code:            compatibility.ProblemVersionMismatch,
		FoundVersion:    2,
		RequiredVersion: compatibility.CurrentDataVersion,
		ConfigPath:      "/app/config.ini",
		StorageRoot:     "/app/storage",
		DatabaseType:    "mysql",
		MySQLDatabase:   "http2smtp",
	})
	server := templateTestServer(t, page.render)
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.Header.Set("Accept-Language", "en")
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, req)

	body := recorder.Body.String()
	for _, want := range []string{
		"The data version is 2, but this program requires version 1.",
		"MySQL database (take a full backup first)",
		"Create a new empty MySQL database",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("MySQL compatibility page does not contain %q:\n%s", want, body)
		}
	}
}

// The compatibility UI must work as a real management route with no model group or
// database attached. /settings is the upgrade path users are most likely to open.
func TestCompatibilityPageRegistersSettingsWithoutDatabase(t *testing.T) {
	previousDevMode := devMode
	devMode = false
	t.Cleanup(func() { devMode = previousDevMode })

	page := NewCompatibilityPage(&compatibility.IncompatibleError{
		Code:            compatibility.ProblemLegacyData,
		ConfigPath:      "/app/config.ini",
		SQLitePath:      "/app/d-mail.db",
		StorageRoot:     "/app/storage",
		DatabaseType:    "sqlite",
		RequiredVersion: compatibility.CurrentDataVersion,
	})
	app := wf.NewBuilder(config.NewConfig()).Rest(page).Build()
	handler := app.GetHandler()
	if handler == nil {
		t.Fatal("compatibility application did not initialize")
	}

	for _, path := range []string{"/", "/settings"} {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:19009"+path, nil)
		req.Header.Set("Accept-Language", "en")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Errorf("%s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), "Upgrade blocked") {
			t.Errorf("%s did not render the compatibility page:\n%s", path, recorder.Body.String())
		}
	}
}
