package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/Masterminds/semver"
	"github.com/robfig/cron"
)

type providerAPIs struct {
	APIs []providerAPI `json:"datasources"`
}

type providerAPI struct {
	Provider string `json:"name"`
	Settings struct {
		Url string `json:"url"`
	} `json:"settings"`
}

type apiVersion struct {
	Provider string
	Version  string `json:"api_version"`
}

type capiMap struct {
	Date       string `json:"published_at"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

type parsedCapiMap struct {
	Date    string
	Version string
}

type finalMap struct {
	Provider string
	Version  string
	Date     string
}

func convergeData(providerVersions []apiVersion, capiVersions []parsedCapiMap) []finalMap {
	var finalArray []finalMap

	for _, provider := range providerVersions {
		for _, release := range capiVersions {
			if provider.Version == release.Version {
				finalArray = append(finalArray, finalMap{Provider: provider.Provider, Version: provider.Version, Date: release.Date})
			}
		}
	}
	return finalArray
}

func fetchAPIs(client http.Client) (providerAPIs, error) {
	var apis providerAPIs
	resp, err := client.Get("https://cf-api-version.mybluemix.net/dashboard.json")
	if err != nil {
		return providerAPIs{}, err
	}
	err = json.NewDecoder(resp.Body).Decode(&apis)
	if err != nil {
		return providerAPIs{}, err
	}
	return apis, nil
}

func getAPIVersions(client http.Client, apis providerAPIs) ([]apiVersion, error) {
	var versions []apiVersion

	for _, api := range apis.APIs {
		var v apiVersion
		v.Provider = api.Provider
		url := api.Settings.Url
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	return versions, nil
}

func constructCapiArray(client http.Client) ([]parsedCapiMap, error) {
	var m []capiMap
	var o []parsedCapiMap

	req, err := http.NewRequest("GET", "https://api.github.com/repos/cloudfoundry/capi-release/releases?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", os.Getenv("TOKEN")))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		return nil, err
	}

	for _, release := range m {
		if release.Prerelease || release.Draft {
			continue
		}
		r := regexp.MustCompile(`CC API Version: ([0-9]+\.[0-9]+\.[0-9]+)`)
		match := r.FindStringSubmatch(release.Body)
		if len(match) < 2 {
			continue
		}
		o = append(o, parsedCapiMap{Date: release.Date, Version: match[1]})
	}
	return o, nil
}

func generateVersions() {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{Timeout: timeout}
	apis, err := fetchAPIs(client)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	fmt.Println("INFO: getting provider api versions")
	versions, err := getAPIVersions(client, apis)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	fmt.Println("INFO: getting capi versions")
	capiArray, err := constructCapiArray(client)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	fmt.Println("INFO: converging json")
	finalMap := convergeData(versions, capiArray)

	sort.Slice(finalMap, func(i, j int) bool {
		v1, _ := semver.NewVersion(finalMap[i].Version)
		v2, _ := semver.NewVersion(finalMap[j].Version)
		return v1.Compare(v2) <= 0
	})

	finalBytes, err := json.Marshal(finalMap)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	fmt.Println("INFO: writing versions file")
	finalBytes = append([]byte("var versions = "), finalBytes...)
	err = ioutil.WriteFile("./static/versions.js", finalBytes, 0666)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
}

func main() {
	c := cron.New()
	c.AddFunc("@every 30s", generateVersions)
	c.Start()

	http.Handle("/", http.FileServer(http.Dir("static")))
	port, set := os.LookupEnv("PORT")
	if !set {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
