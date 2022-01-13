package awsvm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type VM struct {
	KeyPairName       string
	DockerComposePath string
	PoolPath          string
	RunnerEnvPath     string

	// vm instance data
	Image         string
	InstanceType  string
	Subnet        string
	Groups        []string
	AllocPublicIP bool
	Device        string
	VolumeType    string
	VolumeSize    int64
	VolumeIops    int64
	Tags          map[string]string
}

func (vm *VM) Create(client *ec2.EC2) error {
	userData, err := getUserData(vm.DockerComposePath, vm.PoolPath, vm.RunnerEnvPath)
	if err != nil {
		return err
	}

	tags := createCopy(vm.Tags)
	tags["Name"] = "harness-cie-delegate"

	in := &ec2.RunInstancesInput{
		ImageId:      aws.String(vm.Image),
		InstanceType: aws.String(vm.InstanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		UserData: aws.String(
			base64.StdEncoding.EncodeToString(
				[]byte(userData),
			),
		),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags:         convertTags(tags),
			},
		},
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			{
				AssociatePublicIpAddress: aws.Bool(vm.AllocPublicIP),
				DeviceIndex:              aws.Int64(0),
				SubnetId:                 aws.String(vm.Subnet),
				Groups:                   aws.StringSlice(vm.Groups),
			},
		},
	}
	if vm.KeyPairName != "" {
		in.KeyName = aws.String(vm.KeyPairName)
	}

	fmt.Printf("spec ==== %v \n", in)
	ret, err := client.RunInstancesWithContext(context.Background(), in)
	if err != nil {
		logrus.WithError(err).
			Errorln("aws: [provision] failed to create VM")
		return err
	}
	logrus.WithField("ret", ret).Infoln("created the vm")
	return nil
}

func getUserData(dockerComposePath, poolPath, runnerEnvPath string) (string, error) {
	composeData, err := getB64EncodedFile(dockerComposePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode docker compose file")
	}

	poolData, err := getB64EncodedFile(poolPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode pool file")
	}

	envData, err := getB64EncodedFile(runnerEnvPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode .env file")
	}
	return fmt.Sprintf(`
apt:
sources:
	docker.list:
	source: deb [arch=amd64] https://download.docker.com/linux/ubuntu $RELEASE stable
	keyid: 9DC858229FC7DD38854AE2D88D81803C0EBFCD88
packages:
- docker-ce
write_files:
- path: /runner/docker-compose.yml
permissions: '0600'
encoding: b64
content: %s
- path: /runner/.drone_pool.yml
permissions: '0600'
encoding: b64
content: %s
- path: /runner/.env
permissions: '0600'
encoding: b64
content: %s
runcmd:
- sudo curl -L "https://github.com/docker/compose/releases/download/2.2.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
- sudo chmod +x /usr/local/bin/docker-compose
- mkdir -p /runner
- cd /runner
- ssh-keygen -f id_rsa -q -P ""
- sudo docker-compose up -d`, composeData, poolData, envData), nil
}

func getB64EncodedFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// helper function creates a copy of map[string]string
func createCopy(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

// helper function converts an array of tags in string
// format to an array of ec2 tags.
func convertTags(in map[string]string) []*ec2.Tag {
	var out []*ec2.Tag
	for k, v := range in {
		out = append(out, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return out
}
