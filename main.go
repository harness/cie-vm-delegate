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
	app := cli.NewApp()
	app.Name = "CIE VM delegate installer"
	app.Usage = "CIE VM delegate installer"
	app.Version = fmt.Sprintf("%s+%s", version, build)
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "DRONE_SETTINGS_AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "DRONE_SETTINGS_AWS_ACCESS_KEY_SECRET",
		},
		cli.StringFlag{
			Name:   "assume-role",
			Usage:  "aws iam role to assume",
			EnvVar: "DRONE_SETTINGS_AWS_ASSUME_ROLE",
		},
		cli.StringFlag{
			Name:   "assume-role-session-name",
			Usage:  "aws iam role session name to assume",
			Value:  "drone-s3",
			EnvVar: "DRONE_SETTINGS_AWS_ASSUME_ROLE_SESSION_NAME",
		},
		cli.StringFlag{
			Name:   "user-role-arn",
			Usage:  "AWS user role",
			EnvVar: "DRONE_SETTINGS_AWS_USER_ROLE_ARN",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-1",
			EnvVar: "DRONE_SETTINGS_AWS_REGION",
		},
		cli.StringFlag{
			Name:   "key-name",
			Usage:  "aws key pair name",
			Value:  "",
			EnvVar: "DRONE_SETTINGS_KEY_PAIR_NAME",
		},
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run for debug purposes",
			EnvVar: "DRY_RUN",
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
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	if err := compose.Create(); err != nil {
		return err
	}

	cred := &awsvm.Creds{
		AccessKey:             c.String("access-key"),
		SecretKey:             c.String("secret-key"),
		AssumeRole:            c.String("assume-role"),
		AssumeRoleSessionName: c.String("assume-role-session-name"),
		UserRoleArn:           c.String("user-role-arn"),
		Region:                c.String("region"),
	}

	client, err := cred.GetClient()
	if err != nil {
		return err
	}

	vm := &awsvm.VM{
		KeyPairName:       c.String("key-name"),
		DockerComposePath: "docker-compose.yml",
		PoolPath:          "config/.drone_pool.yml",
		RunnerEnvPath:     "config/.env",

		// vm instance data
		Image:        "ami-00517afdd8df42285",
		InstanceType: "t2.medim",
	}
	return vm.Create(client)
}
