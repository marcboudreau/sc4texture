package main

import (
  "fmt"
  "image"
  "image/png"
  "os"

  "github.com/disintegration/imaging"
)

// ImageSet is used to build a unique set of images.  As image.Image instances
// are added to the set they are analysed to determine if another equivalent
// copy already exists in the set.  An equivalent image is one that matches
// exactly, or matches a rotation, a mirrorred, or a rotated mirrorred image.
type ImageSet struct {
  // SourceImagePath is the path to the source image from which 128x128 pixel
  // sub-images are extracted.
  SourceImagePath string

  // protoImages is a map of 64-bit hashes to image.Image instances.  This map
  // contains all of the unique images that were parsed.
  protoImages map[uint64]image.Image

  // images is a slice of uint64 values which are the hashes mapped to the unique
  // images in protoImages.
  images []uint64

  // orientations is a slice of byte values that indicates the orientation of the
  // corresponding image in relation to the prototype image.  The orientations
  // are encoded as follows:
  //
  //  00000mrr
  // m indicates if the image is a mirror copy (flipped horizontally)
  // rr indicates the rotation:
  //    00 means 0 degrees of rotation
  //    01 means 90 degrees of rotation
  //    02 means 180 degrees of rotation
  //    03 means 270 degrees of rotation
  orientations []byte

  // stride consists of the number of images that exist in each row.
  stride int
}

// NewImageSet creates a new image set with the image located at the provided
// file path.
func NewImageSet(imagePath string) *ImageSet {
  return &ImageSet{SourceImagePath: imagePath, protoImages: make(map[uint64]image.Image)}
}

// Process begins the process of loading the source image, partitioning it, and
// determining all of the unique images.  Once complete, it produces a report
// which is written to a file called report.html in the current directory.
func (p *ImageSet) Process() {
  sourceImage := p.getSourceImage()
  if sourceImage == nil {
    return
  }

  bounds := sourceImage.Bounds()
  numX, numY := calculateNumberImages(bounds)

  p.stride = numX
  p.images = make([]uint64, numX * numY)
  p.orientations = make([]byte, numX * numY)

  for y := 0; y < numY; y++ {
    for x := 0; x < numX; x++ {
      bounds := image.Rect(x * 128, y * 128, (x + 1) * 128, (y + 1) * 128)
      p.AddImage(imaging.Crop(sourceImage, bounds), x, y)
    }
  }

  p.WriteImageFiles()
  p.WriteReport(bounds)
}

// calculateNumberImages calculates the number of 128x128 images to extract from
// the source image.
func calculateNumberImages(bounds image.Rectangle) (x, y int) {
  y = int((bounds.Max.Y - bounds.Min.Y) / 128)
  x = calculateStride(bounds)

  return x, y
}

func calculateStride(bounds image.Rectangle) int {
  return int((bounds.Max.X - bounds.Min.X) / 128)
}

// getSourceImage handles loading the source image from the file path set in the
// receiver.
func (p *ImageSet) getSourceImage() image.Image {
  file, e := os.Open(p.SourceImagePath)
  if e != nil {
    if os.IsNotExist(e) {
      fmt.Fprintf(os.Stderr, "The source image file %s does not exist.\n", p.SourceImagePath)
    } else if os.IsPermission(e) {
      fmt.Fprintf(os.Stderr, "The source image file %s is not readable.\n", p.SourceImagePath)
    } else {
      fmt.Fprintf(os.Stderr, "An error occurred while opening the source image file %s.  Error: %s\n", p.SourceImagePath, e)
    }
    return nil
  }

  image, e := png.Decode(file)
  if e != nil {
    fmt.Fprintf(os.Stderr, "An error occurred while loading the source image. Error: %s\n", e)
    return nil
  }
  defer file.Close()

  return image
}

// AddImage examines the provided image and compares it to the images already
// stored in the map of prototype images.  If this image is unique and doesn't
// match any of the prototype images, including rotations or mirror copies, then
// it is added to the map of prototype images.
func (p *ImageSet) AddImage(img image.Image, x, y int) {
  index := y * p.stride + x
  images := make([]image.Image, 8)
  images[0] = img
  images[1] = imaging.Rotate90(img)
  images[2] = imaging.Rotate180(img)
  images[3] = imaging.Rotate270(img)
  images[4] = imaging.FlipH(img)
  images[5] = imaging.Rotate90(images[4])
  images[6] = imaging.Rotate180(images[4])
  images[7] = imaging.Rotate270(images[4])

  found := false
  firstHash := uint64(0)
  for i := 0; i < 8; i++ {
    hash := Hash(images[i])
    if i == 0 {
      firstHash = hash
    }

    if _, ok := p.protoImages[hash]; ok {
      found = true
      p.orientations[index] = byte(i)
      p.images[index] = hash
      break
    }
  }

  if !found {
    p.protoImages[firstHash] = img
    p.images[index] = firstHash
    p.orientations[index] = byte(0)
  }
}

