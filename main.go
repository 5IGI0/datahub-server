package main

import (
	"fmt"
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

func ConnectDatabase() {
	GlobalContext.Database = sqlx.MustConnect("mysql", os.Getenv("DATABASE_URI"))
	GlobalContext.Database.SetConnMaxLifetime(time.Minute * 3)
	GlobalContext.Database.SetMaxOpenConns(10)
	GlobalContext.Database.SetMaxIdleConns(10)
}

func StartApi() {
	ConnectDatabase()
	r := mux.NewRouter()

	/* find individuals by email */
	r.HandleFunc("/api/v1/individuals/email/{username}@{domain}", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/{username}@", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/@{domain}", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/{username}@{domain}/{page}", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/{username}@/{page}", ApiDecorator(ApiIndividualByEmail))
	r.HandleFunc("/api/v1/individuals/email/@{domain}/{page}", ApiDecorator(ApiIndividualByEmail))

	/* domain-related */
	r.HandleFunc("/api/v1/domains/subdomains/{domain}", ApiDecorator(ApiDomainSubs))
	r.HandleFunc("/api/v1/domains/scan/{domain}", ApiDecorator(ApiDomainScan))

	/* http services */
	r.HandleFunc("/api/v1/services/http", ApiDecorator(ApiHttpServicesSearch))
	r.HandleFunc("/api/v1/services/http_by_header", ApiDecorator(ApiHttpServicesSearchByHeader))
	r.HandleFunc("/api/v1/services/http_by_meta", ApiDecorator(ApiHttpServicesSearchByMeta))
	r.HandleFunc("/api/v1/services/http_by_robots_txt", ApiDecorator(ApiHttpServicesSearchByRobotsTxt))
	r.HandleFunc("/api/v1/services/http_by_cert", ApiDecorator(ApiHttpServicesSearchByCert))
	r.HandleFunc("/api/v1/services/http/{page}", ApiDecorator(ApiHttpServicesSearch))
	r.HandleFunc("/api/v1/services/http_by_header/{page}", ApiDecorator(ApiHttpServicesSearchByHeader))
	r.HandleFunc("/api/v1/services/http_by_meta/{page}", ApiDecorator(ApiHttpServicesSearchByMeta))
	r.HandleFunc("/api/v1/services/http_by_robots_txt/{page}", ApiDecorator(ApiHttpServicesSearchByRobotsTxt))
	r.HandleFunc("/api/v1/services/http_by_cert/{page}", ApiDecorator(ApiHttpServicesSearchByCert))

	/* IP-related */
	r.HandleFunc("/api/v1/addrs/addr/{addr}", ApiDecorator(ApiAddrInfo))

	/* feed-related */
	r.HandleFunc("/api/v1/individuals/add", ApiPostDecorator(ApiIndividualAdd))
	r.HandleFunc("/api/v1/domains/add_scan", ApiPostDecorator(ApiDomainAdd))
	r.HandleFunc("/api/v1/domains/outdated", ApiDecorator(ApiDomainsOutdated))

	/* start server */
	panic(http.ListenAndServe(os.Getenv("LISTEN_ADDR"), r))
}

func Usage() {
	fmt.Println("Usage:", os.Args[0], "<subcommand> [ARGS...]")
	fmt.Println("subcommands:")
	fmt.Println(" - api")
	fmt.Println(" - task")
}

func main() {
	if len(os.Args) == 1 {
		Usage()
		return
	}

	switch os.Args[1] {
	case "api":
		StartApi()
	case "task":
		StartTask()
	}
}
