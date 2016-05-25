package gateway

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bcwaldon/farva/pkg/health"
	"github.com/bcwaldon/farva/pkg/logger"
)

type Config struct {
	RefreshInterval  time.Duration
	KubeconfigFile   string
	ClusterZone      string
	NGINXDryRun      bool
	NGINXHealthPort  int
	HTTPListenPort   int
	FarvaHealthPort  int
	AnnotationPrefix string
}

var DefaultConfig = Config{
	HTTPListenPort:  7331,
	FarvaHealthPort: 7333,
}

func DefaultHTTPReverseProxyServers(cfg *Config) []httpReverseProxyServer {
	return []httpReverseProxyServer{
		httpReverseProxyServer{
			ListenPort: cfg.NGINXHealthPort,
			Locations: []httpReverseProxyLocation{
				httpReverseProxyLocation{
					Path:          "/health",
					StaticCode:    200,
					StaticMessage: "Healthy!",
				},
			},
		},
		httpReverseProxyServer{
			ListenPort: cfg.HTTPListenPort,
			StaticCode: 444,
		},
	}
}

func DefaultReverseProxyConfig(cfg *Config) *reverseProxyConfig {
	return &reverseProxyConfig{
		HTTPServers: DefaultHTTPReverseProxyServers(cfg),
	}
}

func New(cfg Config) (*Gateway, error) {
	kc, err := newKubernetesClient(cfg.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	krc := &kubernetesReverseProxyConfigGetterConfig{
		AnnotationPrefix: cfg.AnnotationPrefix,
		ClusterZone:      cfg.ClusterZone,
		ListenPort:       cfg.HTTPListenPort,
	}
	rg := newReverseProxyConfigGetter(kc, krc)

	nginxCfg := newNGINXConfig(cfg.NGINXHealthPort, cfg.ClusterZone)
	var nm NGINXManager
	if cfg.NGINXDryRun {
		nm = newLoggingNGINXManager()
	} else {
		nm = newNGINXManager(nginxCfg)
	}
	logger.Log.Info("Using nginx config: %+v", nginxCfg)

	gw := Gateway{
		cfg: cfg,
		rg:  rg,
		nm:  nm,
	}

	return &gw, nil
}

type Gateway struct {
	cfg Config
	rg  ReverseProxyConfigGetter
	nm  NGINXManager
}

func (gw *Gateway) start() error {
	ok, err := gw.nginxIsRunning()
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	rc := DefaultReverseProxyConfig(&gw.cfg)
	if err := gw.nm.WriteConfig(rc); err != nil {
		return err
	}

	if err := gw.nm.Start(); err != nil {
		return err
	}

	gw.startHTTPServer()

	return nil
}

func (gw *Gateway) startHTTPServer() {
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", gw.cfg.FarvaHealthPort),
		Handler: health.NewHandler(),
	}

	go func() {
		logger.Log.Fatal(s.ListenAndServe())
	}()
}

func (gw *Gateway) nginxIsRunning() (bool, error) {
	logger.Log.Info("Checking if nginx is running")
	st, err := gw.nm.Status()
	if err != nil {
		return false, err
	}
	return st == nginxStatusRunning, nil
}

func (gw *Gateway) refresh() error {
	logger.Log.Info("Refreshing nginx config")
	rc, err := gw.rg.ReverseProxyConfig()
	if err != nil {
		return err
	}

	rc.HTTPServers = append(rc.HTTPServers, DefaultHTTPReverseProxyServers(&gw.cfg)...)

	if err := gw.nm.WriteConfig(rc); err != nil {
		return err
	}
	if err := gw.nm.Reload(); err != nil {
		return err
	}

	return nil
}

func (gw *Gateway) Run() error {
	if err := gw.start(); err != nil {
		return err
	}

	logger.Log.Info("Gateway started successfully, entering refresh loop")

	ticker := time.NewTicker(gw.cfg.RefreshInterval)

	for {
		if err := gw.refresh(); err != nil {
			logger.Log.Infof("Failed refreshing Gateway: %v", err)
		}

		//NOTE(bcwaldon): receive from the ticker at the
		// end of the loop to emulate do-while semantics.
		<-ticker.C
	}
}
