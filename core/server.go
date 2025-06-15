package core

import (
	"html/template"
	"io"
	"net/http"

	"github.com/Her0x27/x-routersbc/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer implements echo.Renderer interface
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// NewServer creates and configures a new Echo server instance
func NewServer() *echo.Echo {
	e := echo.New()

	// Configure template renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*/*.html")),
	}
	e.Renderer = renderer

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static files
	e.Static("/static", "static")

	// Custom error handler
	e.HTTPErrorHandler = customErrorHandler

	// Setup routes
	routes.SetupRoutes(e)

	return e
}

// customErrorHandler handles HTTP errors with custom pages
func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	// Try to render custom error page
	switch code {
	case 404:
		if err := c.Render(code, "404.html", map[string]interface{}{
			"Title":   "Page Not Found",
			"Message": "The requested page could not be found.",
		}); err != nil {
			c.String(code, message)
		}
	case 501:
		if err := c.Render(code, "501.html", map[string]interface{}{
			"Title":   "Not Implemented",
			"Message": "This feature is not yet implemented.",
		}); err != nil {
			c.String(code, message)
		}
	default:
		c.String(code, message)
	}
}
