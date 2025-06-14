#!/bin/bash

# Скрипт генерации модуля для Go + Echo проекта
# Использование: ./generate_module.sh <project_name> <module_name>

set -e

# Проверка параметров
if [ $# -ne 2 ]; then
    echo "Использование: $0 <project_name> <module_name>"
    echo "Пример: $0 network-router dashboard"
    exit 1
fi

PROJECT_NAME=$1
MODULE_NAME=$2
MODULE_TITLE=$(echo "$MODULE_NAME" | sed 's/.*/\u&/')

echo "🚀 Генерация модуля '$MODULE_NAME' для проекта '$PROJECT_NAME'"

# Создание структуры директорий
echo "📁 Создание структуры директорий..."
mkdir -p handlers
mkdir -p routes
mkdir -p models
mkdir -p services
mkdir -p templates/$MODULE_NAME
mkdir -p static/css/$MODULE_NAME
mkdir -p static/js/$MODULE_NAME
mkdir -p tests/handlers
mkdir -p docs

# Генерация обработчика
echo "🔧 Создание обработчика..."
cat > handlers/${MODULE_NAME}_handler.go << EOF
package handlers

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

// ${MODULE_TITLE}Handler структура для обработчиков модуля $MODULE_NAME
type ${MODULE_TITLE}Handler struct {
	// Добавить зависимости (сервисы, репозитории)
}

// New${MODULE_TITLE}Handler создает новый экземпляр обработчика
func New${MODULE_TITLE}Handler() *${MODULE_TITLE}Handler {
	return &${MODULE_TITLE}Handler{}
}

// Index отображает главную страницу модуля $MODULE_NAME
func (h *${MODULE_TITLE}Handler) Index(c echo.Context) error {
	data := map[string]interface{}{
		"Title":      "$MODULE_TITLE",
		"ModuleName": "$MODULE_NAME",
		"PageTitle":  "$MODULE_TITLE Dashboard",
	}
	
	return c.Render(http.StatusOK, "${MODULE_NAME}/index.html", data)
}

// GetData возвращает данные модуля в формате JSON (API endpoint)
func (h *${MODULE_TITLE}Handler) GetData(c echo.Context) error {
	// TODO: Реализовать получение данных
	data := map[string]interface{}{
		"status": "success",
		"module": "$MODULE_NAME",
		"data":   []interface{}{},
	}
	
	return c.JSON(http.StatusOK, data)
}

// Update обновляет настройки модуля
func (h *${MODULE_TITLE}Handler) Update(c echo.Context) error {
	// TODO: Реализовать обновление настроек
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "$MODULE_TITLE settings updated successfully",
	})
}
EOF

# Генерация маршрутов
echo "🛣️  Создание маршрутов..."
cat > routes/${MODULE_NAME}_routes.go << EOF
package routes

import (
	"${PROJECT_NAME}/handlers"
	"github.com/labstack/echo/v4"
)

// Register${MODULE_TITLE}Routes регистрирует маршруты для модуля $MODULE_NAME
func Register${MODULE_TITLE}Routes(e *echo.Echo) {
	handler := handlers.New${MODULE_TITLE}Handler()
	
	// Группа маршрутов для модуля
	${MODULE_NAME}Group := e.Group("/$MODULE_NAME")
	
	// HTML страницы
	${MODULE_NAME}Group.GET("", handler.Index)
	${MODULE_NAME}Group.GET("/", handler.Index)
	
	// API endpoints
	api := ${MODULE_NAME}Group.Group("/api")
	api.GET("/data", handler.GetData)
	api.POST("/update", handler.Update)
	api.PUT("/settings", handler.Update)
}
EOF

# Генерация модели
echo "📊 Создание модели..."
cat > models/${MODULE_NAME}_model.go << EOF
package models

import (
	"time"
)

