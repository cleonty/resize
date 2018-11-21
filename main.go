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
var wg sync.WaitGroup
var tokens chan struct{}

func main() {
	flag.Parse()
	tokens = make(chan struct{}, runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handleInterrupt(cancel)

	for _, fileName := range flag.Args() {
		wg.Add(1)
		go resizeImage(ctx, fileName, *w, *h)
	}
	wg.Wait()
}

func resizeImage(ctx context.Context, fileName string, w, h uint) {
	defer wg.Done()
	select {
	case tokens <- struct{}{}:
		defer func() { <-tokens }()
	case <-ctx.Done():
		return
	}
	log.Printf("resizing %s\n", fileName)
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
	log.Printf("done resizing %s\n", fileName)
}

func handleInterrupt(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("Cancelling...")
	cancel()
}
