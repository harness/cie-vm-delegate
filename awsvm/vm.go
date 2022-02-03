package awsvm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type VM struct {
	Credentials       Creds
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
	IamProfile    string
}

func (vm *VM) Create() error {
	client, err := vm.Credentials.GetClient()
	if err != nil {
		return err
	}

	userDataB64, err := getB64UserData(vm.DockerComposePath, vm.PoolPath, vm.RunnerEnvPath)
	if err != nil {
		return err
	}

	tags := createCopy(vm.Tags)
	tags["Name"] = "harness-cie-delegate"

	var iamProfile *ec2.IamInstanceProfileSpecification
	if vm.IamProfile != "" {
		iamProfile = &ec2.IamInstanceProfileSpecification{
			Name: aws.String(vm.IamProfile),
		}
	}

	in := &ec2.RunInstancesInput{
		ImageId:            aws.String(vm.Image),
		InstanceType:       aws.String(vm.InstanceType),
		MinCount:           aws.Int64(1),
		MaxCount:           aws.Int64(1),
		UserData:           aws.String(userDataB64),
		IamInstanceProfile: iamProfile,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags:         convertTags(tags),
			},
		},
	}
	if vm.KeyPairName != "" {
		in.KeyName = aws.String(vm.KeyPairName)
	}
	if vm.Subnet != "" {
		in.SubnetId = aws.String(vm.Subnet)
	}

	_, err = client.RunInstancesWithContext(context.Background(), in)
	if err != nil {
		logrus.WithError(err).
			Errorln("aws: [provision] failed to create VM")
		return err
	}
	logrus.Infoln("created the vm")
	return nil
}

func (vm *VM) CreateTF() error {
	userDataB64, err := getB64UserData(vm.DockerComposePath, vm.PoolPath, vm.RunnerEnvPath)
	if err != nil {
		return err
	}

	// create new empty hcl file object
	f := hclwrite.NewEmptyFile()

	// initialize the body of the new file object
	rootBody := f.Body()

	providerBlock := rootBody.AppendNewBlock("provider", []string{"aws"})
	providerBody := providerBlock.Body()
	if vm.Credentials.Region != "" {
		providerBody.SetAttributeValue("region", cty.StringVal(vm.Credentials.Region))
	}
	if vm.Credentials.AccessKey != "" {
		providerBody.SetAttributeValue("access_key", cty.StringVal(vm.Credentials.AccessKey))
	}
	if vm.Credentials.SecretKey != "" {
		providerBody.SetAttributeValue("secret_key", cty.StringVal(vm.Credentials.SecretKey))
	}

	vmBlock := rootBody.AppendNewBlock("resource", []string{"aws_instance", "harness_cie_delegate"})
	vmBody := vmBlock.Body()
	vmBody.SetAttributeValue("ami", cty.StringVal(vm.Image))
	vmBody.SetAttributeValue("instance_type", cty.StringVal(vm.InstanceType))
	vmBody.SetAttributeValue("key_name", cty.StringVal(vm.KeyPairName))
	vmBody.SetAttributeValue("user_data_base64", cty.StringVal(userDataB64))

	if vm.Subnet != "" {
		vmBody.SetAttributeValue("subnet_id", cty.StringVal(vm.Subnet))
	}

	if vm.IamProfile != "" {
		vmBody.SetAttributeValue("iam_instance_profile", cty.StringVal(vm.IamProfile))
	}

	tags := cty.MapVal(map[string]cty.Value{
		"Name": cty.StringVal("harness-cie-delegate"),
	})
	vmBody.SetAttributeValue("tags", tags)

	if err := ioutil.WriteFile("vm.tf", f.Bytes(), 0644); err != nil {
		logrus.WithError(err).Errorln("failed to create vm.tf file")
		return err
	}
	return nil
}

func getB64UserData(dockerComposePath, poolPath, runnerEnvPath string) (string, error) {
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

	userData := fmt.Sprintf(`#cloud-config
# vim: syntax=yaml
#
packages:
  - docker.io

# create the docker group
groups:
  - docker

# Add default auto created user to docker group
system_info:
  default_user:
    groups: [docker]

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
  permissions: '0644'
  encoding: b64
  content: %s
runcmd:
  - set -e
  - [ ls, -l, / ]
  - sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  - sudo chmod +x /usr/local/bin/docker-compose
  - ssh-keygen -f /runner/id_rsa -q -P ""
  - cd /runner
  - sudo docker-compose up -d
`, composeData, poolData, envData)

	return base64.StdEncoding.EncodeToString(
		[]byte(userData),
	), nil
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
