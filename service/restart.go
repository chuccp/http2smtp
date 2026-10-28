package service

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/model"
)

// restartDelay is how long the restart waits, so the HTTP response that asked for it
// reaches the client before the listeners close.
const restartDelay = 500 * time.Millisecond

// RestartService restarts the app in place through WebFrame.ReStart.
//
// A restart re-runs the server groups main.go built, and those carry the listen ports
// captured at startup, so a port saved in the settings page would otherwise be ignored.
// The service keeps the *web.ServerConfig of both listeners and rewrites their ports
// before restarting, which is what makes a new port take effect.
//
// Docker is the exception: there the ports come from the command line or the environment
// and the container's port mapping is built around them, so `core.isdocker` makes the
// restart leave them be.
type RestartService struct {
	context *core.Context

	manage  *web.ServerConfig
	api     *web.ServerConfig
	restart func()
}

func (s *RestartService) Init(context *core.Context) error {
	s.context = context
	return nil
}

// Setup wires the listeners and the restart hook. main.go calls it after Build and before
// Start, so nothing reads these fields concurrently.
func (s *RestartService) Setup(manage, api *web.ServerConfig, restart func()) {
	s.manage = manage
	s.api = api
	s.restart = restart
}

// Restart applies the ports currently held in the configuration and restarts the app a
// moment later, returning the ports that will be in effect. It does not persist anything:
// callers save through /reSet first.
func (s *RestartService) Restart() (int, int, error) {
	if s.restart == nil || s.manage == nil || s.api == nil {
		return 0, 0, errors.New("restart is not wired up")
	}
	cfg, err := model.GetConfig(s.context.GetConfig())
	if err != nil {
		return 0, 0, err
	}
	if cfg.Manage == nil || cfg.Api == nil {
		return 0, 0, errors.New("manage/api config is missing")
	}
	if cfg.Core != nil && cfg.Core.IsDocker {
		// In Docker the ports arrive on the command line or the environment and the
		// container's port mapping is built around them, so a restart re-inits without
		// moving them — whatever the settings page wrote to config.ini. The body cap is
		// not a port, so it still follows the restart.
		s.applyMaxBodySize(cfg)
		time.AfterFunc(restartDelay, s.restart)
		return s.manage.Port, s.api.Port, nil
	}
	managePort, apiPort := cfg.Manage.Port, cfg.Api.Port
	if err := checkPorts(managePort, apiPort, s.manage.Port, s.api.Port); err != nil {
		return 0, 0, err
	}
	s.applyMaxBodySize(cfg)
	s.manage.Port = managePort
	s.api.Port = apiPort
	time.AfterFunc(restartDelay, s.restart)
	return managePort, apiPort, nil
}

// applyMaxBodySize carries the configured cap onto both servers. It runs only once the
// restart is actually going ahead, so a refused restart leaves the running limit alone.
func (s *RestartService) applyMaxBodySize(cfg *model.Config) {
	size := model.DefaultMaxBodySize
	if cfg.Core != nil {
		size = cfg.Core.MaxBodySizeValue()
	}
	s.manage.MaxBodySize = size
	s.api.MaxBodySize = size
}

// checkPorts rejects a change the app could not come back up with, leaving the running
// instance untouched. Ports our own listeners already hold are fine — they are released as
// the app stops — but the two must differ: groups sharing a port are merged onto one
// gin engine, and the public API would inherit the management server's auth filter.
func checkPorts(managePort, apiPort int, current ...int) error {
	for _, port := range []int{managePort, apiPort} {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port %d is out of range", port)
		}
	}
	if managePort == apiPort {
		return errors.New("the management port and the API port must differ")
	}
	for _, port := range []int{managePort, apiPort} {
		if slices.Contains(current, port) {
			continue
		}
		listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return fmt.Errorf("port %d is unavailable: %v", port, err)
		}
		_ = listener.Close()
	}
	return nil
}
