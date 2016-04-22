package main

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"math/rand"
	"time"
)

const SHLINK_NAME_LENGTH = 32
const SHLINK_LETTER_BYTES = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var dbmap = initDb()

type Shlink struct {
	Id      int64
	Name    string
	Source  string
	Created int64
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func initDb() *gorp.DbMap {
	db, err := sql.Open("mysql", "shlink:shlink@tcp(localhost:3306)/shlink")
	checkErr(err, "sql.Open failed")

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	table := dbmap.AddTableWithName(Shlink{}, "shlinks").SetKeys(true, "Id")
	table.ColMap("Name").SetMaxSize(32).SetUnique(true)
	table.ColMap("Source").SetMaxSize(2048)

	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func generateShlinkName() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, SHLINK_NAME_LENGTH)

	for i := range b {
		b[i] = SHLINK_LETTER_BYTES[rand.Intn(len(SHLINK_LETTER_BYTES))]
	}

	return string(b)
}

func createShlink(source string) Shlink {
	shlink := Shlink{
		Created: time.Now().UnixNano(),
		Name:    generateShlinkName(),
		Source:  source,
	}

	err := dbmap.Insert(&shlink)
	checkErr(err, "Insert failed")

	return shlink
}

func getShlink(shlink_name string) Shlink {
	shlink := Shlink{}
	dbmap.SelectOne(&shlink, "select * from shlinks where name=?", shlink_name)

	return shlink
}

func ShlinkRedirect(c *gin.Context) {
	shlink_name := c.Params.ByName("shlink_name")
	shlink := getShlink(shlink_name)

	if shlink.Name == shlink_name {
		c.Redirect(301, shlink.Source)
	} else {
		c.Redirect(301, "/")
	}
}

func ShlinkPost(c *gin.Context) {
	var data Shlink

	c.Bind(&data)
	shlink := createShlink(data.Source)

	if shlink.Source == data.Source {
		content := gin.H{
			"shlink_name": shlink.Name,
		}
		c.JSON(200, content)
	} else {
		c.JSON(500, gin.H{"result": "An error occured"})
	}
}

func main() {
	router := gin.Default()

	router.Static("/css", "./front/css")
	router.Static("/fonts", "./front/fonts")
	router.Static("/js", "./front/js")

	router.LoadHTMLFiles("front/index.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})

	router.GET("/r/:shlink_name", ShlinkRedirect)

	router.POST("/", ShlinkPost)

	router.Run() // listen and server on 0.0.0.0:8080
}
