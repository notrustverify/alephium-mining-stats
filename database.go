package main

import (
	"database/sql"
	"fmt"
	"log"
)

func initDb() {

	create := `
  CREATE TABLE IF NOT EXISTS mining_stats (
  block_hash TEXT NOT NULL PRIMARY KEY,
  block_timestamp  UNSIGNED BIG INT NOT NULL,
  groupId TEXT NOT NULL,
  miner_address TEXT NOT NULL
  );`

	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.Exec(create)
	if err != nil {
		log.Fatalln(err)
	}

}

func getStatsData() []miningStat {
	dbGet, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatalln(err)
	}

	if err != nil {
		log.Fatalln(err)
	}
	var miningStats []miningStat

	var (
		block_hash      string
		block_timestamp int64
		groupId         string
		address         string
	)

	for mainGroup := 0; mainGroup <= 3; mainGroup++ {
		for secondGroup := 0; secondGroup <= 3; secondGroup++ {
			if mainGroup > secondGroup {
				continue
			}
			rows, err := dbGet.Query(fmt.Sprintf("select * from mining_stats where groupId='%d%d' order by block_timestamp desc limit 1000", mainGroup, secondGroup))
			if err != nil {
				log.Fatalln(err)
			}
			var stat []groupStat

			for rows.Next() {
				rows.Scan(&block_hash, &block_timestamp, &groupId, &address)
				//fmt.Printf("Group %s\n", groupId)
				stat = append(stat, groupStat{block_hash, block_timestamp, address})
			}
			miningStats = append(miningStats, miningStat{groupId, stat})
			rows.Close()
		}

	}

	return miningStats

}
