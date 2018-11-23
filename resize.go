package main

import (
	"context"
	"flag"
	"image/jpeg"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

var w = flag.Uint("w", 0, "width")
var h = flag.Uint("h", 0, "height")

func main() {
	var wg sync.WaitGroup
	tokens := make(chan struct{}, runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	flag.Parse()

	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			log.Println("cancelling...")
			cancel()
		case <-ctx.Done():
		}
	}()

	for _, fileName := range flag.Args() {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			select {
			case tokens <- struct{}{}:
				defer func() { <-tokens }()
			case <-ctx.Done():
				log.Printf("resizing %s: cancelled", fileName)
				return
			}
			err := resizeImage(fileName, *w, *h)
			if err != nil {
				log.Printf("resizing %s: %v", fileName, err)
			} else {
				log.Printf("resizing %s: ok", fileName)
			}
		}(fileName)
	}
	wg.Wait()
	log.Println("done")
}

func resizeImage(fileName string, width, height uint) error {
	if !strings.HasSuffix(fileName, ".jpg") && !strings.HasSuffix(fileName, ".jpeg") {
		return errors.Errorf("jpeg file is required, %s given", fileName)
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
		return errors.WithMessage(err, "unable to open source jpeg file "+fileName)
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		return errors.WithMessage(err, "unable to decode source jpeg file "+fileName)
	}
	m := resize.Resize(width, height, img, resize.Lanczos3)
	resizedFileName := name + "_resized" + ext
	out, err := os.Create(resizedFileName)
	if err != nil {
		return errors.WithMessage(err, "unable to create target file "+resizedFileName)
	}
	defer out.Close()
	if err := jpeg.Encode(out, m, nil); err != nil {
		return errors.WithMessage(err, "unable to write into file "+resizedFileName)
	}
	return nil
}
