package main

import (
  "flag"
)

func main() {
  filename := flag.String("in", "", "The name of the file to load.")
  flag.Parse()

  imageSet := NewImageSet(*filename)
  imageSet.Process()

}
