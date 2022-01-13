# cie-vm-delegate
Easy installation of docker delegate for CIE AWS VM feature via terraform. With this script, it creates a terraform file which can be used to create the delegate vms for CIE feature.

# Pre-requisites:
1. Set up an access key and access secret [aws secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_CreateAccessKey) which is needed for configuration of the runner.
2. Setup up vpc firewall rules for the build instances ec2 [ec2 authorizing-access-to-an-instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/authorizing-access-to-an-instance.html)
authorizing-access-to-an-instance. We need to allow ingress to port 22 and 9079. (Optional) RDP port 3389 can be opened for debugging purpose. Once complete you will have a security group id, which is needed for configuration of the runner.
3. (windows only instance) You need to add the AdministratorAccess policy to the IAM role associated with the access key and access secret [IAM](https://console.aws.amazon.com/iamv2/home#/users). You will use the instance profile arn iam_profile_arn, in your pipeline.

# Steps:
1. Update `config/.env` file. This file is used by the drone aws runner to connect to AWS. Use [AWS EC2 environment variables](https://docs.drone.io/runner/aws/installation/#aws-ec2-environment-variables) to reference attributes present in .env file.
2. Update `config/.drone_pool.yml` file. This file is used by drone aws runner to instantiate cache of AWS instances which will be used by CIE builds. This reduces the time for build completion since vms would already be present in ready state.
3. Update `config/harness-delegate.yml` file with the docker delegate yaml file generated on adding a new docker delegate in harness.
4. Run: `go run main.go`. This will generate vm.tf file which can be used to create the delegate vm.
5. To create the vm directly, execute `CREATE_VM=true go run main.go`. It is preferable to use step 4 since it allows updating delegate vm with security group, subnet, etc.
