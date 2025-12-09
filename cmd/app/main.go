// @title User Service API
// @version 1.0
// @description Public + Private API for auth, users, and preferences.

// @host localhost:8000
// @BasePath /api/v2

// @securityDefinitions.apikey SessionCookie
// @in cookie
// @name session

package main

import "github.com/KimNattanan/go-user-service/internal/app"

func main() {
	app.Start()
}
