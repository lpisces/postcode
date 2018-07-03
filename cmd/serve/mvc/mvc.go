package mvc

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/lpisces/postcode/cmd/serve"
	"gopkg.in/urfave/cli.v1"
	"net"
)

// serve start web server
func startSrv() (err error) {

	// get config
	config := serve.Conf

	// new echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	//e.Use(middleware.CSRF())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.Secret.Session))))

	// Routes
	Route(e)

	// Start server
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.Srv.Host, config.Srv.Port))
	if err != nil {
		return err
	}

	e.Listener = l
	e.HideBanner = true

	if serve.Debug {
		e.Logger.SetLevel(log.DEBUG)
	} else {
		e.Logger.SetLevel(log.ERROR)
	}

	e.Logger.Infof("http server started on %s:%s in %s model", config.Srv.Host, config.Srv.Port, config.Mode)
	e.Logger.Fatal(e.Start(""))
	return
}

func Run(c *cli.Context) (err error) {

	// Load default config
	config := serve.DefaultConfig()

	// override default config
	configFilePath := c.String("config")
	if configFilePath != "" {
		if err := config.Load(configFilePath); err != nil {
			log.Fatal(err)
		}
	}

	// flag override ini file config
	bind := c.String("bind")
	if bind != "" {
		config.Srv.Host = bind
	}

	port := c.String("port")
	if port != "" {
		config.Srv.Port = port
	}

	env := c.String("env")
	if env != "" {
		config.Mode = env
	}

	source := c.String("source")
	if source != "" {
		config.Source = source
	}

	serve.Conf = config

	// run mode
	if config.Mode != "production" {
		serve.Debug = true
	}

	// start server
	err = startSrv()
	if err != nil {
		log.Fatal(err)
	}
	return
}
