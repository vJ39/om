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
	"strings"
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
		Name        string `xml:"name"`
		Description string `xml:"description"`
		StyleUrl    string `xml:"styleUrl"`
		LineString  struct {
			Tessellate  string `xml:"tessellate"`
			Coordinates string `xml:"coordinates"`
		} `xml:"LineString"`
		Polygon struct {
			OuterBoundaryIs struct {
				LinearRing struct {
					Tessellate  string `xml:"tessellate"`
					Coordinates string `xml:"coordinates"`
				} `xml:"LinearRing"`
			} `xml:"outerBoundaryIs"`
		} `xml:"Polygon"`
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
			var (
				t           int
				coordinates string = ""
			)
			placemark.Point.Coordinates = splitcoordinates(trim(placemark.Point.Coordinates))
			placemark.LineString.Coordinates = splitLinecoordinates(placemark.LineString.Coordinates)
			placemark.Polygon.OuterBoundaryIs.LinearRing.Coordinates = splitLinecoordinates(placemark.Polygon.OuterBoundaryIs.LinearRing.Coordinates)
			if len(placemark.Point.Coordinates) > 0 {
				t = 1
			} else if len(placemark.LineString.Coordinates) > 0 {
				t = 2
				placemark.Point.Coordinates = "\t"
				coordinates = placemark.LineString.Coordinates
			} else if len(placemark.Polygon.OuterBoundaryIs.LinearRing.Coordinates) > 0 {
				t = 3
				placemark.Point.Coordinates = "\t"
				coordinates = placemark.Polygon.OuterBoundaryIs.LinearRing.Coordinates
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
				trimLF(placemark.Name),
				placemark.Point.Coordinates,
				trimLF(placemark.Description),
				trimLF(placemark.ExtendedData.Data.Value),
				trimLF(placemark.StyleUrl),
				t,
				trimLF(coordinates),
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
	var re *regexp.Regexp = regexp.MustCompile(`^[\n\s]*([^\n\s]*)[\n\s]*$`)
	return re.ReplaceAllString(s, "$1")
}

func trim(s string) string {
	var re *regexp.Regexp = regexp.MustCompile(`[\n\s]*([\d\.,]+)[\s\n]*`)
	return re.ReplaceAllString(s, "$1")
}

func splitcoordinates(s string) string {
	var (
		r *regexp.Regexp
	)
	r = regexp.MustCompile(`(\d+\.\d+),(\d+\.\d+),0`)
	return r.ReplaceAllString(s, "$1\t$2")
}

func splitLinecoordinates(s string) string {
	type (
		coordinates struct {
			in  []string
			out []string
		}
	)
	var (
		c coordinates
	)
	c.in = strings.Split(s, "\n")
	for _, coodinate := range c.in {
		var (
			trimed string = trimLF(trim(coodinate))
		)
		if len(trimed) < 1 {
			continue
		}
		c.out = append(c.out, trimed)
	}
	return strings.Join(c.out, ":")
}
