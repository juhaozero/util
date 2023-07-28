package image

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
)

func ChangeImageByte(file []byte) ([]byte, error) {
	f := []byte{}
	// 转化报错说明不是图片类型
	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {

		return f, errors.New("不是图片类型")
	}
	buf := bytes.Buffer{}
	// 压缩成新的
	if err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 40}); err != nil {
		return f, errors.New("压缩图片类型有误")
	}
	f = buf.Bytes()

	return f, nil
}
