package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				log.Printf("KatWorker[%s] NewDocumentFromResponse Error:%s", kat, err)
				break
			}

			doc.Find("tr.text_lauftext_shop").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the band and title
				artikel := s.Find(".spalteArtikel1 > a").Text()

				paranLeft := strings.LastIndex(artikel, "(")
				if paranLeft < 0 {
					log.Printf("KatWorker[%s] Artikel[%s] No left paren\n.", kat, artikel)
					return
				}

				paranRight := strings.LastIndex(artikel, ")")
				if paranRight < 0 {
					log.Printf("KatWorker[%s] Artikel[%s] No right paren\n.", kat, artikel)
					return
				}

				artnrStr := artikel[paranLeft+8 : paranRight]
				if len(artnrStr) == 0 {
					log.Printf("KatWorker[%s] Artikel[%s] No artikelNr\n.", kat, artikel)
					return
				}

				artnr, err := strconv.Atoi(artnrStr)
				if err != nil {
					log.Printf("KatWorker[%s] Artikel[%s] Error:%s", kat, artikel, err)
					return
				}

				// ugly hack because bode remove the class ".spalteArtikel3"
				subSel := s.Find("td")
				if subSel.Length() < 4 {
					log.Printf("SubSel <3 for artikel: %s\nSubSel:%v", artikel, subSel)
					return
				}

				preis := subSel.Eq(3).Text()

				// split preis by the decimal poin
				preisParts := strings.Split(preis, ".")
				if len(preisParts) != 2 {
					log.Printf("len(preisPart) != 2 ! %v\n", preis)
					return
				}

				hunderter, err := strconv.Atoi(preisParts[0])
				if err != nil {
					log.Printf("KatWorker[%s] Artikel[%s] Error:%s", kat, artikel, err)
					return
				}

				// remove the malencoded euro sign from the end
				zehner, err := strconv.Atoi(preisParts[1][:len(preisParts[1])-1])
				if err != nil {
					log.Printf("KatWorker[%s] Artikel[%s] Error:%s", kat, artikel, err)
					return
				}

				out <- Artikel{
					Kategorie: kat,
					Name:      artikel,
					ArtNr:     artnr,
					Preis:     hunderter*100 + zehner,
				}
			})
		}
		close(out)
	}()

	return out

}
