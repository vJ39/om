package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

type (
	Map struct {
		XMLName  xml.Name `xml:"kml"`
		Document struct {
			Name        string     `xml:"name"`
			Description string     `xml:"description"`
			Folder      []InFolder `xml:"Folder"`
		}
	}
	InFolder struct {
		Name      string        `xml:"name"`
		Placemark []InPlacemark `xml:"Placemark"`
	}
	InPlacemark struct {
		Name         string `xml:"name"`
		Description  string `xml:"description"`
		StyleUrl     string `xml:"styleUrl"`
		ExtendedData struct {
			Data struct {
				Value string `xml:"value"`
			} `xml:"Data"`
		} `xml:"ExtendedData"`
		Point struct {
			Coordinates string `xml:"coordinates"`
		} `xml:"Point"`
	}
)

func main() {
	var (
		filename *string
	)
	filename = flag.String("f", "", `Googleマップの.kmzファイル`)
	flag.Parse()
	if flag.NFlag() < 1 {
		flag.Usage()
		return
	}

	if err := readfile(*filename); err != nil {
		log.Fatal(err)
	}
	return
}

func readfile(filename string) error {
	var (
		err     error
		z       *zip.ReadCloser
		r       io.ReadCloser
		xmlbyte []byte
		mapdata Map
	)

	// os.Open (write)
	osWriteFile, err := os.Create(filename + ".json")
	if err != nil {
		return err
	}
	defer osWriteFile.Close()

	// zip
	z, err = zip.OpenReader(filename)
	defer z.Close()

	for _, zippedFile := range z.File {
		if zippedFile.Name == "doc.kml" {
			r, err = zippedFile.Open()
			defer r.Close()
			break
		}
	}

	if r == nil {
		log.Fatal("cannot open .kmz file")
	}

	if xmlbyte, err = ioutil.ReadAll(r); err != nil {
		log.Fatal(err.Error())
	}

	if err := xml.Unmarshal(xmlbyte, &mapdata); err != nil {
		log.Fatal(err.Error())
	}

	for _, folder := range mapdata.Document.Folder {
		for _, placemark := range folder.Placemark {
			placemark.Point.Coordinates = trim(placemark.Point.Coordinates)
			placemark.Point.Coordinates = splitCoodinates(placemark.Point.Coordinates)
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
				trimLF(placemark.Name),
				placemark.Point.Coordinates,
				trimLF(placemark.Description),
				trimLF(placemark.ExtendedData.Data.Value),
				trimLF(placemark.StyleUrl),
			)
		}
	}

	jsonbyte, err := json.Marshal(mapdata)
	if err != nil {
		log.Fatal(err.Error())
	}

	osWriteFile.Write(jsonbyte)
	osWriteFile.Write([]byte("\n"))
	return nil
}

func trimLF(s string) string {
	var re *regexp.Regexp = regexp.MustCompile(`[\n\s]+`)
	return re.ReplaceAllString(s, "")
}

func trim(s string) string {
	var re *regexp.Regexp = regexp.MustCompile(`[\n\s]+([\d\.,]+)[\s\n]+`)
	return re.ReplaceAllString(s, "$1")
}

func splitCoodinates(s string) string {
	var (
		r *regexp.Regexp
	)
	r = regexp.MustCompile(`^(\d+\.\d+),(\d+\.\d+),0$`)
	return r.ReplaceAllString(s, "$1\t$2")
}
