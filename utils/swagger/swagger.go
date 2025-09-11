package swagger

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SwaggerConfig struct {
	Title         string
	SwaggerDocURL string
	AuthURL       string
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.css" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.15.5/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.15.5/favicon-16x16.png" sizes="16x16" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            background: #fafafa;
        }
        
        /* Ensure Swagger UI loads properly */
        #swagger-ui {
            max-width: none !important;
        }
        
        /* Custom login form styling */
        .login-form-section {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 5px;
            padding: 20px;
            margin-bottom: 20px;
        }
        
        .login-form-title {
            margin: 0 0 10px 0;
            font-size: 16px;
            font-weight: bold;
            color: #3b4151;
        }
        
        .login-form-subtitle {
            font-size: 12px;
            color: #888;
            margin-bottom: 15px;
        }
        
        .login-form-group {
            margin-bottom: 15px;
        }
        
        .login-form-label {
            display: block;
            font-size: 12px;
            font-weight: bold;
            color: #3b4151;
            margin-bottom: 5px;
        }
        
        .login-form-input {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid #d9d9d9;
            border-radius: 4px;
            font-size: 14px;
            box-sizing: border-box;
        }
        
        .login-form-button {
            background: #4990e2;
            color: white;
            border: none;
            border-radius: 4px;
            padding: 10px 20px;
            font-size: 14px;
            font-weight: bold;
            cursor: pointer;
        }
        
        .login-form-button:hover {
            background: #357abd;
        }
        
        .login-form-button:disabled {
            background: #6c757d;
            cursor: not-allowed;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>

    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js" charset="UTF-8"> </script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
    <script>
        // Set authentication URL globally
        window.AUTH_URL = "{{.AuthURL}}";

        window.onload = function() {
            // Initialize Swagger UI with stable configuration
            const ui = SwaggerUIBundle({
                url: '{{.SwaggerDocURL}}',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                docExpansion: "list",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                validatorUrl: null,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log("Swagger UI loaded successfully");
                    // Start monitoring for auth dialogs after UI is loaded
                    setTimeout(startAuthDialogMonitoring, 500);
                }
            });

            // Function to add login form to authorization dialog
            function startAuthDialogMonitoring() {
                const checkForAuthDialog = () => {
                    const authWrappers = document.querySelectorAll('.auth-wrapper:not(.login-enhanced)');
                    
                    authWrappers.forEach(wrapper => {
                        wrapper.classList.add('login-enhanced');
                        
                        // Find BearerAuth section
                        const bearerAuth = Array.from(wrapper.querySelectorAll('.auth-container')).find(
                            container => container.textContent && container.textContent.includes('BearerAuth')
                        );
                        
                        if (bearerAuth) {
                            // Create login form
                            const loginForm = document.createElement('div');
                            loginForm.className = 'login-form-section';
                            loginForm.innerHTML = ` + "`" + `
                                <div class="login-form-title">Login</div>
                                <div class="login-form-subtitle">Returns a token for using in BearerAuth</div>
                                
                                <div class="login-form-group">
                                    <label class="login-form-label">Username</label>
                                    <input type="text" id="login-username" class="login-form-input" placeholder="Enter username" />
                                </div>
                                
                                <div class="login-form-group">
                                    <label class="login-form-label">Password</label>
                                    <input type="password" id="login-password" class="login-form-input" placeholder="Enter password" />
                                </div>
                                
                                <button class="login-form-button" onclick="performAuthentication()">Login</button>
                            ` + "`" + `;
                            
                            // Insert before BearerAuth section
                            bearerAuth.parentNode.insertBefore(loginForm, bearerAuth);
                        }
                    });
                };

                // Monitor for new auth dialogs
                const observer = new MutationObserver(checkForAuthDialog);
                observer.observe(document.body, { childList: true, subtree: true });
                
                // Check immediately and periodically
                checkForAuthDialog();
                setInterval(checkForAuthDialog, 1000);
            }

            // Global authentication function
            window.performAuthentication = async function() {
                const username = document.getElementById('login-username')?.value?.trim();
                const password = document.getElementById('login-password')?.value?.trim();
                const button = document.querySelector('.login-form-button');

                if (!username || !password) {
                    alert('Please enter both username and password');
                    return;
                }

                // Disable button during request
                button.disabled = true;
                button.textContent = 'Logging in...';

                try {
                    const response = await fetch(window.AUTH_URL, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            email: username,
                            password: password
                        })
                    });

                    if (!response.ok) {
                        const errorData = await response.json();
                        throw new Error(errorData.message || 'Authentication failed');
                    }

                    const data = await response.json();
                    const accessToken = data.data?.access_token;

                    if (!accessToken) {
                        throw new Error('No access token received');
                    }

                    // Find and populate the Bearer token input field
                    const tokenInput = document.querySelector('input[data-name="Authorization"]');
                    if (tokenInput) {
                        tokenInput.value = 'Bearer ' + accessToken;
                        
                        // Trigger change events to notify Swagger UI
                        tokenInput.dispatchEvent(new Event('input', { bubbles: true }));
                        tokenInput.dispatchEvent(new Event('change', { bubbles: true }));
                        
                        alert('✅ Authentication successful! Bearer token has been automatically filled.');
                    } else {
                        // Fallback: show token to user
                        alert('✅ Authentication successful!\\n\\nToken: Bearer ' + accessToken + '\\n\\nPlease paste this token in the Authorization field.');
                    }

                } catch (error) {
                    console.error('Authentication error:', error);
                    alert('❌ Authentication failed: ' + error.message);
                } finally {
                    // Re-enable button
                    button.disabled = false;
                    button.textContent = 'Login';
                }
            };

            // Start monitoring for auth dialogs
            injectAuthForm();
        };
    </script>
</body>
</html>`

// ServeSwaggerUI serves the Swagger UI with authentication form
func ServeSwaggerUI(config SwaggerConfig) gin.HandlerFunc {
	// Set defaults
	if config.Title == "" {
		config.Title = "API Documentation"
	}
	if config.SwaggerDocURL == "" {
		config.SwaggerDocURL = "/swagger/doc.json"
	}
	if config.AuthURL == "" {
		config.AuthURL = "/api/v1/auth/user/login"
	}

	tmpl := template.Must(template.New("swagger").Parse(swaggerHTML))

	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(c.Writer, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render Swagger UI"})
		}
	}
}
