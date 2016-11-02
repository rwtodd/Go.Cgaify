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
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"github.com/rwtodd/apputil/cmdline"
)

var gmd = flag.String("m", "CGA1", "graphics mode to use for output")
var hlp = flag.Bool("h", false, "display help text")
var rsize = flag.Uint("rsz", 0, "resize the image by x% instead of video-mode size")
var zc = flag.Uint("zc", 0, "set the zero color of a CGA mode to EGA color N")

// help puts the usage information on Stderr and exits with a non-zero code.
func help() {
	fmt.Fprintln(os.Stderr, "usage: cgaify [-h] [-m MODE] [-rsz PCT] [-zc N] file...")
	fmt.Fprintln(os.Stderr, "\nOptions:")
	fmt.Fprintln(os.Stderr, "\t-h   display this help text")
	fmt.Fprintln(os.Stderr, "\t-m   select target graphics mode (default CGA1)")
	fmt.Fprintln(os.Stderr, "\t-rsz resize by PCT% instead of the video-mode size")
	fmt.Fprintln(os.Stderr, "\t-zc  set color 0 to EGA color N (CGA modes only)")

	fmt.Fprintln(os.Stderr, "\nModes:")
	for k, v := range modes {
		fmt.Fprintf(os.Stderr, "\t%s:\t%s\n", k, v.desc)
	}
	os.Exit(1)
}

// globally track the number of errors we encountered
var errCnt = 0

func disperr(ctxt string, err error) {
	errCnt++
	fmt.Fprintf(os.Stderr, "%s: %s\n", ctxt, err.Error())
}

// resizeImage changes the image dimensions to match the graphics mode,
// or if the -rsz option was used, it resizes it by the given percentage
func resizeImage(i image.Image, gmode *mode) image.Image {
	if *rsize == 100 {
		// just return the image if the size is unchanged
		return i
	}

	srcBounds := i.Bounds()
	newW, newH := gmode.width, gmode.height
	if (float64(srcBounds.Dx()) / float64(srcBounds.Dy())) > gmode.aspectRatio() {
		newH = 0
	} else {
		newW = 0
	}

	// If the user has selected a percentage by which to resize,
	// forget what we just calculated and use the percentage instead.
	if *rsize > 0 {
		newW, newH = uint(float64(srcBounds.Dx())*(float64(*rsize)/100.0)), 0
	}
	return resize.Resize(newW, newH, i, resize.Bicubic)
}

// process takes a filename and a graphics mode, and converts the size and
// colors to match that mode.
func process(fname string, gmode *mode) {
	// STEP ONE: Decode the image...
	srcfile, err := os.Open(fname)
	if err != nil {
		disperr(fname, err)
		return
	}
	srcimg, _, err := image.Decode(srcfile)
	srcfile.Close()
	if err != nil {
		disperr(fname, err)
		return
	}

	// STEP TWO: Resize and Re-Palette the image...
	srcimg = resizeImage(srcimg, gmode)
	srcBounds := srcimg.Bounds()
	outimg := image.NewPaletted(srcBounds, gmode.colors)
	draw.FloydSteinberg.Draw(outimg, srcBounds, srcimg, image.ZP)

	// STEP THREE: Output the new image...
	outfile, err := os.OpenFile(
		filepath.Base(fname)+"_"+*gmd+".gif",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0666)
	if err != nil {
		disperr(fname, err)
		return
	}
	err = gif.Encode(
		outfile,
		outimg,
		&gif.Options{len(gmode.colors), nil, nil})
	outfile.Close()
	if err != nil {
		disperr(fname, err)
		return
	}

}

func main() {
	cmdline.GlobArgs()
	flag.Parse()

	gmode, ok := modes[strings.ToUpper(*gmd)]
	if !ok || *hlp || (*zc > 15) || len(flag.Args()) == 0 {
		help()
	}

	// set the zero color in CGA modes
	if len(gmode.colors) == 4 {
		gmode.colors[0] = egacolors[int(*zc)]
	}

	for _, fname := range flag.Args() {
		process(fname, gmode)
	}

	if errCnt > 0 {
		fmt.Fprintf(os.Stderr, "\nThere were %d errors.\n", errCnt)
		os.Exit(1)
	}
}
