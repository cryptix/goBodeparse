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

func kategorieWorker(client *http.Client, jobs <-chan string) <-chan Artikel {

	out := make(chan Artikel)
	go func() {
		for kat := range jobs {

			// request categorie page
			url := fmt.Sprintf("http://bodenaturkost.de/php/loadProductPage.php?wg=%s", kat)
			resp, err := client.Get(url)
			if err != nil {
				log.Printf("KatWorker[%s] Error:%s", kat, err)
				break
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("KatWorker[%s] Error:%s", kat, err)
				break
			}

			// find all the products
			for _, match := range artNrRegexp.FindAllStringSubmatch(string(body), -1) {
				if len(match) > 3 {
					log.Printf("len(match) > 3 ! %v\n", match)
					continue
				}

				// convert artNr to int
				artnr, err := strconv.Atoi(match[1])
				if err != nil {
					log.Printf("KatWorker[%s] Match[%s] Error:%s", kat, match, err)
					break
				}

				// split cost string into ints
				preisParts := strings.Split(match[2], ".")
				if len(preisParts) != 2 {
					log.Printf("len(preisPart) != 2 ! %v\n", preisParts)
					continue
				}
				hunderter, err := strconv.Atoi(preisParts[0])
				if err != nil {
					log.Printf("KatWorker[%s] Match[%s] Error:%s", kat, match, err)
					break
				}

				zehner, err := strconv.Atoi(preisParts[1])
				if err != nil {
					log.Printf("KatWorker[%s] Match[%s] Error:%s", kat, match, err)
					break
				}

				out <- Artikel{artnr, hunderter*100 + zehner, kat}
			}
		}
		close(out)
	}()

	return out

}
