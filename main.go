package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/nfnt/resize"
	"github.com/rwtodd/apputil-go/cmdline"
)

var gmd = flag.String("m", "CGA1", "graphics mode to use for output")
var hlp = flag.Bool("h", false, "display help text")

func help() {
	fmt.Fprintln(os.Stderr, "usage: cgaify [-m MODE] [-h] file...")
	fmt.Fprintln(os.Stderr, "\nModes:")
	for k, v := range modes {
		fmt.Fprintf(os.Stderr, "\t%s:\t%s\n", k, v.desc)
	}
	os.Exit(1)
}

var errCnt = 0

func disperr(ctxt string, err error) {
	errCnt++
	fmt.Fprintf(os.Stderr, "%s: %s\n", ctxt, err.Error())
}

func main() {
	cmdline.GlobArgs()
	flag.Parse()

	gmode, ok := modes[*gmd]
	if !ok || *hlp {
		help()
	}

	for _, fname := range flag.Args() {
		srcfile, err := os.Open(fname)
		if err != nil {
			disperr(fname, err)
			continue
		}

		srcimg, _, err := image.Decode(srcfile)
		srcfile.Close()
		if err != nil {
			disperr(fname, err)
			continue
		}

		// resize image ...
		srcBounds := srcimg.Bounds()
		newW, newH := gmode.width, gmode.height
		fmt.Printf("SRC %f %f == %f\n", float64(srcBounds.Dx()), float64(srcBounds.Dy()),
			(float64(srcBounds.Dx()) / float64(srcBounds.Dy())) )
		fmt.Printf("TGT %f\n", gmode.aspectRatio()) 

		if (float64(srcBounds.Dx()) / float64(srcBounds.Dy())) > gmode.aspectRatio() {
			newH = 0
		} else {
			newW = 0
		}
		fmt.Printf("W, H = %d %d\n", newW, newH) 
		srcimg = resize.Resize(newW, newH, srcimg, resize.Bicubic)

		srcBounds = srcimg.Bounds()
		outimg := image.NewPaletted(srcBounds, gmode.colors)
		draw.FloydSteinberg.Draw(outimg, srcBounds, srcimg, image.ZP)

		outfile, err := os.OpenFile(fname+"_"+*gmd+".gif", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			disperr(fname, err)
			continue
		}
		err = gif.Encode(outfile, outimg, &gif.Options{len(gmode.colors), nil, nil})
		outfile.Close()
		if err != nil {
			disperr(fname, err)
			continue
		}
	}

	if errCnt > 0 {
		fmt.Fprintf(os.Stderr, "\nThere were %d errors.\n", errCnt)
		os.Exit(1)
	}
}
