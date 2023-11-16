package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// select output where transactions hash is coinbase and group 0 -> 1
//select OUTS.block_hash,OUTS.address from outputs OUTS inner join transactions TX ON TX.block_hash=OUTS.block_hash where TX.coinbase=true and (chain_from=0 and chain_to=1) order by TX.block_timestamp limit 1000;

type miningStat struct {
	Group     string      `json:"group"`
	GroupStat []groupStat `json:"group_stat"`
}

type groupStat struct {
	Blockhash      string `json:"hash"`
	BlockTimestamp int64  `json:"block_timestamp"`
	MinerAddress   string `json:"miner_address"`
}

const file string = "mining-stats.db"

func main() {
	initDb()
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("No env file, will use system variable")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	cronUpdate := os.Getenv("CRON_UPDATE_MINING")

	s := gocron.NewScheduler(time.UTC)
	s.Every(cronUpdate).Do(getData, dbHost, dbPort, dbUsername, dbPassword, dbName)
	s.StartAsync()

	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}

	router := gin.Default()
	router.Use(cors.New(corsConfig))
	router.GET("/stats", getStats)

	router.Run("0.0.0.0:8080")

}

func getData(dbHost string, dbPort string, dbUsername string, dbPassword string, dbName string) {
	// encode the special char
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, dbHost, dbPort, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	dbStats, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}

	var (
		block_timestamp int64
		block_hash      string
		address         string
	)

	for mainGroup := 0; mainGroup <= 3; mainGroup++ {
		for secondGroup := 0; secondGroup <= 3; secondGroup++ {
			fmt.Printf("Group %d,%d\n", mainGroup, secondGroup)

			rows, err := db.Query(fmt.Sprintf("select encode(OUTS.block_hash,'hex'),OUTS.address, OUTS.block_timestamp from outputs OUTS inner join transactions TX ON TX.block_hash=OUTS.block_hash where TX.coinbase=true and OUTS.coinbase=true and (chain_from=%d and chain_to=%d) order by TX.block_timestamp desc limit 1000;", mainGroup, secondGroup))

			if err != nil {
				log.Fatalln(err)
			}
			groupId := fmt.Sprintf("%d%d", mainGroup, secondGroup)
			for rows.Next() {
				rows.Scan(&block_hash, &address, &block_timestamp)
				_, err := dbStats.Exec("INSERT OR IGNORE INTO mining_stats VALUES(?,?,?,?);", block_hash, block_timestamp, groupId, address)
				if err != nil {
					log.Fatalln(err)
				}

			}
			rows.Close()

		}

	}
}

func getStats(c *gin.Context) {
	miningStats := getStatsData()
	c.IndentedJSON(http.StatusOK, miningStats)
}
