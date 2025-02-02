package server

import (
	_ "clusterMonitor/docs"
	apiExposedApp "clusterMonitor/internal/application/domain/ports"
	apiExposedCl "clusterMonitor/internal/cluster/domain/ports"
	configPort "clusterMonitor/pkg/config/domain/ports"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"os"
)

type Server struct {
	app          *fiber.App
	networkInfo  NetworkInfo
	helloApp     apiExposedApp.ApiExposed
	helloCl      apiExposedCl.HelloHandlers
	configurator configPort.ConfigApplication
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path"},
	)
)

func init() {
	// Registra le metriche con il registratore Prometheus
	prometheus.MustRegister(httpRequestsTotal)
}

func NewServer(hello_app apiExposedApp.ApiExposed, hello_cl apiExposedCl.HelloHandlers, configurator configPort.ConfigApplication) *Server {

	return &Server{helloApp: hello_app, helloCl: hello_cl, configurator: configurator}
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v2
func (s *Server) Run(port string) error {
	app := fiber.New()
	app.Use(cors.New())
	//app.Use(logger.New())
	app.Use(logger.New(logger.Config{
		Format:     `{"time":"${time}", "ip":"${ip}", "port":"${port}", "status":${status}, "method":"${method}", "path":"${path}", "latency":"${latency}"}` + "\n",
		Output:     os.Stdout,
		TimeFormat: "2006-01-02T15:04:05Z07:00", // ISO 8601 format for time
	}))

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	app.Get("/metrics", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(c.Context())
		return nil
	})
	v1Application := app.Group("/api/v1/application")

	// User Endpoints
	v1Application.Post("", s.helloApp.AddApplicationBySvc)

	v1Application.Put("/monitor", s.helloApp.MonitoringApplication)
	v1Application.Get("/monitor", s.helloApp.GetApplicationMonitored)
	v1Application.Get("/monitor/pod", s.helloApp.GetApplicationMonitoredByPod)

	v1Application.Delete("/monitor/:service", s.helloApp.UnscheduleApplication)
	//v1Application.Post("/check", s.helloApp.Check)

	v1Cluster := app.Group("/api/v1/cluster")

	// User Endpoints
	v1Cluster.Get("/hello", s.helloCl.SalutaDfiPiu)

	s.app = app
	err := app.Listen(":" + port)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	err := s.app.Shutdown()
	if err != nil {
		return err
	}
	return nil
}
