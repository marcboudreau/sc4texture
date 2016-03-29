package main

import (
  "flag"
  "os"
  "fmt"
  "image/png"
)

func main() {
  filename := flag.String("in", "", "The name of the file to load.")
  flag.Parse()

  file, e := os.Open(*filename)
  if e != nil {
    fmt.Printf("Failed to open file %s. Error: %s\n", *filename, e)
  }
  defer file.Close()

  image, e := png.Decode(file)
  if e != nil {
    fmt.Printf("Failed to decode PNG image. Error: %s\n", e)
  }

  fmt.Printf("Image details: %s\n", image.Bounds().String())
}
