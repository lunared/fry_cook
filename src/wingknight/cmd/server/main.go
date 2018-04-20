package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	wingknight "wingknight/internal/api"

	"github.com/go-redis/redis"
)

// GET will fetch up to 20 restaurants in the specified location that have wings
// data should be paginated
// Content-Type: application/json
func GET(w http.ResponseWriter, r *http.Request) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// read query params
	params := r.URL.Query()
	latitude, err := strconv.ParseFloat(params.Get("lat"), 64)
	longitude, err := strconv.ParseFloat(params.Get("lng"), 64)
	log.Print("seaching at ", latitude, ",", longitude)
	if err != nil {
		http.Error(w, "lat and lng must be specified as decimal values", http.StatusBadRequest)
		return
	}
	// fetch restaurants in the area from redis
	data, err := client.GeoRadius("geowings", longitude, latitude,
		&redis.GeoRadiusQuery{
			Radius:    5.0,
			Unit:      "mi",
			WithCoord: true,
			Count:     20,
		}).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Print("found ", len(data), " points")

	// read out the geojson
	var out []*wingknight.GeoJSON
	for _, location := range data {
		restaurant := &wingknight.GeoJSON{}
		if err = json.Unmarshal([]byte(location.Name), restaurant); err != nil {
			log.Fatal("could not decode file as GeoJSON\n", err)
		}
		out = append(out, restaurant)
	}

	// render output differently based on accept header
	if r.Header.Get("accept") == "application/json" {
		// json api for use with web and mobile interfaces
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	} else {
		// provide a readable plain text version of the data
		tmpl, err := template.New("readable").Parse(`
{{define "menuEntry"}}
- {{.Name}}
  {{.Price}} per wing
{{- end}}
{{define "flavor"}}
- {{.Name}}
  Heat Index {{.Heat}}
{{- end}}
{{.Restaurant.Name}}
  {{.Restaurant.Address}}
{{with .Restaurant.Menu -}}
Menu{{range . -}}
{{ template "menuEntry" . }}
{{- end}}
{{- end}}
{{with .Restaurant.Flavors -}}
Flavors{{range . -}}
{{template "flavor" .}}
{{- end}}
{{end}}`)
		if err != nil {
			panic(err)
		}
		for _, location := range out {
			err = tmpl.Execute(w, location)
		}
	}
}

func main() {
	http.HandleFunc("/getwings", GET)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
