package swagger

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerWithLoginTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.css" />
    <style>
        /* Custom login form styles */
        .login-section {
            margin-bottom: 20px;
            padding: 15px;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 5px;
        }
        .login-form {
            display: flex;
            gap: 10px;
            align-items: end;
            flex-wrap: wrap;
        }
        .login-field {
            display: flex;
            flex-direction: column;
        }
        .login-field label {
            font-size: 12px;
            font-weight: bold;
            margin-bottom: 3px;
        }
        .login-field input {
            padding: 8px;
            border: 1px solid #ccc;
            border-radius: 3px;
            width: 150px;
        }
        .login-btn {
            padding: 8px 16px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 3px;
            cursor: pointer;
        }
        .login-btn:hover {
            background: #0056b3;
        }
        .login-status {
            margin-top: 10px;
            padding: 8px;
            border-radius: 3px;
            display: none;
        }
        .login-status.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .login-status.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <!-- Custom login form -->
    <div class="login-section">
        <h3>Quick Login</h3>
        <div class="login-form">
            <div class="login-field">
                <label>Email:</label>
                <input type="email" id="login-email" placeholder="your@email.com" />
            </div>
            <div class="login-field">
                <label>Password:</label>
                <input type="password" id="login-password" placeholder="password" />
            </div>
            <button class="login-btn" onclick="performLogin()">Login & Authorize</button>
        </div>
        <div id="login-status" class="login-status"></div>
    </div>

    <div id="swagger-ui"></div>
    
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        let swaggerUI;

        window.onload = function() {
            swaggerUI = SwaggerUIBundle({
                url: '{{.SpecURL}}',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.presets.standalone
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };

        async function performLogin() {
            const email = document.getElementById('login-email').value;
            const password = document.getElementById('login-password').value;
            const statusDiv = document.getElementById('login-status');

            if (!email || !password) {
                showStatus('Please enter both email and password', 'error');
                return;
            }

            try {
                showStatus('Logging in...', 'info');

                const response = await fetch('/api/v1/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        email: email,
                        password: password
                    })
                });

                if (!response.ok) {
                    const errorData = await response.json();
                    throw new Error(errorData.message || 'Login failed');
                }

                const data = await response.json();
                const token = data.data?.access_token;

                if (!token) {
                    throw new Error('No access token received');
                }

                // Authorize Swagger UI with the token
                swaggerUI.authActions.authorize({
                    BearerAuth: {
                        name: 'BearerAuth',
                        schema: {
                            type: 'apiKey',
                            in: 'header',
                            name: 'Authorization'
                        },
                        value: 'Bearer ' + token
                    }
                });

                showStatus('✓ Login successful! You are now authorized to use all APIs.', 'success');

            } catch (error) {
                console.error('Login failed:', error);
                showStatus('✗ Login failed: ' + error.message, 'error');
            }
        }

        function showStatus(message, type) {
            const statusDiv = document.getElementById('login-status');
            statusDiv.textContent = message;
            statusDiv.className = 'login-status ' + type;
            statusDiv.style.display = 'block';

            if (type === 'success') {
                setTimeout(() => {
                    statusDiv.style.display = 'none';
                }, 5000);
            }
        }

        </script>
</body>
</html>`

type LoginConfig struct {
	Title   string
	SpecURL string
}

func ServeSwaggerWithLogin(config LoginConfig) gin.HandlerFunc {
	tmpl := template.Must(template.New("swagger-login").Parse(swaggerWithLoginTemplate))

	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")

		if err := tmpl.Execute(c.Writer, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render Swagger UI"})
			return
		}
	}
}
