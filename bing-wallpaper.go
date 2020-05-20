package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"io"
	"os"
	"flag"

	"github.com/PuerkitoBio/goquery"
)

const BING_URL = "https://www.bing.com"

// Fetches wallpaper link from Bing
func fetchWallpaperLink() (string, error) {
	resp, err := http.Get(BING_URL)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return "", err
	}

	// Find the #preloadBg element, which contains the wallpaper link
	sel := doc.Find("#preloadBg").First()
	link, exists := sel.Attr("href")

	if !exists {
		return "", errors.New("Could not find #preloadBg element on Bing. Cannot fetch wallpaper link.")
	}

	return fmt.Sprintf("%s%s", BING_URL, link), nil
}

// Returns the filename for a wallpaper link
// Found in the id GET parameter
func getWallpaperName(link string) (string, error) {
	// Parse the wallpaper link into a *URL
	u, err := url.Parse(link)

	if err != nil {
		return "", err
	}

	// Extract GET parameters from the URL
	getParams, err := url.ParseQuery(u.RawQuery)
	idParam, ok := getParams["id"]

	if !ok {
		return "", errors.New(fmt.Sprintf("Could not find id GET parameter in link: %s. Cannot resolve wallpaper filename.", link))
	}

	if len(idParam) != 1 {
		return "", errors.New(fmt.Sprintf("id GET parameter is not valid in link: %s. Cannot resolve wallpaper filename.", link))
	}

	return idParam[0], nil
}

// Saves wallpaper from link to destination
// Returns the final destination path
func saveWallpaper(link string, dest string, filename string) (string, error) {
	outputDest := filepath.Join(dest, filename)

	resp, err := http.Get(link)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	f, err := os.Create(outputDest)

	if err != nil {
		return "", err
	}

	defer f.Close()

	// Copy data to the file
	_, err = io.Copy(f, resp.Body)

	if err != nil {
		return "", err
	}

	return outputDest, nil
}

func main() {
	dest := flag.String("output-dir", "", "Output directory to save wallpaper to")
	flag.Parse()

	if *dest == "" {
		log.Fatal("You must provide an output directory using the -output-dir flag")
	}

	link, err := fetchWallpaperLink()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found wallpaper link: %s\n", link)

	filename, err := getWallpaperName(link)

	if err != nil {
		log.Fatal(err)
	}

	finalDest, err := saveWallpaper(link, *dest, filename)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Saved wallpaper to: %s\n", finalDest)
}
