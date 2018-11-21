package main

import (
	"flag"
	"image/jpeg"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

var w = flag.Uint("w", 0, "width")
var h = flag.Uint("h", 0, "height")
var wg sync.WaitGroup

func main() {
	flag.Parse()
	for _, fileName := range flag.Args() {
		wg.Add(1)
		go resizeImage(fileName, *w, *h)
	}
	wg.Wait()
}

func resizeImage(fileName string, w, h uint) {
	defer wg.Done()
	if !strings.HasSuffix(fileName, ".jpg") && !strings.HasSuffix(fileName, ".jpeg") {
		log.Printf("resize: %q is not a jpeg file", fileName)
		return
	}

	var name, ext string
	if strings.HasSuffix(fileName, ".jpg") {
		name = strings.TrimSuffix(fileName, ".jpg")
		ext = ".jpg"
	}
	if strings.HasSuffix(fileName, ".jpeg") {
		name = strings.TrimSuffix(fileName, ".jpeg")
		ext = ".jpeg"
	}
	file, err := os.Open(fileName)
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
	if err := jpeg.Encode(out, m, nil); err != nil {
		log.Print(err)
		return
	}
}
