Вот структурированный промпт на основе вашего описания:

## **ЧТО НУЖНО СДЕЛАТЬ:**
Создать полноценную веб-систему управления сетевым оборудованием с модульной архитектурой, включающую документацию, инфраструктуру разработки, систему контроля качества кода и веб-интерфейс для управления сетевыми настройками.

## **ГДЕ:**
- **Корневая документация**: `README.md`
- **Инфраструктура**: `.replit`, `workflows/`
- **Документация разработки**: `dev/RULES.md`, `dev/DEV_TASKS.md`
- **Система контроля**: `.backup/`, скрипты валидации
- **Backend**: Go модули с Echo Framework
- **Frontend**: HTML шаблоны в `templates/`
  - `templates/base/` - базовые компоненты
  - `templates/components/` - переиспользуемые элементы
  - `templates/[module_name]/` - модульные страницы

## **КАК:**
- **Backend**: Go 1.23 + Echo Framework с поддержкой HTTP/2 и WebSocket
- **Шаблонизация**: HTML templates с уникальными именами `{{define "content_[module_name]"}}`
- **Архитектура**: Модульная структура с базовыми шаблонами (header, sidebar, layout, footer)
- **Маршрутизация**: RESTful API с четкой структурой URL
- **Визуализация**: jsVectorMap для отображения сетевой топологии
- **Контроль версий**: Автоматическое создание backup'ов перед изменениями

## **ОГРАНИЧЕНИЯ:**
- **ЗАПРЕЩЕНО** использовать фиктивные/симулированные данные
- **ЗАПРЕЩЕНО** игнорировать ошибки в коде
- **ЗАПРЕЩЕНО** писать упрощенный код без proper error handling
- **ЗАПРЕЩЕНО** изменять код без создания backup копии
- **ЗАПРЕЩЕНО** коммитить без подробного описания изменений

## **ДОПОЛНИТЕЛЬНЫЕ ТРЕБОВАНИЯ:**

### **Документация:**
- Подробные комментарии в коде на русском языке
- Описание решения каждой поставленной задачи
- Ведение reports о всех изменениях (что добавлено/изменено/удалено)
- Обязательное указание причины каждого редактирования

### **Система контроля качества:**
- Shell скрипт для автоматической проверки качества кода
- Парсер внутренних URL с проверкой доступности
- Валидация содержимого ответов API
- Система напоминаний о правилах разработки при каждом запуске

### **Модули для реализации:**
1. **Dashboard** (`/dashboard`) - визуализация сети, статус WAN/LAN/Wireless
2. **System Settings** (`/system/settings`) - пароли, NTP, backup/restore
3. **Hardware Info** (`/system/hardware`) - информация о железе
4. **Device Detection** (`/system/devices_detected`) - обнаружение устройств
5. **Network Overview** (`/network/`) - общая информация о сети
6. **WAN Management** (`/network/wans`) - управление WAN подключениями
7. **LAN Settings** (`/network/lan`) - DHCP, DNS настройки
8. **Firewall** (`/network/firewall`) - nftables/iptables
9. **Wireless** (`/network/wireless`) - WiFi настройки

### **Инфраструктура:**
- Настройка Replit для Go 1.23
- Debug и Release конфигурации
- Автоматические workflow для CI/CD
- Система backup'ов с версионированием

Этот промпт обеспечивает четкую структуру для создания профессиональной системы управления сетевым оборудованием с высокими стандартами качества кода и документации.