// WriteImageFiles creates an images directory and creates files in it for each
// prototype image.
func (p *ImageSet) WriteImageFiles() {
  e := os.Mkdir("images", 0755)
  if e != nil {
    fmt.Fprintf(os.Stderr, "An error occurred while creating the images directory. Error: %s\n", e)
  }

  for k, v := range p.protoImages {
    if file, e := os.Create(fmt.Sprintf("images/%x.png", k)); e == nil {
      png.Encode(file, v)
      file.Close()
    }
  }
}

// GetX determines which grid column the provided index references.
func (p *ImageSet) GetX(index int) int {
  return index % p.stride
}

//GetY determines which grid row the provided index references.
func (p *ImageSet) GetY(index int) int {
  return int(index / p.stride)
}

// WriteReport writes an HTML report of the result of the parsing.
func (p *ImageSet) WriteReport(bounds image.Rectangle) {
  if file, e := os.Create("report.html"); e == nil {
    file.WriteString("<html>\n")
    file.WriteString("\t<head>\n")
    file.WriteString("\t\t<title>SC4 Texture Parse Report</title>\n")
    file.WriteString("\t</head>\n")

    file.WriteString("\t<body>\n")
    file.WriteString("\t\t<h1>Source Image</h1>\n")
    h := bounds.Max.Y - bounds.Min.Y
    w := bounds.Max.X - bounds.Min.X
    file.WriteString("\t\t" + WriteSourceImageTag(p.SourceImagePath, h, w) + "\n")
    file.WriteString("\t\t<table border='1'>\n")
    file.WriteString("\t\t\t<tr>\n")
    file.WriteString("\t\t\t\t<th>Src Img Size</th>\n")
    file.WriteString(fmt.Sprintf("\t\t\t\t<td>%d x %d</td>\n", h, w))
    file.WriteString("\t\t\t</tr>\n")
    file.WriteString("\t\t\t<tr>\n")
    file.WriteString("\t\t\t\t<th>Texture Cells</th>\n")
    file.WriteString(fmt.Sprintf("\t\t\t\t<td>%d</td>\n", len(p.images)))
    file.WriteString("\t\t\t</tr>\n")
    file.WriteString("\t\t\t<tr>\n")
    file.WriteString("\t\t\t\t<th>Unique Textures</th>\n")
    file.WriteString(fmt.Sprintf("\t\t\t\t<td>%d</td>\n", len(p.protoImages)))
    file.WriteString("\t\t\t</tr>\n")
    file.WriteString("\t\t</table>\n")

    file.WriteString("\t\t<h1>Output Images</h1>\n")
    file.WriteString("\t\t<table border='1'>\n")
    file.WriteString("\t\t\t<tr>\n")
    file.WriteString("\t\t\t\t<th>X</th>\n")
    file.WriteString("\t\t\t\t<th>Y</th>\n")
    file.WriteString("\t\t\t\t<th>Image</th>\n")
    file.WriteString("\t\t\t\t<th>Hash</th>\n")
    file.WriteString("\t\t\t\t<th>Orientation</th>\n")
    file.WriteString("\t\t\t</tr>\n")

    for i, v := range p.images {
      file.WriteString("\t\t\t<tr>\n")
      file.WriteString(fmt.Sprintf("\t\t\t\t<td>%d</td>\n", p.GetX(i)))
      file.WriteString(fmt.Sprintf("\t\t\t\t<td>%d</td>\n", p.GetY(i)))
      file.WriteString(fmt.Sprintf("\t\t\t\t<td><a href='images/%x.png'><img src='images/%x.png' height='32' width='32'/></a>\n",v , v))
      file.WriteString(fmt.Sprintf("\t\t\t\t<td>%x</td>\n", v))
      file.WriteString(fmt.Sprintf("\t\t\t\t<td>%s</td>\n", GetOrientationLabel(p.orientations[i])))
      file.WriteString("\t\t\t</tr>\n")
    }

    file.WriteString("\t\t</table>\n")
    file.WriteString("\t</body>\n")
    file.WriteString("</html>\n")

    file.Close()
  } else {
    fmt.Fprintf(os.Stderr, "An error occurred while creating the HTML report file report.html\n.")
  }
}

func GetOrientationLabel(orientation byte) string {
  switch orientation & 0x7 {
  case 0:
    return "Standard"
  case 1:
    return "Rotated 90 degrees"
  case 2:
    return "Rotated 180 degrees"
  case 3:
    return "Rotated 270 degrees"
  case 4:
    return "Mirrored"
  case 5:
    return "Mirrored + Rotated 90 degrees"
  case 6:
    return "Mirrored + Rotated 180 degrees"
  case 7:
    return "Mirrored + Rotated 270 degrees"
  }

  return ""
}

// WriteSoureImageTag writes the appropriate <img> tag for the provided image.Image
// instance.
func WriteSourceImageTag(imagePath string, h, w int) string {
  factor := float64(0)
  if h > w {
    factor = float64(640) / float64(h)
  } else {
    factor = float64(640) / float64(w)
  }

  return fmt.Sprintf("<a href='%s'><img src='%s' alt='Source Image' height='%d' width='%d'/></a>", imagePath, imagePath, int(float64(h) * factor), int(float64(w) * factor))
}
