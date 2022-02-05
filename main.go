package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

func main() {
	// Iniciando conexão com redis
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	defer rdb.Close()

	// Testando conexão
	err := rdb.Ping(ctx).Err()

	if err != nil {
		panic(err)
	}

	// Flush na database
	_ = rdb.FlushDB(ctx).Err()

	// Iniciando Echo
	e := echo.New()

	// Configurando Logger
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} [${method}, ${status}] ${uri}\n",
		CustomTimeFormat: "2006/01/02 15:04:05",
	}))

	// Configurando CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"http://localhost:3000"},
	}))

	// Criando hub e inicializando a sua goroutine
	hub := NewHub()
	go hub.run()

	// Endpoint de autenticação
	e.POST("/api/signup", signUp(hub))

	// Grupo de endpoints com middleware de autenticação
	api := e.Group("/api")

	config := middleware.JWTConfig{
		Claims:      &TokenClaims{},
		SigningKey:  []byte("segredo"),
		TokenLookup: "cookie:token",
	}

	api.Use(middleware.JWTWithConfig(config))

	// Websocket
	api.GET("/ws", websocketHandler(hub))

	e.Logger.Fatal(e.Start(":2000"))
}
