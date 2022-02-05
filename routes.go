package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type TokenClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type LoginStruct struct {
	Username string `json:"username"`
}

// Upgrader do websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:3000" {
			return true
		} else {
			return false
		}
	},
}

// Endpoint do websocket
func websocketHandler(h *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Obtendo username do JWT
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*TokenClaims)
		username := claims.Username

		// Fazendo upgrade na conexão
		connection, err := upgrader.Upgrade(c.Response(), c.Request(), nil)

		if err != nil {
			log.Println("Erro ao fazer upgrade na conexão:", err)
			return echo.ErrInternalServerError
		}

		// Criando um cliente
		client := NewClient(username, h, connection)

		// Adicionando cliente no hub
		h.AddClient <- client

		// Iniciando goroutines do cliente
		go client.writePump()
		go client.readPump()

		return nil
	}
}

// Endpoint de autenticação, cria um cookie com um JWT
func signUp(h *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Lendo body
		body, err := ioutil.ReadAll(c.Request().Body)

		if err != nil {
			log.Println("Erro ao ler o body:", err)
			return echo.ErrInternalServerError
		}

		var login LoginStruct

		// Unmarshal do body na estrutura LoginStruct
		json.Unmarshal(body, &login)

		// Verificar se o nome possui um tamanho aceitável de caracteres
		if len(login.Username) < 4 {
			return echo.ErrBadRequest
		}

		// Verificar se já existe um usuário com esse nome
		if h.Clients[login.Username] != nil {
			return echo.ErrBadRequest
		}

		// exp do JWT e cookie
		expirationTime := time.Now().Add(time.Hour * 72)

		// Criando uma estrutura com as informações necessárias
		claims := &TokenClaims{
			Username: login.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		// Criando um JWT com a estrutura claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Usando uma key super secreta para criptografar o JWT
		signedToken, err := token.SignedString([]byte("segredo"))

		if err != nil {
			log.Println("Erro ao criar um token:", err)
			return echo.ErrInternalServerError
		}

		// Criando um header de set-cookie
		http.SetCookie(c.Response(), &http.Cookie{
			Name:     "token",
			Value:    signedToken,
			Expires:  expirationTime,
			Path:     "/",
			Domain:   "localhost",
			SameSite: http.SameSiteLaxMode,
		})

		// Enviando o token como resposta
		return c.JSON(http.StatusOK, map[string]interface{}{
			"token": signedToken,
		})
	}
}
