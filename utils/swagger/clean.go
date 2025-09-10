package swagger

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

const cleanSwaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{{.Title}}</title>
  <link rel="stylesheet" type="text/css" href="https://petstore.swagger.io/swagger-ui.css" />
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
      margin:0;
      background: #fafafa;
    }
    
    /* Custom login form styling to match Swagger UI */
    .custom-auth-container {
      margin-bottom: 20px;
    }
    
    .custom-auth-container h4 {
      color: #3b4151;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      font-size: 14px;
      font-weight: 600;
      margin: 0 0 5px;
      padding: 0;
    }
    
    .custom-auth-container .wrapper {
      padding: 0;
      margin: 0 0 15px;
    }
    
    .custom-auth-container .description {
      color: #3b4151;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      font-size: 12px;
      margin: 0 0 15px;
      padding: 0;
    }
    
    .custom-auth-container .col_header {
      margin-bottom: 5px;
    }
    
    .custom-auth-container .col_header label {
      color: #3b4151;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      font-size: 12px;
      font-weight: 600;
      margin: 0;
      padding: 0;
    }
    
    .custom-auth-container .auth-btn-wrapper {
      margin-bottom: 15px;
      padding: 0;
    }
    
    .custom-auth-container input {
      background: #ffffff !important;
      border: 2px solid #3b82f6 !important;
      border-radius: 4px !important;
      box-sizing: border-box !important;
      color: #3b4151 !important;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif !important;
      font-size: 14px !important;
      outline: none !important;
      padding: 8px 12px !important;
      width: 450px !important;
      max-width: 450px !important;
      height: 40px !important;
      max-height: 40px !important;
      line-height: 1.4 !important;
      transition: all 0.3s ease !important;
    }
    
    .custom-auth-container input:focus {
      border-color: #3b82f6 !important;
      box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1) !important;
      background-color: #fafbfc !important;
    }
    
    .custom-auth-container input:hover {
      border-color: #2563eb !important;
    }
    
    .custom-auth-container .btn.authorize {
      background: #4990e2;
      border: 1px solid #4990e2;
      border-radius: 4px;
      box-shadow: 0 1px 2px rgba(0,0,0,.1);
      color: #ffffff;
      cursor: pointer;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      font-size: 14px;
      font-weight: 600;
      outline: none;
      padding: 8px 16px;
      text-decoration: none;
      transition: all 0.3s;
      user-select: none;
    }
    
    .custom-auth-container .btn.authorize:hover {
      background: #357abd;
      border-color: #357abd;
    }
    
    .custom-auth-container .btn.authorize:disabled {
      background: #6c757d;
      border-color: #6c757d;
      cursor: not-allowed;
      opacity: 0.65;
    }
    
    .custom-auth-container .auth-separator {
      background: #ebebeb;
      border: none;
      height: 1px;
      margin: 20px 0;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://petstore.swagger.io/swagger-ui-bundle.js" crossorigin></script>
  <script src="https://petstore.swagger.io/swagger-ui-standalone-preset.js" crossorigin></script>
  <script>
    // Set authentication URL globally for login functionality
    window.AUTH_URL = '{{.AuthURL}}';

    window.onload = () => {
      window.ui = SwaggerUIBundle({
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
        onComplete: function() {
          console.log("Swagger UI loaded successfully");
          // Start monitoring for authorization dialogs
          attachAuthorizeButtonListener();
        }
      });
    };

    const attachAuthorizeButtonListener = () => {
      document.body.addEventListener("click", (event) => {
        if (event.target.closest(".authorize")) {
          setTimeout(addLoginForm, 500);
        }
      });
    };

    const addLoginForm = () => {
      const modalContent = document.querySelector(".modal-ux .modal-ux-content .auth-container");

      if (!modalContent) {
        console.error("Swagger Authorize modal not found!");
        return;
      }

      if (!document.querySelector(".custom-auth-container")) {
        const authContainer = createAuthContainer();
        modalContent.prepend(authContainer);
        console.log("Custom login form successfully added!");
      }
    };

    const createAuthContainer = () => {
      const authContainer = document.createElement("div");
      authContainer.className = "custom-auth-container";
      authContainer.style.marginBottom = "20px";

      authContainer.innerHTML = ` + "`" + `
        <h4>Login</h4>
        <div class="wrapper">
          <p class="description">Returns a <code>token</code> for using in <code>BearerAuth</code></p>
          
          <div class="col_header">
            <label>Username:</label>
          </div>
          <div class="auth-btn-wrapper">
            <input id="swagger-username" type="text" placeholder="Username" />
          </div>
          
          <div class="col_header">
            <label>Password:</label>
          </div>
          <div class="auth-btn-wrapper">
            <input id="swagger-password" type="password" placeholder="Password" />
          </div>
          
          <div class="auth-btn-wrapper">
            <button id="swagger-login" class="btn authorize unlocked">
              <span>Login</span>
            </button>
          </div>
        </div>
        <hr class="auth-separator">
      ` + "`" + `;

      attachLoginFunctionality(authContainer);
      return authContainer;
    };

    const attachLoginFunctionality = (container) => {
      container.querySelector("#swagger-login").onclick = async function () {
        const username = document.getElementById("swagger-username").value;
        const password = document.getElementById("swagger-password").value;
        const loginBtn = this;

        if (!username || !password) {
          alert("Username and password are required.");
          return;
        }

        // Show loading state
        const originalText = loginBtn.textContent;
        loginBtn.disabled = true;
        loginBtn.innerHTML = '<span>Logging in...</span>';

        try {
          const authUrl = window.AUTH_URL || "/api/v1/auth/login";
          const response = await fetch(authUrl, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: username, password: password }),
          });

          const data = await response.json();

          if (response.ok) {
            const token = "Bearer " + (data.data?.access_token || data.access_token);

            if (window.ui) {
              window.ui.preauthorizeApiKey("BearerAuth", token);
              alert("✅ Login successful! You are now authorized to use all APIs.");
              
              // Reset form
              document.getElementById("swagger-username").value = '';
              document.getElementById("swagger-password").value = '';
            } else {
              console.error("Swagger UI instance not found during login.");
              alert("Login successful but couldn't auto-authorize. Please copy this token: " + token);
            }
          } else {
            alert("❌ Login failed: " + (data.message || data.error || "Unknown error"));
          }
        } catch (err) {
          console.error("Login error:", err);
          alert("❌ An error occurred during login: " + err.message);
        } finally {
          // Reset button state
          loginBtn.disabled = false;
          loginBtn.innerHTML = '<span>Login</span>';
        }
      };
    };
  </script>
</body>
</html>`

func ServeCleanSwagger(config SwaggerConfig) gin.HandlerFunc {
	if config.Title == "" {
		config.Title = "API Documentation"
	}
	if config.SwaggerDocURL == "" {
		config.SwaggerDocURL = "/swagger/doc.json"
	}
	if config.AuthURL == "" {
		config.AuthURL = "/api/v1/auth/login"
	}

	tmpl := template.Must(template.New("swagger").Parse(cleanSwaggerHTML))

	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(c.Writer, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render Swagger UI"})
		}
	}
}