// ${MODULE_TITLE}Config структура конфигурации модуля $MODULE_NAME
type ${MODULE_TITLE}Config struct {
	ID        int       \`json:"id"\`
	Name      string    \`json:"name"\`
	Enabled   bool      \`json:"enabled"\`
	Settings  map[string]interface{} \`json:"settings"\`
	CreatedAt time.Time \`json:"created_at"\`
	UpdatedAt time.Time \`json:"updated_at"\`
}

// ${MODULE_TITLE}Status структура статуса модуля
type ${MODULE_TITLE}Status struct {
	Module    string \`json:"module"\`
	Status    string \`json:"status"\`
	Uptime    int64  \`json:"uptime"\`
	LastCheck time.Time \`json:"last_check"\`
}

// ${MODULE_TITLE}Data основная структура данных модуля
type ${MODULE_TITLE}Data struct {
	Config *${MODULE_TITLE}Config \`json:"config"\`
	Status *${MODULE_TITLE}Status \`json:"status"\`
	Metrics map[string]interface{} \`json:"metrics"\`
}

// Validate валидирует данные модуля
func (d *${MODULE_TITLE}Data) Validate() error {
	// TODO: Добавить валидацию
	return nil
}
EOF

# Генерация сервиса
echo "⚙️  Создание сервиса..."
cat > services/${MODULE_NAME}_service.go << EOF
package services

import (
	"${PROJECT_NAME}/models"
	"fmt"
)

// ${MODULE_TITLE}Service сервис для работы с модулем $MODULE_NAME
type ${MODULE_TITLE}Service struct {
	// Добавить зависимости
}

// New${MODULE_TITLE}Service создает новый экземпляр сервиса
func New${MODULE_TITLE}Service() *${MODULE_TITLE}Service {
	return &${MODULE_TITLE}Service{}
}

// GetData получает данные модуля
func (s *${MODULE_TITLE}Service) GetData() (*models.${MODULE_TITLE}Data, error) {
	// TODO: Реализовать получение данных
	config := &models.${MODULE_TITLE}Config{
		ID:      1,
		Name:    "$MODULE_NAME",
		Enabled: true,
		Settings: make(map[string]interface{}),
	}
	
	status := &models.${MODULE_TITLE}Status{
		Module: "$MODULE_NAME",
		Status: "active",
		Uptime: 0,
	}
	
	data := &models.${MODULE_TITLE}Data{
		Config:  config,
		Status:  status,
		Metrics: make(map[string]interface{}),
	}
	
	return data, nil
}

// UpdateConfig обновляет конфигурацию модуля
func (s *${MODULE_TITLE}Service) UpdateConfig(config *models.${MODULE_TITLE}Config) error {
	// TODO: Реализовать обновление конфигурации
	if err := config.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	return nil
}

// GetStatus получает статус модуля
func (s *${MODULE_TITLE}Service) GetStatus() (*models.${MODULE_TITLE}Status, error) {
	// TODO: Реализовать получение статуса
	status := &models.${MODULE_TITLE}Status{
		Module: "$MODULE_NAME",
		Status: "active",
		Uptime: 0,
	}
	
	return status, nil
}
EOF

# Генерация HTML шаблона
echo "🎨 Создание HTML шаблона..."
cat > templates/$MODULE_NAME/index.html << EOF
{{define "content_${MODULE_NAME}"}}
<div class="page-wrapper">
    <!-- Page header -->
    <div class="page-header d-print-none">
        <div class="container-xl">
            <div class="row g-2 align-items-center">
                <div class="col">
                    <div class="page-pretitle">
                        Module
                    </div>
                    <h2 class="page-title">
                        {{.PageTitle}}
                    </h2>
                </div>
                <div class="col-auto ms-auto d-print-none">
                    <div class="btn-list">
                        <button type="button" class="btn btn-primary" onclick="${MODULE_NAME}Module.refresh()">
                            <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                                <path d="M20 11a8.1 8.1 0 0 0 -15.5 -2m-.5 -4v4h4"/>
                                <path d="M4 13a8.1 8.1 0 0 0 15.5 2m.5 4v-4h-4"/>
                            </svg>
                            Refresh
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Page body -->
    <div class="page-body">
        <div class="container-xl">
            <div class="row row-deck row-cards">
                <!-- Status Card -->
                <div class="col-12 col-md-6 col-lg-4">
                    <div class="card">
                        <div class="card-body">
                            <div class="d-flex align-items-center">
                                <div class="subheader">Status</div>
                                <div class="ms-auto">
                                    <div class="status status-green" id="${MODULE_NAME}-status"></div>
                                </div>
                            </div>
                            <div class="h1 mb-3" id="${MODULE_NAME}-status-text">Active</div>
                            <div class="d-flex mb-2">
                                <div>Uptime</div>
                                <div class="ms-auto">
                                    <span class="text-green d-inline-flex align-items-center lh-1" id="${MODULE_NAME}-uptime">
                                        0s
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Configuration Card -->
                <div class="col-12 col-md-6 col-lg-8">
                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">Configuration</h3>
                        </div>
                        <div class="card-body">
                            <form id="${MODULE_NAME}-config-form">
                                <div class="mb-3">
                                    <label class="form-label">Module Name</label>
                                    <input type="text" class="form-control" name="name" value="{{.ModuleName}}" readonly>
                                </div>
                                <div class="mb-3">
                                    <label class="form-check">
                                        <input type="checkbox" class="form-check-input" name="enabled" checked>
                                        <span class="form-check-label">Enable Module</span>
                                    </label>
                                </div>
                                <div class="form-footer">
                                    <button type="submit" class="btn btn-primary">Save Configuration</button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>

                <!-- Data Table -->
                <div class="col-12">
                    <div class="card">
                        <div class="card-header">
                            <h3 class="card-title">{{.Title}} Data</h3>
                        </div>
                        <div class="card-body">
                            <div class="table-responsive">
                                <table class="table table-vcenter card-table" id="${MODULE_NAME}-data-table">
                                    <thead>
                                        <tr>
                                            <th>Parameter</th>
                                            <th>Value</th>
                                            <th>Status</th>
                                            <th class="w-1">Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <tr>
                                            <td colspan="4" class="text-center text-muted">
                                                Loading data...
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<!-- Подключение модульного CSS и JS -->
<link rel="stylesheet" href="/static/css/{{.ModuleName}}/{{.ModuleName}}.css">
<script src="/static/js/{{.ModuleName}}/{{.ModuleName}}.js"></script>
{{end}}
EOF

# Генерация CSS
echo "🎨 Создание CSS стилей..."
cat > static/css/$MODULE_NAME/${MODULE_NAME}.css << EOF
/* Стили для модуля $MODULE_NAME */
/* Используем Tabler.io как основу */

.${MODULE_NAME}-module {
    /* Основные стили модуля */
}

.${MODULE_NAME}-status-card {
    transition: all 0.3s ease;
}

.${MODULE_NAME}-status-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.${MODULE_NAME}-data-table {
    /* Стили для таблицы данных */
}

.${MODULE_NAME}-config-form {
    /* Стили для формы конфигурации */
}

/* Анимации загрузки */
.${MODULE_NAME}-loading {
    opacity: 0.6;
    pointer-events: none;
}

.${MODULE_NAME}-loading::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 20px;
    height: 20px;
    margin: -10px 0 0 -10px;
    border: 2px solid #f3f3f3;
    border-top: 2px solid #3498db;
    border-radius: 50%;
    animation: ${MODULE_NAME}-spin 1s linear infinite;
}

@keyframes ${MODULE_NAME}-spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}

/* Responsive стили */
@media (max-width: 768px) {
    .${MODULE_NAME}-module .card {
        margin-bottom: 1rem;
    }
}
EOF
