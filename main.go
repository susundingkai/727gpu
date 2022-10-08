package main

import (
	"727gpu_server/config"
	"727gpu_server/src"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const (
	dbDriverName = "sqlite3"
	dbName       = "./database/sql.db"
)

var (
	db *sql.DB
)

func main() {
	stuck := make(chan os.Signal)
	db = GetDB()
	config := config.ReadConfig()
	ServeHTTP(config)

	<-stuck
}

func ServeHTTP(config config.MyConfig) {
	go func() {
		g := gin.New()
		g.Use(gin.Recovery())
		err := g.SetTrustedProxies(nil)
		if err != nil {
			panic(err)
		}

		public := g.Group("/node")
		public.GET(
			"",
			SocketHandler,
		)

		if err := g.Run(fmt.Sprintf(":%d", config.Server.Port)); err != nil {
			panic(err)
		}
	}()
}
func SocketHandler(c *gin.Context) {
	src.SocketHandler(c, db)
}
func GetDB() *sql.DB {
	db, err := sql.Open(dbDriverName, dbName)
	if err != nil {
		panic(err)
	}
	return db
}
