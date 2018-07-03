package c

import (
	"github.com/labstack/echo"
	"html/template"
	"io"
	"os"
	"path/filepath"
)

type TestTemplate struct {
	templates *template.Template
}

func (t *TestTemplate) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func initTestEcho() *echo.Echo {
	e := echo.New()

	viewPath := "../v"
	templates := template.New("")

	if _, err := os.Stat(viewPath); err == nil {
		err = filepath.Walk(viewPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				_, err := templates.ParseGlob(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			e.Logger.Fatal(err)
		}
	}

	tp := &TestTemplate{
		templates: templates,
	}
	e.Renderer = tp

	return e
}
