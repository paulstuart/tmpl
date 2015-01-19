package main

import (
	"os"
	"path"
	"time"

	dbu "github.com/paulstuart/dbutil"
)

var (
	version           = "1.0.0"
	Hostname, _       = os.Hostname()
	Basedir, _        = os.Getwd() // get abs path now, as we will be changing dirs
	log_layout        = "2006-01-02 15:04:05.999"
	start_time        = time.Now()
	http_port         = 1977
	assets_dir        = "assets"
	sqlDir            = "sql" // dir containing sql schemas, etc
	sqlSchema         = path.Join(Basedir, sqlDir, "schema.sql")
	dbFile            = path.Join(Basedir, "system.db")
	systemLocation, _ = time.LoadLocation("Local")
	db                dbu.DBU
)

const (
	pathPrefix     = "/x"
	sessionMinutes = 120
	secretKey      = "secret.key"
	loginPrompt    = ""
)

func main() {
	db = dbu.CreateIfMissing(dbFile, sqlSchema)
	webServer(webHandlers)
}
