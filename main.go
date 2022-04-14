package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/harness/cie-vm-delegate/awsvm"
	"github.com/harness/cie-vm-delegate/compose"
)

var (
	version = "0.0.0"
	build   = "0"
)

func main() {
	if err := godotenv.Load("config/.env"); err != nil {
		logrus.Fatalln(err)
	}

	app := cli.NewApp()
	app.Name = "CIE VM delegate installer"
	app.Usage = "CIE VM delegate installer"
	app.Version = fmt.Sprintf("%s+%s", version, build)
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "DELEGATE_AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "DELEGATE_AWS_ACCESS_KEY_SECRET",
		},
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run for debug purposes",
			EnvVar: "DRY_RUN",
		},
		cli.StringFlag{
			Name:   "delegate-ami",
			Usage:  "AMI for the delegate VM",
			EnvVar: "DELEGATE_AMI",
			Value:  "ami-03a0c45ebc70f98ea",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-2",
			EnvVar: "DELEGATE_REGION",
		},
		cli.StringFlag{
			Name:   "availability-zone",
			Usage:  "aws availability zone",
			Value:  "us-east-2c",
			EnvVar: "DELEGATE_AZ",
		},
		cli.StringFlag{
			Name:   "key-name",
			Usage:  "aws key pair name",
			Value:  "",
			EnvVar: "DELEGATE_KEY_PAIR_NAME",
		},
		cli.StringFlag{
			Name:   "delegate-subnet",
			Usage:  "Subnet for the delegate vm",
			EnvVar: "DELEGATE_SUBNET",
		},
		cli.StringFlag{
			Name:   "delegate-iam-profile",
			Usage:  "IAM profile name for the delegate vm",
			EnvVar: "DELEGATE_IAM_PROFILE_NAME",
		},
		cli.StringSliceFlag{
			Name:   "delegate-security-groups",
			Usage:  "Security groups for delegate vm",
			EnvVar: "DELEGATE_SECURITY_GROUPS",
		},
		cli.StringFlag{
			Name:  "env-file",
			Usage: "source env file",
			Value: "config/.env",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if err := compose.Create(); err != nil {
		return err
	}

	cred := awsvm.Creds{
		AccessKey:        c.String("access-key"),
		SecretKey:        c.String("secret-key"),
		Region:           c.String("region"),
		AvailabilityZone: c.String("availability-zone"),
	}

	vm := &awsvm.VM{
		Credentials:       cred,
		KeyPairName:       c.String("key-name"),
		DockerComposePath: "docker-compose.yml",
		PoolPath:          "config/pool.yml",
		RunnerEnvPath:     "config/.env",

		Image:        c.String("delegate-ami"),
		Subnet:       c.String("delegate-subnet"),
		IamProfile:   c.String("delegate-iam-profile"),
		InstanceType: "t2.medium",
		Groups:       c.StringSlice("delegate-security-groups"),
	}
	return vm.CreateTF()
}
