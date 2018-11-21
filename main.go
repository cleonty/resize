package main

import (
	"context"
	"flag"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

var w = flag.Uint("w", 0, "width")
var h = flag.Uint("h", 0, "height")

func main() {
	var wg sync.WaitGroup
	tokens := make(chan struct{}, runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
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
			log.Println("Cancelling...")
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
}

func resizeImage(fileName string, w, h uint) error {
	if !strings.HasSuffix(fileName, ".jpg") && !strings.HasSuffix(fileName, ".jpeg") {
		return fmt.Errorf("jpeg file is required")
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
		return err
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}
	file.Close()
	m := resize.Resize(w, h, img, resize.Lanczos3)
	out, err := os.Create(name + "_resized" + ext)
	if err != nil {
		return err
	}
	defer out.Close()
	if err := jpeg.Encode(out, m, nil); err != nil {
		return err
	}
	return nil
}
