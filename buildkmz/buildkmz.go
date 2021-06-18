package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	OM struct {
		z     *zip.Writer
		kml   *os.File
		flags struct {
			tsv *string
		}
		template struct {
			root                     []byte
			DocumentFolder0Placemark bytes.Buffer
			placemark                []byte
			placemarkExtendedData    []byte
			polygon                  []byte
		}
		val struct {
			placemark struct {
				DocumentFolder0PlacemarkName                  []byte
				DocumentFolder0PlacemarkDescription           []byte
				DocumentFolder0PlacemarkStyleUrl              []byte
				DocumentFolder0PlacemarkExtendedDataDataValue []byte
				DocumentFolder0PlacemarkPointCoordinates      []byte
				DocumentFolder0PlacemarkExtendedData          []byte
			}
			linestring struct {
				DocumentFolder0PlacemarkLineStringCoordinates []byte
			}
			polygon struct {
				DocumentFolder0PlacemarkPolygonOuterBoundaryIsLinearRingCoordinates []byte
			}
		}
	}
)

func (o *OM) InitFlag() {
	if file, err := os.Executable(); err == nil {
		os.Chdir(filepath.Dir(file))
	}
	o.flags.tsv = flag.String("f", "master.tsv", `マスターシートからコピペしたtsv`)
	flag.Parse()
}

func (o *OM) LoadTSV() {
	var (
		r   *os.File
		c   *csv.Reader
		err error
	)
	if r, err = os.Open(*o.flags.tsv); err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	c = csv.NewReader(r)
	c.Comma = '\t'
	c.LazyQuotes = true
	for {
		var line []string
		if line, err = c.Read(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		o.ParseTSV(line)
		o.template.root = o.loadFile("templates/template.kml")
		switch line[6] {
		case "1":
			o.template.placemark = o.loadFile("templates/template.placemark.kml")
		case "2":
			o.template.placemark = o.loadFile("templates/template.linestring.kml")
		case "3":
			o.template.placemark = o.loadFile("templates/template.polygon.kml")
		}
		o.template.placemarkExtendedData = o.loadFile("templates/template.placemark.ExtendedData.kml")
		o.template.DocumentFolder0Placemark.Write(
			o.SetValPlacemark(),
		)
	}
}

func (p *OM) loadFile(filename string) []byte {
	var (
		r   *os.File
		b   []byte
		err error
	)
	if r, err = os.Open(filename); err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	if b, err = ioutil.ReadAll(r); err != nil {
		log.Fatal(err)
	}
	return b
}

func (o *OM) SetValTemplate() {
	o.template.root = bytes.ReplaceAll(o.template.root, []byte(`$$$DocumentName$$$`), []byte(fmt.Sprintf("OM - %s", time.Now().String())))
	o.template.root = bytes.ReplaceAll(o.template.root, []byte(`$$$DocumentDescription$$$`), []byte(""))
	o.template.root = bytes.ReplaceAll(o.template.root, []byte(`$$$DocumentFolder0Name$$$`), []byte("街頭ポイント"))
	o.template.root = bytes.ReplaceAll(o.template.root, []byte(`$$$DocumentFolder0Placemark$$$`), o.template.DocumentFolder0Placemark.Bytes())
}

func (o *OM) ParseTSV(tsv []string) {
	o.val.placemark.DocumentFolder0PlacemarkName = []byte(tsv[0])
	o.val.placemark.DocumentFolder0PlacemarkDescription = []byte(tsv[3])
	o.val.placemark.DocumentFolder0PlacemarkExtendedDataDataValue = []byte(tsv[4])
	o.val.placemark.DocumentFolder0PlacemarkStyleUrl = []byte(tsv[5])
	switch tsv[6] {
	case "1":
		o.val.placemark.DocumentFolder0PlacemarkPointCoordinates = []byte(fmt.Sprintf("\n            %s,%s,0\n          ", tsv[1], tsv[2]))
	case "2":
		var coordinates = strings.ReplaceAll(tsv[7], ":", "\n")
		o.val.linestring.DocumentFolder0PlacemarkLineStringCoordinates = []byte(coordinates)
	case "3":
		var coordinates = strings.ReplaceAll(tsv[7], ":", "\n")
		o.val.polygon.DocumentFolder0PlacemarkPolygonOuterBoundaryIsLinearRingCoordinates = []byte(coordinates)
	}
}

func (o *OM) SetValPlacemark() []byte {
	var (
		p []byte = o.template.placemark
		e []byte = o.template.placemarkExtendedData
	)
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkName$$$`), o.val.placemark.DocumentFolder0PlacemarkName)
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkDescription$$$`), o.val.placemark.DocumentFolder0PlacemarkDescription)
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkStyleUrl$$$`), o.val.placemark.DocumentFolder0PlacemarkStyleUrl)
	if len(o.val.placemark.DocumentFolder0PlacemarkExtendedDataDataValue) > 0 {
		e = bytes.ReplaceAll(e, []byte(`$$$DocumentFolder0PlacemarkExtendedDataDataValue$$$`), o.val.placemark.DocumentFolder0PlacemarkExtendedDataDataValue)
	} else {
		e = []byte("")
	}
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkPointCoordinates$$$`), o.val.placemark.DocumentFolder0PlacemarkPointCoordinates)
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkLineStringCoordinates$$$`), o.val.linestring.DocumentFolder0PlacemarkLineStringCoordinates)
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkPolygonOuterBoundaryIsLinearRingCoordinates$$$`), o.val.polygon.DocumentFolder0PlacemarkPolygonOuterBoundaryIsLinearRingCoordinates)
	p = bytes.ReplaceAll(p, []byte(`$$$DocumentFolder0PlacemarkExtendedData$$$`), e)
	return p
}

