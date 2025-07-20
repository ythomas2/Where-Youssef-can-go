package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var visaMap = make(map[string]string) //map the country code to the visa type   

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

func getVisaMap(){
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

		if line[0]=="EGY"{
			for i := 1; i < len(line); i++ {
				visaMap[header[i]] = line[i]
			}
		}
	}

	
	// read the egypt line
	// populate map variable
}


func MapHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/map.html") // Assuming file.html is in the same directory
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
	// unmarshall data into json object
	return jsonData
}


func CountriesJsonHandler(w http.ResponseWriter, r *http.Request){
	jsonFile,err := os.ReadFile("pages/countries.json")
	jsonFile = annotateJson(jsonFile) // there is probably alot of copying in this implementation
	if err!=nil{
		http.Error(w, "Error reading JSON file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error reading countries.json for /api/countries: %v", err)
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(jsonFile)

}

func main() {
	getVisaMap()

	http.HandleFunc("/", MapHandler)
	http.HandleFunc("/api/countries", CountriesJsonHandler)

	fmt.Printf("port running on http://localhost:8080/\n")
	if err := http.ListenAndServe(":8080", nil); err != nil{
		log.Fatal(err)
	}
}
