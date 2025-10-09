package swagger

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// LoginFormTestSuite defines a test suite for loginform functions
type LoginFormTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupTest runs before each test
func (suite *LoginFormTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
}

// TestServeSwaggerWithLogin tests the ServeSwaggerWithLogin function
func (suite *LoginFormTestSuite) TestServeSwaggerWithLogin() {
	config := LoginConfig{
		Title:   "Login Test API",
		SpecURL: "/swagger/doc.json",
	}
	
	handler := ServeSwaggerWithLogin(config)
	suite.router.GET("/swagger-login", handler)
	
	req, err := http.NewRequest("GET", "/swagger-login", nil)
	require.NoError(suite.T(), err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	
	body := w.Body.String()
	assert.Contains(suite.T(), body, "Login Test API")
	assert.Contains(suite.T(), body, "/swagger/doc.json")
	assert.Contains(suite.T(), body, "swagger-ui-bundle.js")
	assert.Contains(suite.T(), body, "performLogin")
}

// TestServeSwaggerWithLoginDefaults tests ServeSwaggerWithLogin with empty config
func (suite *LoginFormTestSuite) TestServeSwaggerWithLoginDefaults() {
	config := LoginConfig{} // Empty config
	
	handler := ServeSwaggerWithLogin(config)
	suite.router.GET("/swagger-login", handler)
	
	req, err := http.NewRequest("GET", "/swagger-login", nil)
	require.NoError(suite.T(), err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	body := w.Body.String()
	// Check for empty values in template
	assert.Contains(suite.T(), body, "<title></title>")
	assert.Contains(suite.T(), body, "url: ''")
}

// TestServeSwaggerWithLoginPartialConfig tests partial configuration
func (suite *LoginFormTestSuite) TestServeSwaggerWithLoginPartialConfig() {
	config := LoginConfig{
		Title: "Partial Config API",
		// SpecURL left empty
	}
	
	handler := ServeSwaggerWithLogin(config)
	suite.router.GET("/swagger-login", handler)
	
	req, err := http.NewRequest("GET", "/swagger-login", nil)
	require.NoError(suite.T(), err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	body := w.Body.String()
	assert.Contains(suite.T(), body, "Partial Config API")
	assert.Contains(suite.T(), body, "url: ''") // Empty SpecURL
}

// TestServeSwaggerWithLoginHTMLStructure tests the HTML structure
func (suite *LoginFormTestSuite) TestServeSwaggerWithLoginHTMLStructure() {
	config := LoginConfig{
		Title:   "Structure Test API",
		SpecURL: "/api/swagger.json",
	}
	
	handler := ServeSwaggerWithLogin(config)
	suite.router.GET("/swagger-login", handler)
	
	req, err := http.NewRequest("GET", "/swagger-login", nil)
	require.NoError(suite.T(), err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Test HTML structure
	assert.Contains(suite.T(), body, "<!DOCTYPE html>")
	assert.Contains(suite.T(), body, "<html>")
	assert.Contains(suite.T(), body, "<head>")
	assert.Contains(suite.T(), body, "<body>")
	assert.Contains(suite.T(), body, "</html>")
	
	// Test login form elements
	assert.Contains(suite.T(), body, "login-section")
	assert.Contains(suite.T(), body, "Quick Login")
	assert.Contains(suite.T(), body, "login-email")
	assert.Contains(suite.T(), body, "login-password")
	assert.Contains(suite.T(), body, "performLogin()")
	
	// Test Swagger UI elements
	assert.Contains(suite.T(), body, "<div id=\"swagger-ui\">")
	assert.Contains(suite.T(), body, "swagger-ui-bundle.js")
	assert.Contains(suite.T(), body, "swagger-ui-standalone-preset.js")
}

// TestServeSwaggerWithLoginJavaScript tests JavaScript functionality
func (suite *LoginFormTestSuite) TestServeSwaggerWithLoginJavaScript() {
	config := LoginConfig{
		Title:   "JS Test API",
		SpecURL: "/js/swagger.json",
	}
	
	handler := ServeSwaggerWithLogin(config)
	suite.router.GET("/swagger-login", handler)
	
	req, err := http.NewRequest("GET", "/swagger-login", nil)
	require.NoError(suite.T(), err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Test JavaScript functions
	assert.Contains(suite.T(), body, "function performLogin()")
	assert.Contains(suite.T(), body, "function showStatus(")
	assert.Contains(suite.T(), body, "SwaggerUIBundle")
	assert.Contains(suite.T(), body, "/api/v1/auth/user/login")
	assert.Contains(suite.T(), body, "swaggerUI.authActions.authorize")
	
	// Test JavaScript variables
	assert.Contains(suite.T(), body, "let swaggerUI")
	assert.Contains(suite.T(), body, "url: '/js/swagger.json'")
	
	// Test event handlers
	assert.Contains(suite.T(), body, "window.onload")
}

// TestServeSwaggerWithLoginCSS tests CSS styling
func (suite *LoginFormTestSuite) TestServeSwaggerWithLoginCSS() {
	config := LoginConfig{
		Title:   "CSS Test API",
		SpecURL: "/css/swagger.json",
	}
	
	handler := ServeSwaggerWithLogin(config)
	suite.router.GET("/swagger-login", handler)
	
	req, err := http.NewRequest("GET", "/swagger-login", nil)
	require.NoError(suite.T(), err)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Test CSS classes
	assert.Contains(suite.T(), body, ".login-section")
	assert.Contains(suite.T(), body, ".login-form")
	assert.Contains(suite.T(), body, ".login-field")
	assert.Contains(suite.T(), body, ".login-btn")
	assert.Contains(suite.T(), body, ".login-status")
	assert.Contains(suite.T(), body, ".login-status.success")
	assert.Contains(suite.T(), body, ".login-status.error")
	
	// Test CSS styling
	assert.Contains(suite.T(), body, "background: #f8f9fa")
	assert.Contains(suite.T(), body, "background: #007bff")
	assert.Contains(suite.T(), body, "cursor: pointer")
}

// TestLoginConfigStruct tests LoginConfig struct
func (suite *LoginFormTestSuite) TestLoginConfigStruct() {
	config := LoginConfig{
		Title:   "Test Title",
		SpecURL: "Test URL",
	}
	
	assert.Equal(suite.T(), "Test Title", config.Title)
	assert.Equal(suite.T(), "Test URL", config.SpecURL)
}

// TestLoginTemplateExecution tests template execution directly
func (suite *LoginFormTestSuite) TestLoginTemplateExecution() {
	// Test that the template can be parsed and executed without errors
	tmpl, err := template.New("test").Parse(swaggerWithLoginTemplate)
	require.NoError(suite.T(), err)
	
	config := LoginConfig{
		Title:   "Template Test API",
		SpecURL: "/template/swagger.json",
	}
	
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, config)
	require.NoError(suite.T(), err)
	
	result := buf.String()
	assert.Contains(suite.T(), result, "Template Test API")
	assert.Contains(suite.T(), result, "/template/swagger.json")
}

// Run the test suite
func TestLoginFormTestSuite(t *testing.T) {
	suite.Run(t, new(LoginFormTestSuite))
}

// Standalone tests

func TestLoginConfigDefaults(t *testing.T) {
	testCases := []struct {
		name        string
		inputConfig LoginConfig
		description string
	}{
		{
			name:        "Empty config",
			inputConfig: LoginConfig{},
			description: "Should handle empty configuration",
		},
		{
			name: "Title only",
			inputConfig: LoginConfig{
				Title: "Title Only API",
			},
			description: "Should handle title-only configuration",
		},
		{
			name: "SpecURL only",
			inputConfig: LoginConfig{
				SpecURL: "/spec-only/swagger.json",
			},
			description: "Should handle SpecURL-only configuration",
		},
		{
			name: "Full config",
			inputConfig: LoginConfig{
				Title:   "Full Config API",
				SpecURL: "/full/swagger.json",
			},
			description: "Should handle full configuration",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			handler := ServeSwaggerWithLogin(tc.inputConfig)
			router.GET("/test", handler)
			
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code, tc.description)
			assert.NotEmpty(t, w.Body.String(), tc.description)
			
			body := w.Body.String()
			if tc.inputConfig.Title != "" {
				assert.Contains(t, body, tc.inputConfig.Title)
			}
			if tc.inputConfig.SpecURL != "" {
				assert.Contains(t, body, tc.inputConfig.SpecURL)
			}
		})
	}
}

func TestLoginFormSpecialCharacters(t *testing.T) {
	// Test that special characters in config are properly handled
	config := LoginConfig{
		Title:   "API & Documentation <test>",
		SpecURL: "/swagger/doc.json?version=1.0&format=json",
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	
	// The template should properly escape these characters in HTML
	assert.Contains(t, body, "API &amp; Documentation &lt;test&gt;")
	// URLs should be preserved as-is in JavaScript
	assert.Contains(t, body, "/swagger/doc.json?version=1.0&format=json")
}

func TestLoginFormJavaScriptSyntax(t *testing.T) {
	config := LoginConfig{
		Title:   "JavaScript Test",
		SpecURL: "/test/swagger.json",
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Test for proper JavaScript syntax
	assert.Contains(t, body, "async function performLogin()")
	assert.Contains(t, body, "const email = document.getElementById('login-email').value")
	assert.Contains(t, body, "const password = document.getElementById('login-password').value")
	assert.Contains(t, body, "try {")
	assert.Contains(t, body, "} catch (error) {")
	assert.Contains(t, body, "await fetch(")
	
	// Test for proper error handling
	assert.Contains(t, body, "if (!email || !password)")
	assert.Contains(t, body, "if (!response.ok)")
	assert.Contains(t, body, "if (!token)")
}

func TestLoginFormInputValidation(t *testing.T) {
	config := LoginConfig{
		Title:   "Validation Test",
		SpecURL: "/validation/swagger.json",
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Test for input validation in JavaScript
	assert.Contains(t, body, "if (!email || !password)")
	assert.Contains(t, body, "Username and password are required")
	assert.Contains(t, body, "type=\"email\"")
	assert.Contains(t, body, "type=\"password\"")
	assert.Contains(t, body, "placeholder=\"your@email.com\"")
	assert.Contains(t, body, "placeholder=\"password\"")
}

func TestLoginFormStatusMessages(t *testing.T) {
	config := LoginConfig{
		Title:   "Status Test",
		SpecURL: "/status/swagger.json",
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Test for status message handling
	assert.Contains(t, body, "function showStatus(message, type)")
	assert.Contains(t, body, "showStatus('Logging in...', 'info')")
	assert.Contains(t, body, "Login successful!")
	assert.Contains(t, body, "Login failed:")
	assert.Contains(t, body, "login-status")
}

func TestLoginFormConcurrency(t *testing.T) {
	config := LoginConfig{
		Title:   "Concurrency Test",
		SpecURL: "/concurrency/swagger.json",
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	// Test concurrent requests to ensure no race conditions
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "Concurrency Test")
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestLoginFormErrorHandling(t *testing.T) {
	// Test that the function handles errors gracefully
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Test with valid config
	config := LoginConfig{
		Title:   "Error Test",
		SpecURL: "/error/test",
	}
	
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should not panic and should return valid response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
	
	body := w.Body.String()
	// Should contain error handling code
	assert.Contains(t, body, "try {")
	assert.Contains(t, body, "} catch (error) {")
	assert.Contains(t, body, "console.error")
}

func TestLoginFormResponseSize(t *testing.T) {
	config := LoginConfig{
		Title:   "Size Test API Documentation with Long Title",
		SpecURL: "/very/long/path/to/swagger/documentation.json",
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerWithLogin(config)
	router.GET("/test", handler)
	
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should return a substantial HTML document
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, w.Body.Len(), 1000)
	
	body := w.Body.String()
	assert.Contains(t, body, config.Title)
	assert.Contains(t, body, config.SpecURL)
}