func (o *OM) SaveKml() {
	var (
		err error
	)
	if o.kml, err = os.Create("doc.kml"); err != nil {
		log.Fatal(err)
	}
	defer o.kml.Close()
	o.kml.Write(o.template.root)
}

func (o *OM) Zip() {
	var (
		r   *os.File
		err error
	)
	if r, err = os.Create("master.kmz"); err != nil {
		log.Fatal(err)
	}
	o.z = zip.NewWriter(r)
	defer o.z.Close()

	for _, file := range []string{
		`doc.kml`,
		`images/icon-1.png`,
		`images/icon-10.png`,
		`images/icon-11.png`,
		`images/icon-12.png`,
		`images/icon-2.png`,
		`images/icon-3.png`,
		`images/icon-4.png`,
		`images/icon-5.png`,
		`images/icon-6.png`,
		`images/icon-7.png`,
		`images/icon-8.png`,
		`images/icon-9.png`,
		`images/c000.jpg`,
		`images/c020.jpg`,
		`images/c045.jpg`,
		`images/c070.jpg`,
		`images/c090.jpg`,
		`images/c110.jpg`,
		`images/c135.jpg`,
		`images/c160.jpg`,
		`images/c180.jpg`,
		`images/c200.jpg`,
		`images/c225.jpg`,
		`images/c250.jpg`,
		`images/c270.jpg`,
		`images/c290.jpg`,
		`images/c315.jpg`,
		`images/c340.jpg`,
	} {
		o.addFile(file)
	}
}

func (o *OM) addFile(path string) {
	var (
		info   os.FileInfo
		header *zip.FileHeader
		w      io.Writer
		err    error
	)
	if info, err = os.Lstat(path); err != nil {
		log.Fatal(err)
	}
	if header, err = zip.FileInfoHeader(info); err != nil {
		log.Fatal(err)
	}

	header.Name = path
	header.Method = zip.Deflate
	if w, err = o.z.CreateHeader(header); err != nil {
		log.Fatal(err)
	}

	if _, err = w.Write(o.loadFile(path)); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var (
		om = OM{}
	)
	om.InitFlag()
	om.LoadTSV()
	om.SetValTemplate()
	om.SaveKml()
	om.Zip()
}
