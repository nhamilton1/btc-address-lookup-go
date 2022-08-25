package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/btc/{address}", btcLookup).Methods("GET")
	r.HandleFunc("/", healthCheck).Methods("GET")

	log.Println("Listening on :42069")
	http.ListenAndServe(":42069", r)
	log.Fatal(http.ListenAndServe(":42069", nil))
	http.Handle("/", r)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	//specify status code
	w.WriteHeader(http.StatusOK)

	//update response writer
	fmt.Fprintf(w, "API is up and running")
}

// type errorMessage struct {
// 	Message string `json:"message"`
// }

func btcLookup(w http.ResponseWriter, r *http.Request) {
	//gets from url params
	address := mux.Vars(r)["address"]
	w.Header().Set("Content-Type", "application/json")

	// // testing for valid btc address
	// addressTest := ValidateAddress(address)
	// var errMessage errorMessage
	// if addressTest != "true" {
	// 	errMessage = errorMessage{Message: addressTest}
	// 	j, _ := json.Marshal(errMessage)
	// 	fmt.Fprintf(w, "%v", string(j))
	// 	return
	// }

	//running cmd line
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("./contrib/history.sh", "--venv", address)
	cmd.Dir = "/home/ubuntu/electrs"
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Printf("error: %v\n", err)
	}

	formattedAddr := btcAddressFormatter(stdout.String())
	j, _ := json.Marshal(formattedAddr)

	fmt.Fprintf(w, "%v", string(j))

}

type AddressData struct {
	Txid           string `json:"txid"`
	Blocktimestamp string `json:"blocktimestamp"`
	Height         int    `json:"height"`
	Confirmations  int    `json:"confirmations"`
	Delta          string `json:"delta"`
	Total          string `json:"total"`
}

func btcAddressFormatter(addressInfo string) []AddressData {
	addressInfo = strings.ReplaceAll(addressInfo, " ", "")
	newAddress := strings.Split(addressInfo, "\n")

	var addressStruct AddressData
	results := []AddressData{}

	for _, lines := range newAddress {
		if !strings.ContainsAny(lines, "+") {
			if len([]rune(lines)) > 100 {

				newAddress = strings.Split(strings.TrimSpace(lines), "|")

				height, err := strconv.Atoi(newAddress[3])
				if err != nil {
					fmt.Println(err)
				}
				confirmations, err := strconv.Atoi(newAddress[4])
				if err != nil {
					fmt.Println(err)
				}

				addressStruct = AddressData{Txid: newAddress[1], Blocktimestamp: newAddress[2], Height: height, Confirmations: confirmations, Delta: newAddress[5], Total: newAddress[6]}

				results = append(results, addressStruct)

			}
		}
	}

	return results
}
