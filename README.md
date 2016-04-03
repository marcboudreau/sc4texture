# SC4Texture

The sc4texture application processes a PNG image containing base textures.  The image is partitioned into 128 by 128 pixel images.  Each image is compared to previously processed images, in order to avoid duplicates.  Those unique images are written to a [.dat file](http://www.wiki.sc4devotion.com/index.php?title=DBPF) along with an HTML report file.

## Usage

This application takes three inputs:
- A pathname that refers to a file containing the PNG image to process
- A pathname that refers to the .dat file to produce/update
- A starting texture Instance ID value (a 32-bit hexadecimal value)

The input PNG image is partitioned in 128 x 128 pixel sub-images, starting at the top-left corner.  If the dimensions of the image are not exact multiples of 128, the partial images along the right edge and/or bottom are not processed.

If the specified output .dat file doesn't exist, a new one is created, otherwise the texture images are added to the existing.

The starting texture Instance ID value is used to assign Instance ID values to each unique texture image extracted from the input image.  The application handles creating the five zoom levels: 128x128, 64x64, 32x32, 16x16, and 8x8;  assigning an Instance ID value for each.  When updating an existing .dat file, if a texture is assigned an Instance ID that's already in use in the file, the application will overwrite that entry with the extracted texture image.  This behaviour allows rebuilding the .dat file repeatedly.
