package main

import (
	"fmt"
	"os"
)

func main() {
   outfl,_ := os.OpenFile("vgapal.go", os.O_WRONLY|os.O_CREATE, 0666)
  
   fmt.Fprintln(outfl, `package main 

// this file is generated via code 

import "image/color"

var vgacolors = color.Palette{`)

   for r := 0; r < 8; r++ {
     red := int(0.5+float64(r)*255.0/7.0)
     for g := 0; g < 8; g++ {
       green := int(0.5+float64(g)*255.0/7.0)
       for b := 0; b < 4; b ++ {
       		blue := int(0.5+float64(b)*255.0/3.0)
		fmt.Fprintf(outfl, "\tcolor.RGBA{0x%02x, 0x%02x, 0x%02x, 0xFF},\n",red,green,blue)
       }
     }
   }
   fmt.Fprintln(outfl, "}\n")
   outfl.Close()
}

