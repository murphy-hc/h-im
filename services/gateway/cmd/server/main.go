package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	pkgmetrics "github.com/murphy-hc/h-im/pkg/metrics"
	pkgtracing "github.com/murphy-hc/h-im/pkg/tracing"
	"github.com/rs/xid"

	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/server"
)

var (
	Name     string = "him-gateway"
	Version  string
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(ws *server.WSServer, hs *khttp.Server, gs *kratosgrpc.Server, pss *server.PubSubServer) *kratos.App {
	id := xid.New().String()
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(log.DefaultLogger),
		kratos.Server(ws, hs, gs, pss),
	)
}

func main() {
	flag.Parse()
	c := config.New(config.WithSource(file.NewSource(flagconf)))
	defer c.Close()
	if err := c.Load(); err != nil {
		panic(err)
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	otelCfg := bc.GetOtel()
	timeout := otelCfg.GetTimeout().AsDuration()
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	shutdown, err := pkgtracing.InitTracer(&pkgtracing.Config{
		ServiceName: otelCfg.GetServiceName(),
		Endpoint:    otelCfg.GetEndpoint(),
		Rate:        otelCfg.GetSampleRate(),
		Path:        otelCfg.GetPath(),
		Timeout:     timeout,
		Insecure:    otelCfg.GetInsecure(),
		Transport:   pkgtracing.Transport(otelCfg.GetTransport()),
	})
	if err != nil {
		log.Fatalf("init tracer: %v", err)
	}
	defer shutdown()
	meter, err := pkgmetrics.NewPrometheusMeter(Name, bc.GetServer().GetEnv())
	if err != nil {
		log.Fatalf("init metrics: %v", err)
	}
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.Caller(4),
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	logger = log.NewFilter(logger, log.FilterLevel(log.ParseLevel(bc.GetLog().GetLevel())))
	log.SetLogger(logger)
	app, cleanup, err := wireApp(&bc, meter)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	// Graceful shutdown on SIGTERM/SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Infof("gateway: shutting down...")
		app.Stop()
	}()
	if err := app.Run(); err != nil {
		log.Errorf("gateway: %v", err)
	}
}
