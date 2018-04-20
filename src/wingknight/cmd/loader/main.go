// WingKnight Loader
// How to use
//   go run src/wingknight/cmd/loader/main.go [WingDBDir]
//
// make sure you have a version of redis installed that has georedis
//
// Loader will populate the redis with our WingDB so we can do queries against it
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	wk "wingknight/internal/api"

	"github.com/go-redis/redis"
)

// Rev Up Those Friers and populate the redis with our WingDB
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must specify WingDB directory in arguments\n")
	}
	wingdb := os.Args[1] // first arg should be the filepath of the WingDb

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal("could not connect with redis\n", err)
		panic(err)
	}

	var batch []*redis.GeoLocation

	// read wingdb dir for files
	files, err := ioutil.ReadDir(wingdb)
	if err != nil {
		log.Fatal("could not read WingDB directory\n", err)
	}

	// iterate through the files
	for _, f := range files {
		// only read json files
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}

		file, err := os.Open(path.Join(wingdb, f.Name()))
		if err != nil {
			log.Fatal("opening file failed\n", err.Error())
		}
		// load data as a geojson for validation
		parse := json.NewDecoder(file)
		data := &wk.GeoJSON{}
		if err = parse.Decode(data); err != nil {
			log.Fatal("could not decode file as GeoJSON\n", err)
		}
		log.Print("data loaded ", data)

		// reencode as safe data and pack for redis
		out, _ := json.Marshal(data)
		batch = append(batch, &redis.GeoLocation{
			Latitude:  data.Geometry.Coordinates[0],
			Longitude: data.Geometry.Coordinates[1],
			Name:      string(out),
		})

		// batch our requests to redis for uploading
		if len(batch) >= 100 {
			_, err := client.GeoAdd("geowings", batch...).Result()
			if err != nil {
				log.Fatal("could not load files\n", err.Error())
			}
			batch = batch[:0]
		}
		log.Print("done reading files")
	}
	// submit last of the files
	if len(batch) >= 0 {
		_, err := client.GeoAdd("geowings", batch...).Result()
		if err != nil {
			log.Fatal("could not load files\n", err.Error())
		}
	}

	client.Close()
}
