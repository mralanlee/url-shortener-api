package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Client struct {
	Sql *sql.DB
}

type UrlDocument struct {
	Slug   string `db:"slug"`
	Source string `db:"source"`
}

type StatsDocument struct {
	TwentyFourHours string `db:"twenty_four_hours" json:"twenty_four_hours"`
	Weekly          string `db:"week" json:"week"`
	Lifetime        string `db:"lifetime" json:"lifetime"`
}

func NewClient() *sql.DB {
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/redirects")

	if err != nil {
		log.Fatal("Error with connection at DB")
		panic(err)
	}

	log.Println("DB connected!")
	return db
}

func Init() *sql.DB {
	db := NewClient()
	tables := []string{
		`
	CREATE TABLE IF NOT EXISTS visits (
		id int(6) unsigned NOT NULL AUTO_INCREMENT,
		slug varchar(8) NOT NULL UNIQUE, 
		timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id) 
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
	`,
		`
	CREATE TABLE IF NOT EXISTS url (
		id int(6) unsigned NOT NULL AUTO_INCREMENT,
		slug varchar(8) NOT NULL UNIQUE,
		source varchar(128) NOT NULL,
		PRIMARY KEY (id)
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
	`,
	}

	for _, s := range tables {
		stmt, prepErr := db.Prepare(s)
		if prepErr != nil {
			log.Fatalf(prepErr.Error())
			panic(prepErr.Error())
		}

		_, stmtErr := stmt.Exec()

		if stmtErr != nil {
			log.Fatal("Error with creating table")
			panic(stmtErr.Error())
		}
	}

	return db
}

func (c *Client) randomSlug() string {
	var chars = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	length := 8

	var b strings.Builder

	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	slug := string(b.String())

	var tmp struct {
		Count int
	}
	query := c.Sql.QueryRow("SELECT COUNT(id) FROM url where slug=?;", slug)
	err := query.Scan(&tmp.Count)

	if err != nil {
		log.Fatal("Error at slug check")
	}

	if tmp.Count > 0 {
		return c.randomSlug()
	}

	return slug
}

func (c *Client) Insert(source string) string {
	stmt, err := c.Sql.Prepare("INSERT INTO url(slug, source) VALUES(?,?);")

	if err != nil {
		panic(err.Error())
	}

	id := c.randomSlug()
	_, insErr := stmt.Exec(id, source)

	if insErr != nil {
		fmt.Println(insErr)
		panic(err.Error())
	}

	defer stmt.Close()

	return id
}

func (c *Client) FindSlug(slug string) string {
	var record UrlDocument
	query := c.Sql.QueryRow("SELECT source FROM url WHERE slug=?;", slug)
	err := query.Scan(&record.Source)

	if err != nil {
		record.Source = ""
	}

	return record.Source
}

func (c *Client) Increment(slug string) {
	stmt, err := c.Sql.Prepare("INSERT INTO visits(slug) VALUES(?);")

	_, insErr := stmt.Exec(slug)

	if insErr != nil {
		panic(err.Error())
	}

	defer stmt.Close()
}

func (c *Client) ReadStats(slug string) StatsDocument {
	var stats StatsDocument
	agg := c.Sql.QueryRow("SELECT COUNT(id) as lifetime FROM visits WHERE slug=?;", slug)

	if aggErr := agg.Scan(&stats.Lifetime); aggErr != nil {
		panic(aggErr.Error())
	}

	// if lifetime is 0 the others should also be 0, no need to run the other two queries
	if stats.Lifetime == "0" {
		stats.TwentyFourHours = "0"
		stats.Weekly = "0"

		return stats
	}

	today := time.Now().UTC().Format("2006-01-02 15:04:05")
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
	week := time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02 15:04:05")

	daily := c.Sql.QueryRow("SELECT COUNT(id) as twenty_four_hours FROM visits WHERE slug=? AND timestamp between ? AND ?", slug, yesterday, today)

	if dailyErr := daily.Scan(&stats.TwentyFourHours); dailyErr != nil {
		panic(dailyErr.Error())
	}

	weekly := c.Sql.QueryRow("SELECT COUNT(id) as weekly FROM visits WHERE slug=? AND timestamp between ? AND ?", slug, week, today)
	if weeklyErr := weekly.Scan(&stats.Weekly); weeklyErr != nil {
		panic(weeklyErr.Error())
	}

	return stats
}
