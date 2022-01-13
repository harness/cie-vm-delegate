package awsvm

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	log "github.com/sirupsen/logrus"
)

type Creds struct {
	AccessKey             string
	SecretKey             string
	AssumeRole            string
	AssumeRoleSessionName string
	UserRoleArn           string
	Region                string
}

func (c *Creds) GetClient() (*ec2.EC2, error) {
	// create the client
	conf := &aws.Config{
		Region: aws.String(c.Region),
	}

	if c.AccessKey != "" && c.SecretKey != "" {
		conf.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
	} else if c.AssumeRole != "" {
		conf.Credentials = getAssumeRoleCred(c.AssumeRole, c.AssumeRoleSessionName)
	} else {
		log.Warn("AWS Key and/or Secret not provided (falling back to ec2 instance profile)")
	}

	sess, err := session.NewSession(conf)
	if err != nil {
		log.WithError(err).Errorln("could not instantiate session")
		return nil, err
	}

	var client *ec2.EC2
	// If user role ARN is set then assume role here
	if len(c.UserRoleArn) > 0 {
		confRoleArn := aws.Config{
			Region:      aws.String(c.Region),
			Credentials: stscreds.NewCredentials(sess, c.UserRoleArn),
		}

		client = ec2.New(sess, &confRoleArn)
	} else {
		client = ec2.New(sess)
	}
	return client, nil
}

func getAssumeRoleCred(roleArn, roleSessionName string) *credentials.Credentials {
	client := sts.New(session.New())
	duration := time.Hour * 1
	stsProvider := &stscreds.AssumeRoleProvider{
		Client:          client,
		Duration:        duration,
		RoleARN:         roleArn,
		RoleSessionName: roleSessionName,
	}

	return credentials.NewCredentials(stsProvider)
}
