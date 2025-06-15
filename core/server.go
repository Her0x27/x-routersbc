package core

import (
        "database/sql"
        "html/template"
        "io"
        "net/http"

        "github.com/labstack/echo/v4"
        "github.com/labstack/echo/v4/middleware"
)

type Server struct {
        Echo *echo.Echo
        DB   *sql.DB
}

type TemplateRenderer struct {
        templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
        return t.templates.ExecuteTemplate(w, name, data)
}

func NewServer(db *sql.DB) *Server {
        e := echo.New()
        
        // Set custom error handler
        e.HTTPErrorHandler = customHTTPErrorHandler
        
        // Load templates
        t := &TemplateRenderer{
                templates: template.Must(template.ParseGlob("templates/**/*.html")),
        }
        e.Renderer = t
        
        // Middleware
        e.Use(middleware.Logger())
        e.Use(middleware.Recover())
        e.Use(middleware.CORS())
        e.Use(SessionMiddleware(db))
        
        // Static files
        e.Static("/static", "static")
        
        // Create server instance
        server := &Server{
                Echo: e,
                DB:   db,
        }
        
        // Setup WebSocket
        SetupWebSocket(e, db)
        
        return server
}

func (s *Server) Start(address string) error {
        return s.Echo.Start(address)
}

// SetupRoutes allows external packages to configure routes
func (s *Server) SetupRoutes(setupFunc func(*echo.Echo, *sql.DB)) {
        setupFunc(s.Echo, s.DB)
}

func customHTTPErrorHandler(err error, c echo.Context) {
        code := http.StatusInternalServerError
        if he, ok := err.(*echo.HTTPError); ok {
                code = he.Code
        }
        
        switch code {
        case 404:
                c.Render(code, "404.html", map[string]interface{}{
                        "Title": "Page Not Found",
                        "Error": "The requested page could not be found.",
                })
        case 501:
                c.Render(code, "501.html", map[string]interface{}{
                        "Title": "Not Implemented",
                        "Error": "This feature is not yet implemented.",
                })
        default:
                c.JSON(code, map[string]string{
                        "error": err.Error(),
                })
        }
}
