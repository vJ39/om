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
		}
		val struct {
			placemark struct {
				DocumentFolder0PlacemarkName                  []byte
				DocumentFolder0PlacemarkDescription           []byte
				DocumentFolder0PlacemarkStyleUrl              []byte
				DocumentFolder0PlacemarkExtendedDataDataValue []byte
				DocumentFolder0PlacemarkPointCoordinates      []byte
			}
		}
	}
)

func (o *OM) InitFlag() {
	o.flags.tsv = flag.String("tsv", "master.tsv", `マスターシートからコピペしたtsv`)
	flag.Parse()
}

func (o *OM) LoadTemplate() {
	o.template.root = o.loadFile("template.kml")
	o.template.placemark = o.loadFile("template.placemark.kml")
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
	o.val.placemark.DocumentFolder0PlacemarkPointCoordinates = []byte(fmt.Sprintf("%s,%s,0", tsv[1], tsv[2]))
	o.val.placemark.DocumentFolder0PlacemarkDescription = []byte(tsv[3])
	o.val.placemark.DocumentFolder0PlacemarkExtendedDataDataValue = []byte(tsv[4])
	o.val.placemark.DocumentFolder0PlacemarkStyleUrl = []byte(tsv[5])
}

func (o *OM) SetValPlacemark() []byte {
	var b []byte = o.template.placemark
	b = bytes.ReplaceAll(b, []byte(`$$$DocumentFolder0PlacemarkName$$$`), o.val.placemark.DocumentFolder0PlacemarkName)
	b = bytes.ReplaceAll(b, []byte(`$$$DocumentFolder0PlacemarkDescription$$$`), o.val.placemark.DocumentFolder0PlacemarkDescription)
	b = bytes.ReplaceAll(b, []byte(`$$$DocumentFolder0PlacemarkStyleUrl$$$`), o.val.placemark.DocumentFolder0PlacemarkStyleUrl)
	b = bytes.ReplaceAll(b, []byte(`$$$DocumentFolder0PlacemarkExtendedDataDataValue$$$`), o.val.placemark.DocumentFolder0PlacemarkExtendedDataDataValue)
	b = bytes.ReplaceAll(b, []byte(`$$$DocumentFolder0PlacemarkPointCoordinates$$$`), o.val.placemark.DocumentFolder0PlacemarkPointCoordinates)
	return b
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

	o.addFile("doc.kml")
	o.addFile("images/icon-2.png")
	o.addFile("images/icon-3.png")
	o.addFile("images/icon-4.png")
	o.addFile("images/icon-5.png")

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
	om.LoadTemplate()
	om.LoadTSV()
	om.SetValTemplate()
	om.SaveKml()
	om.Zip()
}
