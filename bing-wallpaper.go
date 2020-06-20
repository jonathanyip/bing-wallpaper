package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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
// If overrideName is non-empty, uses that as the name instead
func getWallpaperName(link string, overrideName string) (string, error) {
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

	filename := idParam[0]

	if overrideName != "" {
		ext := filepath.Ext(filename)
		filename = fmt.Sprintf("%s%s", overrideName, ext)
	}

	return filename, nil
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
	overrideFilename := flag.String("filename", "", "Name to give the wallpaper picture. Extension is automatically added.")
	flag.Parse()

	if *dest == "" {
		log.Fatal("You must provide an output directory using the -output-dir flag")
	}

	link, err := fetchWallpaperLink()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found wallpaper link: %s\n", link)

	filename, err := getWallpaperName(link, *overrideFilename)

	if err != nil {
		log.Fatal(err)
	}

	finalDest, err := saveWallpaper(link, *dest, filename)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Saved wallpaper to: %s\n", finalDest)
}
