package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"sync"
)

var (
	wg sync.WaitGroup
)

type Artikel struct {
	ArtNr, Preis int
	Name         string
	Kategorie    string
	GebindeGr    int
	Menge        string
	Einheit      string
}

func mergeWorkers(cs ...<-chan Artikel) <-chan Artikel {
	var wg sync.WaitGroup
	out := make(chan Artikel)

	output := func(c <-chan Artikel) {
		for a := range c {
			out <- a
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {

	// start the workers!
	jobs, client, err := loginAndGetWarengruppen(user, passw)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Logged in!\n")

	workerCnt := 10
	workerChans := make([]<-chan Artikel, workerCnt)
	for i := 0; i < workerCnt; i++ {
		workerChans[i] = kategorieWorker(client, jobs)
	}

	artikels := mergeWorkers(workerChans...)

	// prepare the output file
	file, err := os.Create("bode.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bodecsv := csv.NewWriter(file)

	for a := range artikels {
		rec := []string{
			a.Kategorie,
			a.Name,
			strconv.Itoa(a.GebindeGr),
			a.Menge,
			a.Einheit,
			strconv.Itoa(a.ArtNr),
			strconv.Itoa(a.Preis),
		}
		log.Printf("Artikel:%v\n", rec)
		bodecsv.Write(rec)
	}
	bodecsv.Flush()

}
