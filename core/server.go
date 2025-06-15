package core

import (
	"html/template"
	"io"
	"net/http"

	"github.com/Her0x27/x-routersbc/handlers"
	"github.com/Her0x27/x-routersbc/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global template data
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

// Server represents the main server instance
type Server struct {
	echo     *echo.Echo
	database *Database
	auth     *AuthManager
	ws       *WebSocketManager
}

// NewServer creates a new server instance
func NewServer() *Server {
	e := echo.New()
	
	// Initialize components
	db, err := NewDatabase()
	if err != nil {
		e.Logger.Fatal("Failed to initialize database:", err)
	}
	
	authManager := NewAuthManager(db)
	wsManager := NewWebSocketManager()

	// Setup middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	
	// Custom error handler
	e.HTTPErrorHandler = customErrorHandler
	
	// Setup template renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/**/*.html")),
	}
	e.Renderer = renderer
	
	// Serve static files
	e.Static("/static", "static")
	
	server := &Server{
		echo:     e,
		database: db,
		auth:     authManager,
		ws:       wsManager,
	}
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authManager, db)
	networkHandler := handlers.NewNetworkHandler()
	systemHandler := handlers.NewSystemHandler()
	
	// Setup routes
	routes.SetupRoutes(e, authHandler, networkHandler, systemHandler, wsManager)
	
	return server
}

// Start starts the server
func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}

// customErrorHandler handles errors globally
func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"
	
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
	}
	
	// Log the error
	c.Logger().Error(err)
	
	// Render error page
	switch code {
	case 404:
		if err := c.Render(code, "404.html", map[string]interface{}{
			"title": "Page Not Found",
			"error": message,
		}); err != nil {
			c.Logger().Error(err)
		}
	case 501:
		if err := c.Render(code, "501.html", map[string]interface{}{
			"title": "Not Implemented",
			"error": message,
		}); err != nil {
			c.Logger().Error(err)
		}
	default:
		if err := c.Render(code, "404.html", map[string]interface{}{
			"title": "Error",
			"error": message,
		}); err != nil {
			c.Logger().Error(err)
		}
	}
}
