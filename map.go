package main


import (
"fmt"
"os"
"log"
"net/http"
)

func getCountries(VisaType string) []string{
	switch VisaType {
		case "EVisa":
			return []string{"thailand","qatar"}

		case "NoVisa":
			return []string{"kenya","zimbabwe"}

		case "VisaOnEntry":
			return []string{"indonesia"}

		case "HardVisa":
			return []string{"USA","canada","mexico"}
		
		default:
			return[]string{"you don messed up aaron"}
	}
}
func MapHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/map.html") // Assuming file.html is in the same directory
}

func CountriesJsonHandler(w http.ResponseWriter, r *http.Request){
	jsonFile,err := os.ReadFile("pages/countries.json")
	
	if err!=nil{
		http.Error(w, "Error reading JSON file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error reading countries.json for /api/countries: %v", err)
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(jsonFile)

}

func main() {
// fileserver := http.FileServer(http.Dir("./pages/"))
// http.Handle("/", fileserver)

http.HandleFunc("/", MapHandler)
http.HandleFunc("/api/countries", CountriesJsonHandler)

fmt.Printf("port running on http://localhost:8080/\n")
  if err := http.ListenAndServe(":8080", nil); err != nil{
    log.Fatal(err)
  }
}

