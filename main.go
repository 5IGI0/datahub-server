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

	/* find individuals by email */
	r.HandleFunc("/api/v1/individuals/email/{username}@{domain}", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/{username}@", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/@{domain}", ApiDecorator(ApiIndividualByEmail))

	/* feed-related */
	r.HandleFunc("/api/v1/individuals/add", ApiPostDecorator(ApiIndividualAdd))

	/* start server */
	panic(http.ListenAndServe(os.Getenv("LISTEN_ADDR"), r))
}
