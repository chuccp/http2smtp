package service_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/service"
)

// restartWait is how long the service delays a restart so the response can go out first.
const restartWait = 700 * time.Millisecond

type probeRest struct{ tag string }

func (p *probeRest) Init(context *core.Context) error {
	context.Get("/probe", func(*web.Request) (any, error) { return p.tag, nil })
	return nil
}

// startProbeApp mirrors main.go: two rest groups, each holding the *web.ServerConfig the
// restart has to rewrite.
func startProbeApp(t *testing.T, cfg config.IConfig, managePort, apiPort int) (*service.RestartService, *web.ServerConfig, *web.ServerConfig) {
	t.Helper()

	manageConfig := &web.ServerConfig{Port: managePort}
	apiConfig := &web.ServerConfig{Port: apiPort}

	builder := wf.NewBuilder(cfg)
	manageGroup := core.NewRestGroupBuilder().Port(managePort).ContextPath("/api").ServerConfig(manageConfig)
	manageGroup.Rest(&probeRest{tag: "manage"})
	builder.RestGroup(manageGroup.Build())
	apiGroup := core.NewRestGroupBuilder().Port(apiPort).ServerConfig(apiConfig)
	apiGroup.Rest(&probeRest{tag: "api"})
	builder.RestGroup(apiGroup.Build())

	restartService := &service.RestartService{}
	builder.Service(restartService)
	app := builder.Build()
	restartService.Setup(manageConfig, apiConfig, app.ReStart)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = app.Run(ctx) }()

	return restartService, manageConfig, apiConfig
}

// probe returns the body served at url, or ok=false when nothing answers there.
func probe(url string) (string, bool) {
	resp, err := http.Get(url)
	if err != nil {
		return "", false
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}
	return string(body), true
}

func waitBody(t *testing.T, url, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		if body, ok := probe(url); ok && body == want {
			return
		} else if ok {
			last = body
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("waiting for %q from %s, last body %q", want, url, last)
}

func closed(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, ok := probe(url); !ok {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s still answers after the restart", url)
}

// A restart reuses the server groups built at startup, so only rewriting their ports makes
// a saved port change take effect. This covers that whole path.
func TestRestartService_AppliesNewPorts(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Put("manage.port", 19121)
	cfg.Put("api.port", 19122)
	cfg.Put("core.maxbodysize", "10M")

	restartService, manageConfig, apiConfig := startProbeApp(t, cfg, 19121, 19122)

	waitBody(t, "http://127.0.0.1:19121/api/probe", "manage", 5*time.Second)
	waitBody(t, "http://127.0.0.1:19122/probe", "api", 5*time.Second)

	// What the settings page does: persist the new ports and body cap, then ask for the
	// restart.
	cfg.Put("manage.port", 19123)
	cfg.Put("api.port", 19124)
	cfg.Put("core.maxbodysize", "20M")
	managePort, apiPort, err := restartService.Restart()
	if err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if managePort != 19123 || apiPort != 19124 {
		t.Fatalf("Restart reported ports %d/%d, want 19123/19124", managePort, apiPort)
	}
	if manageConfig.MaxBodySize != 20<<20 || apiConfig.MaxBodySize != 20<<20 {
		t.Errorf("body cap = %d/%d, want 20 MB on both servers", manageConfig.MaxBodySize, apiConfig.MaxBodySize)
	}

	waitBody(t, "http://127.0.0.1:19123/api/probe", "manage", 5*time.Second)
	waitBody(t, "http://127.0.0.1:19124/probe", "api", 5*time.Second)
	closed(t, "http://127.0.0.1:19121/api/probe", 5*time.Second)
	closed(t, "http://127.0.0.1:19122/probe", 5*time.Second)
}

// In Docker the ports come from the run command and the container's port mapping is built
// around them, so a port saved in the UI must not move the listeners: the restart re-inits
// in place and keeps serving where it was.
func TestRestartService_DockerKeepsCommandLinePorts(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Put("core.isdocker", "true")
	cfg.Put("manage.port", 19125) // the file disagrees with the run command
	cfg.Put("api.port", 19126)

	restartService, _, _ := startProbeApp(t, cfg, 19127, 19128)
	waitBody(t, "http://127.0.0.1:19127/api/probe", "manage", 5*time.Second)

	cfg.Put("manage.port", 19129)
	managePort, apiPort, err := restartService.Restart()
	if err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if managePort != 19127 || apiPort != 19128 {
		t.Fatalf("Restart reported ports %d/%d, want the command-line 19127/19128", managePort, apiPort)
	}

	// Serving again after the restart, still on the ports it was started with.
	time.Sleep(restartWait)
	waitBody(t, "http://127.0.0.1:19127/api/probe", "manage", 5*time.Second)
	waitBody(t, "http://127.0.0.1:19128/probe", "api", 5*time.Second)
	if _, ok := probe("http://127.0.0.1:19129/api/probe"); ok {
		t.Error("the restart moved the management server to the saved port")
	}
}

// A port the app cannot bind would kill it on the next startup, so the restart refuses the
// change instead and leaves the running instance alone.
func TestRestartService_RejectsUnusablePorts(t *testing.T) {
	busy, err := net.Listen("tcp", ":19131")
	if err != nil {
		t.Fatalf("holding a port for the test: %v", err)
	}
	defer func() { _ = busy.Close() }()

	cfg := config.NewConfig()
	cfg.Put("manage.port", 19125)
	cfg.Put("api.port", 19126)
	restartService, _, _ := startProbeApp(t, cfg, 19125, 19126)
	waitBody(t, "http://127.0.0.1:19125/api/probe", "manage", 5*time.Second)

	cfg.Put("manage.port", 19131)
	if _, _, err := restartService.Restart(); err == nil {
		t.Fatal("Restart accepted a port that is already taken")
	}

	cfg.Put("manage.port", 19126) // the API port: the two groups would share one engine
	if _, _, err := restartService.Restart(); err == nil {
		t.Fatal("Restart accepted the same port for both servers")
	}

	// Still serving on the original port, with nothing restarted.
	waitBody(t, "http://127.0.0.1:19125/api/probe", "manage", 2*time.Second)
}

// An unwired service must report an error rather than panic on a nil restart hook.
func TestRestartService_UnwiredIsAnError(t *testing.T) {
	unwired := &service.RestartService{}
	if err := unwired.Init(nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, _, err := unwired.Restart(); err == nil {
		t.Fatal("Restart on an unwired service returned no error")
	}
}
