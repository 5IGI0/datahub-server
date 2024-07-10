package main

import (
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

var GlobalContext = struct {
	Database *sqlx.DB
}{}

func main() {
	GlobalContext.Database = sqlx.MustConnect("mysql", os.Getenv("DATABASE_URI"))
	GlobalContext.Database.SetConnMaxLifetime(time.Minute * 3)
	GlobalContext.Database.SetMaxOpenConns(10)
	GlobalContext.Database.SetMaxIdleConns(10)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/individuals/domain/{domain}", ApiDecorator(ApiIndividualByDomain))
	r.HandleFunc("/api/v1/individuals/user/{username}", ApiDecorator(ApiIndividualByUsername))
	r.HandleFunc("/api/v1/individuals/email/{username}@{domain}", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/add", ApiPostDecorator(ApiIndividualAdd))
	panic(http.ListenAndServe(os.Getenv("LISTEN_ADDR"), r))
}
