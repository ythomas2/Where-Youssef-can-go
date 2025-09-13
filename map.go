package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

//map the country code to the visa type required for the selected passport
var visaMap = make(map[string]string)

// couldnt find in map:  ATA
// couldnt find in map:  ATF
// couldnt find in map:  BMU
// couldnt find in map:  -99
// couldnt find in map:  FLK
// couldnt find in map:  GRL
// couldnt find in map:  GUF
// couldnt find in map:  CS-KM
// couldnt find in map:  NCL
// couldnt find in map:  PRI
// couldnt find in map:  ESH
// couldnt find in map:  -99

type FeatureCollection struct {
	Type     string    `json:"type"` // Should be "FeatureCollection"
	Features []Feature `json:"features"`
}


// Feature represents a single geographic feature
type Feature struct {
	Type       string     `json:"type"`       // e.g., "Feature"
	ID         string     `json:"id"`         // e.g., "AGO"
	Properties Properties `json:"properties"` // Nested properties
	Geometry   Geometry   `json:"geometry"`   // Nested geometry
}

// Properties holds the specific attributes of the feature
type Properties struct {
	Name string `json:"name"` // e.g., "Angola"
	Visa string `json:"visa"` // e.g., "no"
}

// Geometry holds the geometric definition of the feature
type Geometry struct {
	Type        string        `json:"type"`        // e.g., "MultiPolygon"
	Coordinates any `json:"coordinates"` // Multi-dimensional array for coordinates
}

var validVisaTypes = []string{"visa on arrival","visa free","eta","e-visa","visa required","no admission"}

func isValidVisaType(s string) bool{
	return slices.Contains(validVisaTypes,s)
}



func getVisaMap(passport string) error {
	// read csv
	file,err := os.Open("../passport-index-dataset/passport-index-matrix-iso3.csv")
	if err!=nil{
		log.Fatal("couldnt read the countries csv")
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	header,err := csvReader.Read()
	if err!=nil{
		log.Fatal("couldnt read the countries csv")
	}
	for{
		line,err := csvReader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }

		if line[0]==passport{
			for i := 1; i < len(line); i++ {
				visaMap[header[i]] = line[i]
			}
			return nil
		}
	}
	// if we reach here then no entry was found in the dataset
	return errors.New("Did not find the given country in the dataset")

}


func MapHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "pages/form.html")
}


func annotateJson(jsonData []byte) []byte {
	var parsedJson FeatureCollection
	err := json.Unmarshal(jsonData,&parsedJson)
	if err !=nil{
		log.Fatal("could not parse json data into object")
	}


	for i := range parsedJson.Features {
		visaType,exists := visaMap[parsedJson.Features[i].ID]
		if exists ==false{
			fmt.Println("couldnt find in map: ",parsedJson.Features[i].ID)
		}
		parsedJson.Features[i].Properties.Visa = visaType 
	}

	jsonData,err = json.Marshal(parsedJson)
	if err!=nil{
		log.Fatal("could not parse json object into data")
	}
	return jsonData
}


func getCountriesJson(w http.ResponseWriter, r *http.Request){
	jsonFile,err := os.ReadFile("pages/countries.json")
	jsonFile = annotateJson(jsonFile) // there is probably alot of copying in this implementation
	if err!=nil{
		http.Error(w, "Error reading JSON file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error reading countries.json for /api/countries: %v", err)
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(jsonFile)
}

func postCountriesJson(w http.ResponseWriter, r *http.Request){
	jsonData := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&jsonData)
	defer r.Body.Close()
	if err!=nil{
		http.Error(w, "Error reading JSON file: "+err.Error(), http.StatusInternalServerError)
		log.Print("fucked up json man")
	}
	for key,val := range jsonData{
		if isValidVisaType(val){
			visaMap[key]=val
			} else{
				http.Error(w, "invalid visa type for "+key+" "+val, http.StatusInternalServerError)
				fmt.Fprintln(w, "valid ones are"+strings.Join(validVisaTypes,", "))
				log.Print("invalid visa type for",key,val)
				return
		}
	}
	//if we reach here then we return OK
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "JSON parsed successfully")
}

func countriesApiHandler(w http.ResponseWriter, r *http.Request){
	if r.URL.Path != "/api/countries" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodGet{
		getCountriesJson(w, r)
	}
	if r.Method == http.MethodPost{
		basicAuth(postCountriesJson)(w,r)
	}
}


func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameMatch := (username==os.Getenv("AUTH_USERNAME"))
			passwordMatch := (password==os.Getenv("AUTH_PASSWORD"))

			if usernameMatch && passwordMatch {
				next(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func submitHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Redirect(w,r,"/",http.StatusSeeOther)
		return
	}
	name := r.FormValue("name")
	passport := r.FormValue("passport")
	if (name == "") || (passport == ""){
		http.Error(w, "Fields cannot be empty", http.StatusBadRequest)
		return
	}
	err := getVisaMap(passport)
	if err!= nil{
		http.Error(w, "Did not find your country", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, "pages/map.html")
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", MapHandler)
    mux.HandleFunc("/submit", submitHandler)
    mux.HandleFunc("/api/countries", countriesApiHandler)
	 
    srv := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        IdleTimeout:  time.Minute,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
    }

	 log.Printf("starting server on localhost%s", srv.Addr)
    err := srv.ListenAndServe()
    log.Fatal(err)
}
