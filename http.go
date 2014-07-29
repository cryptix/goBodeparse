package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

var (
	loginCheckRegexp   = regexp.MustCompile(`(angemeldet!)`)
	warengruppenRegexp = regexp.MustCompile(`loadProductPage.php\?wg=([\w-]+)">`)
)

func loginAndGetWarengruppen(user, passw string) (<-chan string, *http.Client, error) {
	// preserve cookies from the response
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{Jar: jar}

	// try to log in
	loginCreds := url.Values{
		"Kunde":     []string{user},
		"Pass":      []string{passw},
		"sentLogin": []string{"1"},
	}

	resp, err := client.PostForm("http://bodenaturkost.de/php/loadPage.php", loginCreds)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("Login Status Code: %v\n", resp.StatusCode)
	}

	loginBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if loginCheckRegexp.Match(loginBody) == false {
		return nil, nil, fmt.Errorf("Login gescheitert.\n")
	}

	jobs := make(chan string)
	go func() {
		// find all categories from the landing page
		for _, gruppe := range warengruppenRegexp.FindAllStringSubmatch(string(loginBody), -1) {
			if len(gruppe) != 2 {
				log.Printf("ERROR: Strange Warengruppe:%v\n", gruppe)
				continue
			}

			// send categorie to the workers
			jobs <- gruppe[1]
		}
		close(jobs) // no more categories to parse
	}()

	return jobs, client, nil
}
