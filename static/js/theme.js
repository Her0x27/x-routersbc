/**
 * Theme management for RouterSBC
 * Handles dark/light theme switching and user preferences
 */

(function() {
    'use strict';
    
    // Theme configuration
    const THEMES = {
        LIGHT: 'light',
        DARK: 'dark',
        AUTO: 'auto'
    };
    
    const STORAGE_KEY = 'routersbc-theme';
    
    class ThemeManager {
        constructor() {
            this.currentTheme = this.getStoredTheme() || THEMES.AUTO;
            this.init();
        }
        
        init() {
            this.applyTheme(this.currentTheme);
            this.setupEventListeners();
            this.createThemeToggle();
        }
        
        /**
         * Get theme from localStorage
         */
        getStoredTheme() {
            return localStorage.getItem(STORAGE_KEY);
        }
        
        /**
         * Store theme in localStorage
         */
        setStoredTheme(theme) {
            localStorage.setItem(STORAGE_KEY, theme);
        }
        
        /**
         * Get preferred color scheme from system
         */
        getPreferredTheme() {
            if (this.currentTheme !== THEMES.AUTO) {
                return this.currentTheme;
            }
            
            return window.matchMedia('(prefers-color-scheme: dark)').matches ? THEMES.DARK : THEMES.LIGHT;
        }
        
        /**
         * Apply theme to document
         */
        applyTheme(theme) {
            const actualTheme = theme === THEMES.AUTO ? this.getPreferredTheme() : theme;
            
            document.documentElement.setAttribute('data-bs-theme', actualTheme);
            
            // Update theme toggle icons
            this.updateThemeToggle(theme);
            
            // Dispatch theme change event
            window.dispatchEvent(new CustomEvent('themeChanged', {
                detail: { theme: actualTheme, setting: theme }
            }));
        }
        
        /**
         * Switch to next theme in cycle
         */
        toggleTheme() {
            const themes = [THEMES.LIGHT, THEMES.DARK, THEMES.AUTO];
            const currentIndex = themes.indexOf(this.currentTheme);
            const nextIndex = (currentIndex + 1) % themes.length;
            
            this.setTheme(themes[nextIndex]);
        }
        
        /**
         * Set specific theme
         */
        setTheme(theme) {
            if (!Object.values(THEMES).includes(theme)) {
                console.warn('Invalid theme:', theme);
                return;
            }
            
            this.currentTheme = theme;
            this.setStoredTheme(theme);
            this.applyTheme(theme);
        }
        
        /**
         * Setup event listeners
         */
        setupEventListeners() {
            // Listen for system theme changes
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
                if (this.currentTheme === THEMES.AUTO) {
                    this.applyTheme(THEMES.AUTO);
                }
            });
            
            // Listen for storage changes (multi-tab support)
            window.addEventListener('storage', (e) => {
                if (e.key === STORAGE_KEY) {
                    this.currentTheme = e.newValue || THEMES.AUTO;
                    this.applyTheme(this.currentTheme);
                }
            });
        }
        
        /**
         * Create theme toggle button
         */
        createThemeToggle() {
            // Check if theme toggle already exists
            if (document.getElementById('theme-toggle')) {
                return;
            }
            
            // Create toggle button
            const toggle = document.createElement('button');
            toggle.id = 'theme-toggle';
            toggle.className = 'btn btn-ghost-secondary btn-icon';
            toggle.setAttribute('title', 'Toggle theme');
            toggle.setAttribute('data-bs-toggle', 'tooltip');
            toggle.innerHTML = this.getThemeIcon(this.currentTheme);
            
            toggle.addEventListener('click', () => {
                this.toggleTheme();
            });
            
            // Add to navbar if it exists
            const navbar = document.querySelector('.navbar-nav');
            if (navbar) {
                const li = document.createElement('li');
                li.className = 'nav-item';
                li.appendChild(toggle);
                navbar.appendChild(li);
            }
        }
        
        /**
         * Update theme toggle button
         */
        updateThemeToggle(theme) {
            const toggle = document.getElementById('theme-toggle');
            if (toggle) {
                toggle.innerHTML = this.getThemeIcon(theme);
                
                // Update tooltip
                const tooltip = bootstrap.Tooltip.getInstance(toggle);
                if (tooltip) {
                    tooltip.setContent({ '.tooltip-inner': this.getThemeTooltip(theme) });
                }
            }
        }
        
        /**
         * Get theme icon
         */
        getThemeIcon(theme) {
            const icons = {
                [THEMES.LIGHT]: `
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <circle cx="12" cy="12" r="4"/>
                        <path d="m3 12h1m8 -9v1m8 8h1m-9 8v1m-6.4 -15.4l.7 .7m12.1 -.7l-.7 .7m0 11.4l.7 .7m-12.1 -.7l-.7 .7"/>
                    </svg>
                `,
                [THEMES.DARK]: `
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <path d="M12 3c.132 0 .263 0 .393 0a7.5 7.5 0 0 0 7.92 12.446a9 9 0 1 1 -8.313 -12.446z"/>
                    </svg>
                `,
                [THEMES.AUTO]: `
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <rect x="2" y="3" width="20" height="14" rx="2"/>
                        <line x1="8" y1="21" x2="16" y2="21"/>
                        <line x1="12" y1="17" x2="12" y2="21"/>
                    </svg>
                `
            };
            
            return icons[theme] || icons[THEMES.AUTO];
        }
        
        /**
         * Get theme tooltip text
         */
        getThemeTooltip(theme) {
            const tooltips = {
                [THEMES.LIGHT]: 'Switch to dark theme',
                [THEMES.DARK]: 'Switch to auto theme',
                [THEMES.AUTO]: 'Switch to light theme'
            };
            
            return tooltips[theme] || 'Toggle theme';
        }
    }
    
    // Initialize theme manager
    window.themeManager = new ThemeManager();
    
    // Expose theme functions globally
    window.setTheme = (theme) => window.themeManager.setTheme(theme);
    window.toggleTheme = () => window.themeManager.toggleTheme();
    window.getCurrentTheme = () => window.themeManager.getPreferredTheme();
    
})();
