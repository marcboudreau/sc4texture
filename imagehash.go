package main

import (
  "hash/fnv"
  "image"
)

func Hash(image image.Image) uint64 {
  hash := fnv.New64a()
  rect := image.Bounds()
  size := (rect.Max.X - rect.Min.X) * (rect.Max.Y - rect.Min.Y) * 16
  data := make([]byte, size)
  pos := 0

  for y := rect.Min.Y; y < rect.Max.Y; y++ {
    for x := rect.Min.X; x < rect.Max.X; x++ {
      color := image.At(x, y)
      r, g, b, a := color.RGBA()

      convertUint32ToBytes(r, data[pos:pos + 4])
      pos += 4
      convertUint32ToBytes(g, data[pos:pos + 4])
      pos += 4
      convertUint32ToBytes(b, data[pos:pos + 4])
      pos += 4
      convertUint32ToBytes(a, data[pos:pos + 4])
    }
  }

  hash.Write(data)
  return hash.Sum64()
}

func convertUint32ToBytes(value uint32, bytes []byte) {
  bytes[0] = byte(value >> 24 & 0xFF)
  bytes[1] = byte(value >> 16 & 0xFF)
  bytes[2] = byte(value >> 8 & 0xFF)
  bytes[3] = byte(value & 0xFF)
}
