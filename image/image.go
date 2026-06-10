package image

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/bmp"
)

type ImageType string

const (
	Jpeg ImageType = "jpeg"
	Png  ImageType = "png"
	Bmp  ImageType = "bmp"
)

// 改变图片类型
func ChangeImageByteToJpeg(file []byte, imageType ImageType) ([]byte, error) {
	f := []byte{}
	// 转化报错说明不是图片类型
	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {

		return f, errors.New("不是图片类型")
	}
	buf := bytes.Buffer{}
	switch imageType {
	case Jpeg:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 40})
	case Png:
		err = png.Encode(&buf, img)
	case Bmp:
		err = bmp.Encode(&buf, img)
	}
	if err != nil {
		return nil, errors.New("压缩图片类型有误")
	}
	return buf.Bytes(), nil
}
