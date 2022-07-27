package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/drone-runners/drone-runner-aws/command/config"
	"github.com/drone-runners/drone-runner-aws/types"
	"gopkg.in/yaml.v2"
)

type (
	poolDefinition struct {
		Name        string   `json:"name,omitempty"`
		MinPoolSize int      `json:"min_pool_size,omitempty" yaml:"min_pool_size"`
		MaxPoolSize int      `json:"max_pool_size,omitempty" yaml:"max_pool_size"`
		InitScript  string   `json:"init_script,omitempty" yaml:"init_script"`
		Platform    platform `json:"platform,omitempty"`
		Account     account  `json:"account,omitempty"`
		Instance    instance `json:"instance,omitempty"`
	}

	// account provides account settings
	account struct {
		AccessKeyID     string `json:"access_key_id,omitempty"  yaml:"access_key_id"`
		AccessKeySecret string `json:"access_key_secret,omitempty" yaml:"access_key_secret"`
		Region          string `json:"region,omitempty"`
	}

	platform struct {
		OS      string `json:"os,omitempty"`
		Arch    string `json:"arch,omitempty"`
		Variant string `json:"variant,omitempty"`
		Version string `json:"version,omitempty"`
	}

	// instance provides instance settings.
	instance struct {
		AMI           string            `json:"ami,omitempty"`
		Tags          map[string]string `json:"tags,omitempty"`
		IAMProfileARN string            `json:"iam_profile_arn,omitempty" yaml:"iam_profile_arn"`
		Type          string            `json:"type,omitempty"`
		User          string            `json:"user,omitempty"`
		PrivateKey    string            `json:"private_key,omitempty" yaml:"private_key"`
		PublicKey     string            `json:"public_key,omitempty" yaml:"public_key"`
		UserData      string            `json:"user_data,omitempty"`
		Disk          disk              `json:"disk,omitempty"`
		Network       network           `json:"network,omitempty"`
		Device        device            `json:"device,omitempty"`
		ID            string            `json:"id,omitempty"`
		IP            string            `json:"ip,omitempty"`
	}

	// network provides network settings.
	network struct {
		VPC               string   `json:"vpc,omitempty"`
		VPCSecurityGroups []string `json:"vpc_security_group_ids,omitempty" yaml:"vpc_security_groups"`
		SecurityGroups    []string `json:"security_groups,omitempty" yaml:"security_groups"`
		SubnetID          string   `json:"subnet_id,omitempty" yaml:"subnet_id"`
		PrivateIP         bool     `json:"private_ip,omitempty" yaml:"private_ip"`
	}

	// disk provides disk size and type.
	disk struct {
		Size int64  `json:"size,omitempty"`
		Type string `json:"type,omitempty"`
		Iops int64  `json:"iops,omitempty"`
	}

	// device provides the device settings.
	device struct {
		Name string `json:"name,omitempty"`
	}
)

func ProcessPoolFile(rawFile string) error {
	rawPool, err := os.ReadFile(rawFile)
	if err != nil {
		err = fmt.Errorf("unable to read file %s: %w", rawFile, err)
		return err
	}

	poolDef := config.PoolFile{
		Instances: []config.Instance{},
	}

	buf := bytes.NewBuffer(rawPool)
	dec := yaml.NewDecoder(buf)

	for {
		oldPoolDef := new(poolDefinition)
		err := dec.Decode(oldPoolDef)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		poolDef.Instances = append(poolDef.Instances, convert(*oldPoolDef))
	}
	fmt.Println(poolDef)

	file, err := os.OpenFile("update_pool.yaml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("error opening/creating file: %v", err)
	}
	defer file.Close()

	enc := yaml.NewEncoder(file)
	if err = enc.Encode(poolDef); err != nil {
		log.Fatalf("error encoding: %v", err)
	}
	return nil
}

func convert(p poolDefinition) config.Instance {
	return config.Instance{
		Name:  p.Name,
		Type:  "amazon",
		Pool:  p.MinPoolSize,
		Limit: p.MaxPoolSize,
		Platform: types.Platform{
			OS:      p.Platform.OS,
			Arch:    p.Platform.Arch,
			Variant: p.Platform.Variant,
			Version: p.Platform.Version,
		},
		Spec: config.Amazon{
			AMI: p.Instance.AMI,
			Account: config.AmazonAccount{
				Region:      "",
				KeyPairName: "",
			},
			Size: p.Instance.Type,
			Network: config.AmazonNetwork{
				VPC:               p.Instance.Network.VPC,
				VPCSecurityGroups: p.Instance.Network.VPCSecurityGroups,
				SecurityGroups:    p.Instance.Network.SecurityGroups,
				SubnetID:          p.Instance.Network.SubnetID,
				PrivateIP:         p.Instance.Network.PrivateIP,
			},
			IamProfileArn: p.Instance.IAMProfileARN,
			Tags:          p.Instance.Tags,
			DeviceName:    p.Instance.Device.Name,
			UserData:      p.Instance.UserData,
		},
	}
}

func main() {
	ProcessPoolFile("pool.yml")
}
