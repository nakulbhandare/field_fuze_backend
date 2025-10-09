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

// SwaggerTestSuite defines a test suite for swagger functions
type SwaggerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupTest runs before each test
func (suite *SwaggerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
}

// TestServeSwaggerUI tests the ServeSwaggerUI function
func (suite *SwaggerTestSuite) TestServeSwaggerUI() {
	config := SwaggerConfig{
		Title:         "Test API",
		SwaggerDocURL: "/swagger/doc.json",
		AuthURL:       "/api/v1/auth/user/login",
	}

	handler := ServeSwaggerUI(config)
	suite.router.GET("/swagger", handler)

	req, err := http.NewRequest("GET", "/swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(suite.T(), body, "Test API")
	assert.Contains(suite.T(), body, "/swagger/doc.json")
	assert.Contains(suite.T(), body, "/api/v1/auth/user/login")
	assert.Contains(suite.T(), body, "swagger-ui-bundle.js")
	assert.Contains(suite.T(), body, "performAuthentication")
}

// TestServeSwaggerUIWithDefaults tests ServeSwaggerUI with default values
func (suite *SwaggerTestSuite) TestServeSwaggerUIWithDefaults() {
	config := SwaggerConfig{} // All empty values

	handler := ServeSwaggerUI(config)
	suite.router.GET("/swagger", handler)

	req, err := http.NewRequest("GET", "/swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	body := w.Body.String()
	// Check default values
	assert.Contains(suite.T(), body, "API Documentation")
	assert.Contains(suite.T(), body, "/swagger/doc.json")
	assert.Contains(suite.T(), body, "/api/v1/auth/user/login")
}

// TestServeSwaggerUIWithPartialConfig tests ServeSwaggerUI with partial configuration
func (suite *SwaggerTestSuite) TestServeSwaggerUIWithPartialConfig() {
	config := SwaggerConfig{
		Title: "Custom Title",
		// SwaggerDocURL and AuthURL left empty to test defaults
	}

	handler := ServeSwaggerUI(config)
	suite.router.GET("/swagger", handler)

	req, err := http.NewRequest("GET", "/swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(suite.T(), body, "Custom Title")
	assert.Contains(suite.T(), body, "/swagger/doc.json")       // default
	assert.Contains(suite.T(), body, "/api/v1/auth/user/login") // default
}

// TestServeSwaggerUIWithCustomConfig tests ServeSwaggerUI with all custom values
func (suite *SwaggerTestSuite) TestServeSwaggerUIWithCustomConfig() {
	config := SwaggerConfig{
		Title:         "My Custom API Documentation",
		SwaggerDocURL: "/custom/swagger.json",
		AuthURL:       "/custom/auth/login",
	}

	handler := ServeSwaggerUI(config)
	suite.router.GET("/swagger", handler)

	req, err := http.NewRequest("GET", "/swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(suite.T(), body, "My Custom API Documentation")
	assert.Contains(suite.T(), body, "/custom/swagger.json")
	assert.Contains(suite.T(), body, "/custom/auth/login")
}

// TestServeSwaggerUIHTMLStructure tests the HTML structure
func (suite *SwaggerTestSuite) TestServeSwaggerUIHTMLStructure() {
	config := SwaggerConfig{
		Title:         "Test API",
		SwaggerDocURL: "/swagger/doc.json",
		AuthURL:       "/api/v1/auth/user/login",
	}

	handler := ServeSwaggerUI(config)
	suite.router.GET("/swagger", handler)

	req, err := http.NewRequest("GET", "/swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	body := w.Body.String()

	// Test HTML structure
	assert.Contains(suite.T(), body, "<!DOCTYPE html>")
	assert.Contains(suite.T(), body, "<html lang=\"en\">")
	assert.Contains(suite.T(), body, "<head>")
	assert.Contains(suite.T(), body, "<body>")
	assert.Contains(suite.T(), body, "<div id=\"swagger-ui\">")
	assert.Contains(suite.T(), body, "</html>")

	// Test CSS and JS includes
	assert.Contains(suite.T(), body, "swagger-ui-bundle.css")
	assert.Contains(suite.T(), body, "swagger-ui-bundle.js")
	assert.Contains(suite.T(), body, "swagger-ui-standalone-preset.js")

	// Test JavaScript functions
	assert.Contains(suite.T(), body, "window.AUTH_URL")
	assert.Contains(suite.T(), body, "SwaggerUIBundle")
	assert.Contains(suite.T(), body, "performAuthentication")
}

// TestServeCleanSwagger tests the ServeCleanSwagger function
func (suite *SwaggerTestSuite) TestServeCleanSwagger() {
	config := SwaggerConfig{
		Title:         "Clean API",
		SwaggerDocURL: "/swagger/doc.json",
		AuthURL:       "/api/v1/auth/user/login",
	}

	handler := ServeCleanSwagger(config)
	suite.router.GET("/clean-swagger", handler)

	req, err := http.NewRequest("GET", "/clean-swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(suite.T(), body, "Clean API")
	assert.Contains(suite.T(), body, "/swagger/doc.json")
	assert.Contains(suite.T(), body, "/api/v1/auth/user/login")
	assert.Contains(suite.T(), body, "swagger-ui-bundle.js")
	assert.Contains(suite.T(), body, "attachAuthorizeButtonListener")
}

// TestServeCleanSwaggerWithDefaults tests ServeCleanSwagger with default values
func (suite *SwaggerTestSuite) TestServeCleanSwaggerWithDefaults() {
	config := SwaggerConfig{} // All empty values

	handler := ServeCleanSwagger(config)
	suite.router.GET("/clean-swagger", handler)

	req, err := http.NewRequest("GET", "/clean-swagger", nil)
	require.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	body := w.Body.String()
	// Check default values
	assert.Contains(suite.T(), body, "API Documentation")
	assert.Contains(suite.T(), body, "/swagger/doc.json")
	assert.Contains(suite.T(), body, "/api/v1/auth/user/login")
}

// TestTemplateExecution tests template execution directly
func (suite *SwaggerTestSuite) TestTemplateExecution() {
	// Test that the template can be parsed and executed without errors
	tmpl, err := template.New("test").Parse(swaggerHTML)
	require.NoError(suite.T(), err)

	config := SwaggerConfig{
		Title:         "Test API",
		SwaggerDocURL: "/test/swagger.json",
		AuthURL:       "/test/auth",
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, config)
	require.NoError(suite.T(), err)

	result := buf.String()
	assert.Contains(suite.T(), result, "Test API")
	assert.Contains(suite.T(), result, "/test/swagger.json")
	assert.Contains(suite.T(), result, "/test/auth")
}

// TestCleanTemplateExecution tests clean template execution directly
func (suite *SwaggerTestSuite) TestCleanTemplateExecution() {
	// Test that the clean template can be parsed and executed without errors
	tmpl, err := template.New("test").Parse(cleanSwaggerHTML)
	require.NoError(suite.T(), err)

	config := SwaggerConfig{
		Title:         "Clean Test API",
		SwaggerDocURL: "/clean/swagger.json",
		AuthURL:       "/clean/auth",
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, config)
	require.NoError(suite.T(), err)

	result := buf.String()
	assert.Contains(suite.T(), result, "Clean Test API")
	assert.Contains(suite.T(), result, "/clean/swagger.json")
	assert.Contains(suite.T(), result, "/clean/auth")
}

// Run the test suite
func TestSwaggerTestSuite(t *testing.T) {
	suite.Run(t, new(SwaggerTestSuite))
}

// Standalone tests

func TestSwaggerConfigStruct(t *testing.T) {
	// Test SwaggerConfig struct creation and field assignment
	config := SwaggerConfig{
		Title:         "Test Title",
		SwaggerDocURL: "Test URL",
		AuthURL:       "Test Auth URL",
	}

	assert.Equal(t, "Test Title", config.Title)
	assert.Equal(t, "Test URL", config.SwaggerDocURL)
	assert.Equal(t, "Test Auth URL", config.AuthURL)
}

func TestSwaggerConfigDefaults(t *testing.T) {
	testCases := []struct {
		name           string
		inputConfig    SwaggerConfig
		expectedTitle  string
		expectedDocURL string
		expectedAuth   string
	}{
		{
			name:           "Empty config uses all defaults",
			inputConfig:    SwaggerConfig{},
			expectedTitle:  "API Documentation",
			expectedDocURL: "/swagger/doc.json",
			expectedAuth:   "/api/v1/auth/user/login",
		},
		{
			name: "Partial config uses some defaults",
			inputConfig: SwaggerConfig{
				Title: "Custom Title",
			},
			expectedTitle:  "Custom Title",
			expectedDocURL: "/swagger/doc.json",
			expectedAuth:   "/api/v1/auth/user/login",
		},
		{
			name: "Full config uses no defaults",
			inputConfig: SwaggerConfig{
				Title:         "Custom Title",
				SwaggerDocURL: "/custom/doc.json",
				AuthURL:       "/custom/auth",
			},
			expectedTitle:  "Custom Title",
			expectedDocURL: "/custom/doc.json",
			expectedAuth:   "/custom/auth",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test ServeSwaggerUI
			gin.SetMode(gin.TestMode)
			router := gin.New()
			handler := ServeSwaggerUI(tc.inputConfig)
			router.GET("/test", handler)

			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			body := w.Body.String()
			assert.Contains(t, body, tc.expectedTitle)
			assert.Contains(t, body, tc.expectedDocURL)
			assert.Contains(t, body, tc.expectedAuth)
		})
	}
}

func TestSwaggerUISpecialCharacters(t *testing.T) {
	// Test that special characters in config are properly escaped
	config := SwaggerConfig{
		Title:         "API & Documentation <test>",
		SwaggerDocURL: "/swagger/doc.json?version=1.0&format=json",
		AuthURL:       "/api/v1/auth/user/login?redirect=/home",
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerUI(config)
	router.GET("/test", handler)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// The template should properly escape these characters
	assert.Contains(t, body, "API &amp; Documentation &lt;test&gt;")
	// URLs should be preserved as-is in JavaScript
	assert.Contains(t, body, "/swagger/doc.json?version=1.0&format=json")
	assert.Contains(t, body, "/api/v1/auth/user/login?redirect=/home")
}

func TestSwaggerUIJavaScriptContent(t *testing.T) {
	config := SwaggerConfig{
		Title:         "Test API",
		SwaggerDocURL: "/test/swagger.json",
		AuthURL:       "/test/auth",
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeSwaggerUI(config)
	router.GET("/test", handler)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Test JavaScript functions and variables
	assert.Contains(t, body, "window.AUTH_URL = \"/test/auth\"")
	assert.Contains(t, body, "url: '/test/swagger.json'")
	assert.Contains(t, body, "SwaggerUIBundle")
	assert.Contains(t, body, "performAuthentication")
	assert.Contains(t, body, "startAuthDialogMonitoring")

	// Test CSS classes
	assert.Contains(t, body, "login-form-section")
	assert.Contains(t, body, "login-form-button")
	assert.Contains(t, body, "auth-wrapper")
}

func TestCleanSwaggerUIContent(t *testing.T) {
	config := SwaggerConfig{
		Title:         "Clean Test API",
		SwaggerDocURL: "/clean/swagger.json",
		AuthURL:       "/clean/auth",
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := ServeCleanSwagger(config)
	router.GET("/test", handler)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body := w.Body.String()

	// Test JavaScript functions specific to clean version
	assert.Contains(t, body, "window.AUTH_URL = '\\/clean\\/auth'") // Escaped slashes
	assert.Contains(t, body, "url: '\\/clean\\/swagger.json'") // Escaped slashes
	assert.Contains(t, body, "attachAuthorizeButtonListener")
	assert.Contains(t, body, "addLoginForm")
	assert.Contains(t, body, "createAuthContainer")
	assert.Contains(t, body, "attachLoginFunctionality")

	// Test CSS classes specific to clean version
	assert.Contains(t, body, "custom-auth-container")
	assert.Contains(t, body, "col_header")
	assert.Contains(t, body, "auth-btn-wrapper")
}

func TestErrorHandling(t *testing.T) {
	// Test error handling when template execution fails
	// This is hard to test directly since the templates are well-formed,
	// but we can test that the function returns without panicking

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test with valid config
	config := SwaggerConfig{
		Title:         "Test",
		SwaggerDocURL: "/test",
		AuthURL:       "/auth",
	}

	handler := ServeSwaggerUI(config)
	router.GET("/test", handler)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not panic and should return valid response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestHTTPHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := SwaggerConfig{
		Title:         "Header Test",
		SwaggerDocURL: "/test",
		AuthURL:       "/auth",
	}

	handler := ServeSwaggerUI(config)
	router.GET("/test", handler)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Test content type header
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	// Test that response is not empty
	assert.NotEmpty(t, w.Body.String())
	assert.Greater(t, w.Body.Len(), 1000) // Should be a substantial HTML document
}

func TestConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := SwaggerConfig{
		Title:         "Concurrent Test",
		SwaggerDocURL: "/test",
		AuthURL:       "/auth",
	}

	handler := ServeSwaggerUI(config)
	router.GET("/test", handler)

	// Test concurrent requests to ensure no race conditions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "Concurrent Test")
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
