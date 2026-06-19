package main

import (
	"flag"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	pkgmetrics "github.com/murphy-hc/h-im/pkg/metrics"
	pkgtracing "github.com/murphy-hc/h-im/pkg/tracing"
	"github.com/rs/xid"

	"github.com/murphy-hc/h-im/services/auth/internal/conf"
)

var (
	Name     string = "him-auth"
	Version  string
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(gs *grpc.Server, hs *http.Server) *kratos.App {
	id := xid.New().String()
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(log.DefaultLogger),
		kratos.Server(gs, hs),
	)
}

func main() {
	flag.Parse()
	c := config.New(config.WithSource(file.NewSource(flagconf)))
	defer c.Close()
	if err := c.Load(); err != nil { panic(err) }
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil { panic(err) }
	otelCfg := bc.GetOtel()
	timeout := otelCfg.GetTimeout().AsDuration()
	if timeout <= 0 { timeout = 5 * time.Second }
	shutdown, err := pkgtracing.InitTracer(&pkgtracing.Config{
		ServiceName: otelCfg.GetServiceName(),
		Endpoint:    otelCfg.GetEndpoint(),
		Rate:        otelCfg.GetSampleRate(),
		Path:        otelCfg.GetPath(),
		Timeout:     timeout,
		Insecure:    otelCfg.GetInsecure(),
		Transport:   pkgtracing.Transport(otelCfg.GetTransport()),
	})
	if err != nil { log.Fatalf("init tracer: %v", err) }
	defer shutdown()
	meter, err := pkgmetrics.NewPrometheusMeter(Name, bc.GetServer().GetEnv())
	if err != nil { log.Fatalf("init metrics: %v", err) }
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
	if err != nil { panic(err) }
	defer cleanup()
	if err := app.Run(); err != nil { panic(err) }
}
