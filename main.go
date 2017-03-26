package main

import (
	"bitbucket.org/balpha/go-unicornify/unicornify"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	var mail, hash string
	var random, free, zoomOut, nodouble, shading, grass, serial bool
	var size int
	var outfile string

	flag.StringVar(&mail, "m", "", "the email address for which a unicorn avatar should be generated")
	flag.StringVar(&hash, "h", "", "the hash for which a unicorn avatar should be generated")
	flag.BoolVar(&random, "r", false, "generate a random unicorn avatar")
	flag.IntVar(&size, "s", 256, "the size of the generated unicorn avatar in pixels (in either direction)")
	flag.StringVar(&outfile, "o", "", "filename of the output PNG image, defaults to {hash}.png")
	flag.BoolVar(&free, "f", false, "generate a free unicorn avatar, i.e. with a transparent background")
	flag.BoolVar(&zoomOut, "z", false, "zoom out, so the unicorn is fully visible")
	flag.BoolVar(&nodouble, "noaa", false, "no antialiasing")
	flag.BoolVar(&shading, "shading", false, "add shading that gives the unicorns more depth")
	flag.BoolVar(&grass, "grass", false, "add grass to the ground")
	flag.BoolVar(&serial, "serial", false, "do not parallelize the drawing")

	flag.Parse()
	inputs := 0
	if mail != "" {
		inputs++
	}
	if hash != "" {
		inputs++
	}
	if random {
		inputs++
	}
	if inputs == 0 {
		os.Stderr.WriteString("Must specify an email (via -m), a hash (via -h), or random generation (via -r)\n")
		os.Exit(1)
	}
	if inputs > 1 {
		os.Stderr.WriteString("Cannot specify more than one of -m, -h, and -r.\n")
		os.Exit(1)
	}
	if size <= 0 {
		os.Stderr.WriteString("Size (argument to -s) must be a positive number")
		os.Exit(1)
	}

	if random {
		hash = randomHash()
	} else if mail != "" {
		hash = mail2hash(mail)
	}
	if outfile == "" {
		outfile = hash + ".png"
	}

	fmt.Printf("Creating size %v avatar for hash %v, writing into %v\n", size, hash, outfile)
	actualSize := size
	if !nodouble {
		actualSize *= 2
	}

	yCallback := func(y int) {
		perc := (y + 1) * 100 / actualSize
		fmt.Printf("\r%v%%    ", perc)
	}

	err, img := unicornify.MakeAvatar(hash, actualSize, !free, zoomOut, shading, grass, !serial, yCallback)
	fmt.Print("\r    \r")
	if err != nil {
		os.Stderr.WriteString("Not a valid hexadecimal number: " + hash + "\n")
		os.Exit(1)
	}

	if !nodouble {
		img = downscale(img)
	}

	f, err := os.Create(outfile)
	if err != nil {
		os.Stderr.WriteString("Could not create output file " + outfile + "\n")
		os.Exit(1)
	}

	defer f.Close()
	buf := bufio.NewWriter(f)
	png.Encode(buf, img)
	err = buf.Flush()
	if err != nil {
		os.Stderr.WriteString("Error writing to output file\n")
		os.Exit(1)
	}
}

func mail2hash(mail string) string {
	mail = strings.ToLower(strings.TrimSpace(mail))
	mailbytes := make([]byte, len(mail))
	copy(mailbytes, mail)
	md5sum := md5.Sum(mailbytes)
	return hex.EncodeToString(md5sum[:])
}

func randomHash() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 16)
	for i := byte(0); i < 16; i += 4 {
		n := r.Uint32()
		for j := byte(0); j < 4; j++ {
			b[i+j] = byte((n >> (8 * j)) & 255)
		}
	}
	return hex.EncodeToString(b)
}

func downscale(img *image.RGBA) *image.RGBA {
	origsize := img.Bounds().Dx()

	result := image.NewRGBA(image.Rect(0, 0, origsize/2, origsize/2))

	inpix := img.Pix
	outpix := result.Pix

	for y := 0; y < origsize/2; y++ {
		for x := 0; x < origsize*2; x++ {
			inpos := (x/4)*8 + x%4 + y*2*img.Stride
			v := uint32(inpix[inpos]) + uint32(inpix[inpos+4]) + uint32(inpix[inpos+img.Stride]) + uint32(inpix[inpos+img.Stride+4])
			outpix[x+y*result.Stride] = uint8(v / 4)
		}
	}
	return result

}
