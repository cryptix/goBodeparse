package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

var (
	loginCheckRegexp   = regexp.MustCompile(`(angemeldet!)`)
	warengruppenRegexp = regexp.MustCompile(`loadProductPage.php\?wg=([\w-]+)">`)
)

type Artikel struct {
	ArtNr, Preis int
	Kategorie    string
}

func main() {

	// preserve cookies from the response
	jar, err := cookiejar.New(nil)
	checkErr(err)

	client := http.Client{nil, nil, jar}

	// try to log in
	loginCreds := url.Values{
		"Kunde":     []string{user},
		"Pass":      []string{passw},
		"sentLogin": []string{"1"},
	}

	resp, err := client.PostForm("http://bodenaturkost.de/php/loadPage.php", loginCreds)
	checkErr(err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Login Status Code: %v\n", resp.StatusCode)
	}

	loginBody, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	if loginCheckRegexp.Match(loginBody) == false {
		log.Fatal("Login gescheitert.\n")
	}

	log.Printf("Logged in!\n")

	// start the workers!
	jobs := make(chan string, 50) // ~35 categories on the catalog currently
	artikel := make(chan *Artikel, 100)

	for i := 0; i < 4; i++ {
		go kategorieWorker(client, jobs, artikel)
	}

	// find all categories from the landing page
	for _, gruppe := range warengruppenRegexp.FindAllStringSubmatch(string(loginBody), -1) {
		if len(gruppe) != 2 {
			log.Fatalf("Strange Warengruppe:%v\n", gruppe)
		}

		// send categorie to the workers
		jobs <- gruppe[1]
	}
	close(jobs) // no more categories to parse

	// prepare the output file
	file, err := os.Create("bode.csv")
	checkErr(err)
	defer file.Close()

	bodecsv := csv.NewWriter(file)

	// read all the articles from the chanel
	for arti := range artikel {
		rec := []string{
			strconv.Itoa(arti.ArtNr),
			strconv.Itoa(arti.Preis),
			arti.Kategorie}
		log.Printf("Artikel:%v\n", rec)
		bodecsv.Write(rec)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
