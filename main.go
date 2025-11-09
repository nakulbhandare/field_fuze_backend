package main

import (
	"context"
	"fieldfuze-backend/controller"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils"
	"fieldfuze-backend/utils/logger"
	"fieldfuze-backend/worker"
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

var config *models.Config

func Init() {
	fmt.Println("Inside Init :: ")
	var err error
	config, err = utils.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

// @title FieldFuze Backend API
// @version 1.0
// @description FieldFuze Backend API with DynamoDB and Telnyx Integration
// @description
// @description ## 🔥 AUTHENTICATION FLOW:
// @description ### Step 1: Register Customer
// @description **POST /auth/register** - Create customer account (no token generated)
// @description `{"email": "user@example.com", "username": "john", "password": "pass123", "first_name": "John", "last_name": "Doe"}`
// @description
// @description ## 🚀 QUICK AUTHENTICATION:
// @description ### Using the Authorize Button (Recommended)
// @description 1. Click the **🔓 Authorize** button (top right of any API section)
// @description 2. In the authorization dialog, use the **Login** form:
// @description    - Enter your **Username** (email)
// @description    - Enter your **Password**
// @description    - Click **Login** button
// @description 3. Your Bearer token will be **automatically applied** to all API calls!
// @description 4. All protected endpoints will now work without manual token entry
// @description
// @description ### Manual Token Entry (Alternative)
// @description If you prefer manual setup: Get token from `/auth/login`, then paste `Bearer YOUR_TOKEN` in the Authorization field
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host
// @BasePath /api/v1/auth

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Enter 'Bearer' [space] and then your token in the text input below.
func main() {
	Init()
	fmt.Println("Hello, World!")
	fmt.Println("Config Loaded ::", dal.PrintPrettyJSON(config))

	ctx := context.Background()

	r := gin.New()

	// Add Gin's default logger middleware for colored API call logs
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	c := controller.NewController(context.Background(), config, logger.NewLogger(config.LogLevel, config.LogFormat))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.RegisterRoutes(context.Background(), config, r, config.BasePath) // should call r.Run()
	}()

	// 🚀 START INFRASTRUCTURE WORKER (CRON JOB)
	infraWorker, err := worker.NewService(ctx, config, logger.NewLogger(config.LogLevel, config.LogFormat))
	if err != nil {
		log.Fatalf("Failed to create infrastructure worker: %v", err)
	}

	// Start infrastructure worker in background
	if err := infraWorker.StartInBackground(); err != nil {
		log.Fatalf("Failed to start infrastructure worker: %v", err)
	}

	wg.Wait()
}
