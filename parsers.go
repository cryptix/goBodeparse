package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	artNrRegexp = regexp.MustCompile(`\(Art.Nr.([0-9]*)\)<td class="spalteArtikel2" align="right" valign="middle">[0-9/]*k?g</td><td align="right" valign="middle">([0-9]*\.[0-9]*).</td><td class="spalteArtikel4`)
)

func kategorieWorker(client http.Client, jobs <-chan string, artikelChan chan<- *Artikel) {
	// read jobs until the channel is closed
	for kat := range jobs {
		// request categorie page
		url := fmt.Sprintf("http://bodenaturkost.de/php/loadProductPage.php?wg=%s", kat)
		resp, err := client.Get(url)
		checkErr(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		checkErr(err)

		// find all the products
		for _, match := range artNrRegexp.FindAllStringSubmatch(string(body), -1) {
			if len(match) > 3 {
				log.Fatalf("len(match) > 3 ! %v\n", match)
			}

			// convert artNr to int
			artnr, err := strconv.Atoi(match[1])
			checkErr(err)

			// split cost string into ints
			preisParts := strings.Split(match[2], ".")
			if len(preisParts) != 2 {
				log.Fatalf("len(preisPart) != 2 ! %v\n", preisParts)
			}
			hunderter, err := strconv.Atoi(preisParts[0])
			checkErr(err)

			zehner, err := strconv.Atoi(preisParts[1])
			checkErr(err)

			artikelChan <- &Artikel{artnr, hunderter*100 + zehner, kat}
		}
	}
	close(artikelChan)
}
