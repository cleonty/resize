package main

import (
	"flag"
	"image/jpeg"
	"log"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

var w = flag.Uint("w", 0, "width")
var h = flag.Uint("h", 0, "height")

func main() {
	ch := make(chan bool)
	flag.Parse()
	for _, arg := range flag.Args() {
		go resizeImage(arg, *w, *h, ch)
	}
	for range flag.Args() {
		<-ch
	}
}

func resizeImage(arg string, w, h uint, done chan<- bool) {
	defer func() {
		done <- true
	}()
	if !strings.HasSuffix(arg, ".jpg") && !strings.HasSuffix(arg, ".jpeg") {
		log.Printf("resize: %q is not a jpeg file", arg)
		return
	}

	var name, ext string
	if strings.HasSuffix(arg, ".jpg") {
		name = strings.TrimSuffix(arg, ".jpg")
		ext = ".jpg"
	}
	if strings.HasSuffix(arg, ".jpeg") {
		name = strings.TrimSuffix(arg, ".jpeg")
		ext = ".jpeg"
	}
	file, err := os.Open(arg)
	if err != nil {
		log.Print(err)
		return
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Print(err)
		return
	}
	file.Close()
	m := resize.Resize(w, h, img, resize.Lanczos3)
	out, err := os.Create(name + "_resized" + ext)
	if err != nil {
		log.Print(err)
		return
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)
}
