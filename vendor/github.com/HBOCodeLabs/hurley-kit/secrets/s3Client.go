package secrets

import (
	"errors"
	"io/ioutil"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// s3SvcValue atomically stores and retrieves the value of
// s3 service client
var s3SvcValue atomic.Value

// sessionValue atomically stores and retrieves aws session object
var sessionValue atomic.Value

// s3SvcCreationMutex help to ensure only 1 instance of
// aws session and s3 service are created
var s3SvcCreationMutex sync.Mutex

// s3ObjectGetter provides an API to return s3 object when
// called by region, bucket and key
type s3ObjectGetter interface {
	getObject(region, bucket, key string) ([]byte, error)
}

// getS3Service will return pointer to singleton of S3 service.
// A singleton of session will created inside this function, which
// is used to create the S3 service object.
//
// According to https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
//     Sessions should be cached when possible, because creating a new
//     Session will load all configuration values from the environment,
//     and config files each time the Session is created.
func getS3Service(region string) (*s3.S3, error) {
	if s3SvcStored := s3SvcValue.Load(); s3SvcStored != nil {
		return s3SvcStored.(*s3.S3), nil
	}

	s3SvcCreationMutex.Lock()
	defer s3SvcCreationMutex.Unlock()

	var s3Svc *s3.S3
	if s3SvcStored := s3SvcValue.Load(); s3SvcStored == nil {
		var sessionInstance *session.Session
		if sessionStored := sessionValue.Load(); sessionStored == nil {
			var err error
			sessionInstance, err = session.NewSession(&aws.Config{
				Region: aws.String(region),
			})

			if err != nil {
				return nil, err
			}

			sessionValue.Store(sessionInstance)
		} else {
			sessionInstance = sessionStored.(*session.Session)
		}

		s3Svc = s3.New(sessionInstance)
		s3SvcValue.Store(s3Svc)
	} else {
		s3Svc = s3SvcStored.(*s3.S3)
	}

	return s3Svc, nil
}

// s3Client provides an API to return s3 object when
// called by region, bucket and key. It also takes care
// of creating singleton of aws session.
type s3Client struct{}

func (s *s3Client) getObject(region, bucket, key string) ([]byte, error) {
	s3Svc, err := getS3Service(region)
	if err != nil {
		return nil, err
	}

	if s3Svc == nil {
		return nil, errors.New("s3Svc is nil")
	}

	resp, err := s3Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
