package main

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

const (
	accessKey     = ""
	secretKey     = ""
	bucket        = ""
	downloadPath  = "/tmp"
	resizedWidth  = 263
	resizedHeight = 160
	sufix         = "_thumbnail_lambda"
)

const (
	JPG = "jpg"
	GIF = "gif"
	PNG = "png"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] != "" {
		t := os.Args[1]
		tmp := strings.Split(t, "files/")
		fmt.Println(tmp[1])
		fileName := tmp[1]
		if strings.Index(fileName, sufix) > -1 {
			return
		}

		err := GetOriginFile(fileName)
		if err != nil {
			fmt.Printf("GetOriginFile error %s", err.Error())

		}

		err = ResizeFile(fileName)
		if err != nil {
			fmt.Println(err)
		}

		err = UploadThumbFile(fileName)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("しっぱい")
	}
}

func ResizeFile(fileName string) error {
	f, err := os.Open(fmt.Sprintf("%s/%s", downloadPath, fileName))
	if err != nil {
		return err
	}
	defer f.Close()
	format := GetFormat(f)

	img, err := GetDecodedImage(f, format)
	if err != nil {
		return err
	}
	realWidth := img.Bounds().Size().X
	realHeight := img.Bounds().Size().Y

	var setWidth, setHeight uint
	if realWidth > realHeight {
		setWidth = resizedWidth
		setHeight = 0
	} else {
		setWidth = 0
		setHeight = resizedHeight
	}
	resizedImg := resize.Resize(setWidth, setHeight, img, resize.NearestNeighbor)

	output, err := os.Create(fmt.Sprintf("%s/%s%s", downloadPath, fileName, sufix))
	if err != nil {
		return err
	}
	defer output.Close()
	switch format {
	case JPG:
		jpeg.Encode(output, resizedImg, nil)
	case GIF:
		gif.Encode(output, resizedImg, nil)
	case PNG:
		png.Encode(output, resizedImg)
	}
	return nil
}

func GetDecodedImage(file io.Reader, format string) (img image.Image, err error) {
	switch format {
	case JPG:
		img, err = jpeg.Decode(file)
	case GIF:
		img, err = gif.Decode(file)
	case PNG:
		img, err = png.Decode(file)
	default:
		img = nil
		err = errors.New("other image format")
	}
	return

}

func GetFormat(file *os.File) string {
	bytes := make([]byte, 4)
	n, _ := file.ReadAt(bytes, 0)
	if n < 4 {
		return ""
	}
	if bytes[0] == 0x89 && bytes[1] == 0x50 && bytes[2] == 0x4E && bytes[3] == 0x47 {
		return PNG
	}
	if bytes[0] == 0xFF && bytes[1] == 0xD8 {
		return JPG
	}
	if bytes[0] == 0x47 && bytes[1] == 0x49 && bytes[2] == 0x46 && bytes[3] == 0x38 {
		return GIF
	}
	if bytes[0] == 0x42 && bytes[1] == 0x4D {
		return "bmp"
	}
	return ""
}
