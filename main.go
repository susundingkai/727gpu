package main

import (
	"727gpu_server/config"
	"727gpu_server/src"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/unrolled/secure"
	"os"
)

const (
	dbDriverName = "sqlite3"
	dbName       = "./database/sql.db"
)

var (
	db *sql.DB
)
var myConfig = config.ReadConfig()

func main() {

	stuck := make(chan os.Signal)
	db = GetDB()

	ServeHTTP(myConfig)

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
		public = g.Group("/portal")
		public.GET(
			"",
			ProtalHandler,
		)

		g.Use(TlsHandler())
		// 开启端口监听
		g.RunTLS(fmt.Sprintf(":%d", myConfig.Server.Port), "./cert/pris.ssdk.icu.pem", "./cert/pris.ssdk.icu.key")
	}()
}

func TlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     fmt.Sprintf("localhost:%d", myConfig.Server.Port),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}
func SocketHandler(c *gin.Context) {
	src.NodeHandler(c, db)
}
func ProtalHandler(c *gin.Context) {
	src.PortalHandler(c, db)
}
func GetDB() *sql.DB {
	db, err := sql.Open(dbDriverName, dbName)
	if err != nil {
		panic(err)
	}
	return db
}
