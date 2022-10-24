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
		public = g.Group("/portal")
		public.GET(
			"",
			ProtalHandler,
		)
		secureMiddleware := secure.New(secure.Options{
			FrameDeny: true,
		})
		secureFunc := func() gin.HandlerFunc {
			return func(c *gin.Context) {
				err := secureMiddleware.Process(c.Writer, c.Request)

				// If there was an error, do not continue.
				if err != nil {
					c.Abort()
					return
				}

				// Avoid header rewrite if response is a redirection.
				if status := c.Writer.Status(); status > 300 && status < 399 {
					c.Abort()
				}
			}
		}()
		g.Use(secureFunc)
		g.RunTLS(fmt.Sprintf(":%d", config.Server.Port), "./cert/pris.ssdk.icu.pem", "./cert/pris.ssdk.icu.key")
		//// 强制ipv4
		//server := &http.Server{Addr: fmt.Sprintf(":%d", config.Server.Port), Handler: g}
		//ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", config.Server.Port))
		//if err != nil {
		//	panic(err)
		//}
		//type tcpKeepAliveListener struct {
		//	*net.TCPListener
		//}
		//server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
		//if err := g.Run(fmt.Sprintf(":%d", config.Server.Port)); err != nil {
		//	panic(err)
		//}
	}()
}
func SocketHandler(c *gin.Context) {
	src.SocketHandler(c, db)
}
func ProtalHandler(c *gin.Context) {
	src.ProtalHandler(c, db)
}
func GetDB() *sql.DB {
	db, err := sql.Open(dbDriverName, dbName)
	if err != nil {
		panic(err)
	}
	return db
}
