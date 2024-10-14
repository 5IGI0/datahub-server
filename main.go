package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

var GlobalContext = struct {
	Database               *sqlx.DB
	DefaultRateLimit       int64
	DefaultRateLimitWindow int64
	ForwardedFromHdr       string
}{}

func ConnectDatabase() {
	GlobalContext.Database = sqlx.MustConnect("mysql",
		fmt.Sprintf("%s:%s@(%s:%s)/%s",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME")))
	GlobalContext.Database.SetConnMaxLifetime(time.Minute * 3)
	GlobalContext.Database.SetMaxOpenConns(10)
	GlobalContext.Database.SetMaxIdleConns(10)
}

func StartApi() {
	ConnectDatabase()
	r := mux.NewRouter()

	var err error
	GlobalContext.DefaultRateLimit, err = strconv.ParseInt(os.Getenv("DEFAULT_RATELIMIT"), 10, 64)
	AssertError(err)
	GlobalContext.DefaultRateLimitWindow, err = strconv.ParseInt(os.Getenv("DEFAULT_RATELIMIT_WINDOW"), 10, 64)
	AssertError(err)
	GlobalContext.ForwardedFromHdr = os.Getenv("FORWARDED_FROM_HEADER")

	/* find individuals by email */
	r.HandleFunc("/api/v1/individuals/email/{username}@{domain}", ApiDecorator(ApiIndividualByEmail, 0))
	r.HandleFunc("/api/v1/individuals/email/{username}@", ApiDecorator(ApiIndividualByEmail, 0))
	r.HandleFunc("/api/v1/individuals/email/@{domain}", ApiDecorator(ApiIndividualByEmail, 0))
	r.HandleFunc("/api/v1/individuals/email/{username}@{domain}/{page}", ApiDecorator(ApiIndividualByEmail, 0))
	r.HandleFunc("/api/v1/individuals/email/{username}@/{page}", ApiDecorator(ApiIndividualByEmail, 0))
	r.HandleFunc("/api/v1/individuals/email/@{domain}/{page}", ApiDecorator(ApiIndividualByEmail, 0))

	/* domain-related */
	r.HandleFunc("/api/v1/domains/subdomains/{domain}", ApiDecorator(ApiDomainSubs, 0))
	r.HandleFunc("/api/v1/domains/scan/{domain}", ApiDecorator(ApiDomainScan, 0))

	/* http services */
	r.HandleFunc("/api/v1/services/http", ApiDecorator(ApiHttpServicesSearch, 0))
	r.HandleFunc("/api/v1/services/http_by_header", ApiDecorator(ApiHttpServicesSearchByHeader, 0))
	r.HandleFunc("/api/v1/services/http_by_meta", ApiDecorator(ApiHttpServicesSearchByMeta, 0))
	r.HandleFunc("/api/v1/services/http_by_robots_txt", ApiDecorator(ApiHttpServicesSearchByRobotsTxt, 0))
	r.HandleFunc("/api/v1/services/http_by_cert", ApiDecorator(ApiHttpServicesSearchByCert, 0))
	r.HandleFunc("/api/v1/services/http/{page}", ApiDecorator(ApiHttpServicesSearch, 0))
	r.HandleFunc("/api/v1/services/http_by_header/{page}", ApiDecorator(ApiHttpServicesSearchByHeader, 0))
	r.HandleFunc("/api/v1/services/http_by_meta/{page}", ApiDecorator(ApiHttpServicesSearchByMeta, 0))
	r.HandleFunc("/api/v1/services/http_by_robots_txt/{page}", ApiDecorator(ApiHttpServicesSearchByRobotsTxt, 0))
	r.HandleFunc("/api/v1/services/http_by_cert/{page}", ApiDecorator(ApiHttpServicesSearchByCert, 0))

	/* IP-related */
	r.HandleFunc("/api/v1/addrs/addr/{addr}", ApiDecorator(ApiAddrInfo, 0))

	/* feed-related */
	r.HandleFunc("/api/v1/individuals/add", ApiPostDecorator(ApiIndividualAdd, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/domains/add", ApiPostDecorator(ApiDomainAdd, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/domains/add_scan", ApiPostDecorator(ApiDomainAddScan, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/domains/outdated", ApiDecorator(ApiDomainsOutdated, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/by_hash_id/{hash_id}/feed_state", ApiDecorator(ApiGetDiscourseInstanceState, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/add_instance", ApiPostDecorator(ApiAddDiscourseInstance, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/by_hash_id/{hash_id}/add_categories", ApiPostDecorator(ApiAddDiscourseCategories, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/by_hash_id/{hash_id}/add_topics", ApiPostDecorator(ApiAddDiscourseTopics, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/by_hash_id/{hash_id}/add_posts", ApiPostDecorator(ApiAddDiscoursePosts, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/by_hash_id/{hash_id}/add_users", ApiPostDecorator(ApiAddDiscourseUsers, API_NO_RATELIMIT|API_FEED))
	r.HandleFunc("/api/v1/discourses/list", ApiDecorator(ApiGetDiscourseInstanceList, API_NO_RATELIMIT|API_FEED))

	/* misc */
	r.HandleFunc("/api/v1/stats", ApiDecorator(ApiStats, 0))
	r.HandleFunc("/api/v1/token_info", ApiDecorator(ApiTokenInfo, API_NO_RATELIMIT))

	/* admin endpoints */
	r.HandleFunc("/api/v1/admin/create_token", ApiPostDecorator(ApiTokenCreate, API_ADMIN))

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
