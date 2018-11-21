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
	for _, arg := range flag.Args() {
		wg.Add(1)
		go resizeImage(arg, *w, *h)
	}
	wg.Wait()
}

func resizeImage(arg string, w, h uint) {
	defer wg.Done()
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
	if err := jpeg.Encode(out, m, nil); err != nil {
		log.Print(err)
		return
	}
}
