package aws

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/mgo.v2/bson"
)

type AwsS3 struct {
	session       *session.Session // 会话对象
	fileDirectory string           // 文件夹
	fileSuffix    string           // 文件夹后缀
	bucket        string           // s3桶
}

// 新建会话对象
func NewAwsS3(secretId, secretKey, region, fileDirectory, fileSuffix, bucket string) (*AwsS3, error) {
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(region), //region
		Credentials: credentials.NewStaticCredentials(
			secretId,  // secret-id
			secretKey, // secret-key
			""),
	})
	if err != nil {
		return nil, err
	}

	return &AwsS3{
		session:       s,
		fileDirectory: fileDirectory,
		fileSuffix:    fileSuffix,
		bucket:        bucket,
	}, nil

}

func (as *AwsS3) UploadFileToS3(file []byte) (string, error) {
	tempFileName := as.fileDirectory + "/" + bson.NewObjectId().Hex() + as.fileSuffix

	_, err := s3.New(as.session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(as.bucket), // bucket名称，把自己创建的bucket名称替换到此处即可
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"), // 只读和
		Body:                 bytes.NewReader(file),
		ContentLength:        aws.Int64(int64(len(file))),
		ContentType:          aws.String("text/plain"),
		ContentDisposition:   aws.String("inline"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	if err != nil {
		return "", err
	}

	return tempFileName, err
}
