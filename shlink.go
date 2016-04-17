package main

import (
    "github.com/gin-gonic/gin"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/coopernurse/gorp"
    "log"
    "time"
    "math/rand"
)

const SHLINK_NAME_LENGTH = 32
const SHLINK_LETTER_BYTES = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var dbmap = initDb()

type Shlink struct {
    Id int64 `db:"id"`
    Name string
    Url string
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
    dbmap.AddTableWithName(Shlink{}, "shlinks").SetKeys(true, "Id").ColMap("Name").SetUnique(true)

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

func createShlink(url string) Shlink {
    shlink := Shlink{
        Created:    time.Now().UnixNano(),
        Name:       generateShlinkName(),
        Url:        url,
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

func ShlinkRedirect (c *gin.Context) {
    shlink_name := c.Params.ByName("shlink_name")
    shlink := getShlink(shlink_name)

    if shlink.Name == shlink_name {
        c.Redirect(301, shlink.Url)
    } else {
        c.Redirect(301, "/")
    }
}

func ShlinkPost (c *gin.Context) {
    var data Shlink

    c.Bind(&data)
    shlink := createShlink(data.Url)

    if shlink.Url == data.Url {
        content := gin.H{
            "shlink_name": shlink.Name,
        }
        c.JSON(201, content)
    } else {
        c.JSON(500, gin.H{"result": "An error occured"})
    }
}

func main() {
    router := gin.Default()

    router.GET("/:shlink_name", ShlinkRedirect)
    router.POST("/", ShlinkPost)

    router.Run() // listen and server on 0.0.0.0:8080
}
