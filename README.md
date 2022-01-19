# cie-vm-delegate
Easy installation of docker delegate for CIE AWS VM feature via terraform. With this script, it creates a terraform file which can be used to create the delegate vms for CIE feature.

Before following the guide, please verify **go** and **terraform** are installed in your env 

# Pre-requisites:
1. Set up an "access key ID" and "access key secret" [aws secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_CreateAccessKey) to be used during the configuration of the runner for communication with AWS.
2. Setup up vpc firewall rules for the build instances ec2 [ec2 authorizing-access-to-an-instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/authorizing-access-to-an-instance.html)
authorizing-access-to-an-instance. We need to allow ingress to port 22 and 9079. (Optional) RDP port 3389 can be opened for debugging purpose. Once complete you will have a security group id, which is needed for configuration of the runner.
3. (windows only instance) You need to add the AdministratorAccess policy to the IAM role associated with the access key and access secret [IAM](https://console.aws.amazon.com/iamv2/home#/users). You will use the instance profile arn iam_profile_arn, in your pipeline.

# Steps:
1. Clone this git repo 
2. Edit file `config/.env` and provide values for the following keys in the file:
    * DRONE_SETTINGS_AWS_REGION
    * DRONE_SETTINGS_AWS_ACCESS_KEY_ID
    * DRONE_SETTINGS_AWS_ACCESS_KEY_SECRET
    * DRONE_SETTINGS_KEY_PAIR_NAME  <br /> 
This file is used by the drone aws runner to connect to AWS. Use [AWS EC2 environment variables](https://docs.drone.io/runner/aws/installation/#aws-ec2-environment-variables) to reference attributes present in .env file.  <br />
4. Update `config/.drone_pool.yml` file. This file is used by the drone aws runner to instantiate a pool of AWS instances which will be used by Hanress CIE builds. This reduces the time for builds by cutting the time it takes to provision a VM. Use [Pool](https://docs.drone.io/runner/aws/configuration/pool/) to reference attributes of .drone_pool.yml file. 
5. Install a docker delegate from harness UI and copy the docker delegate yaml to use it in the next step.  
6. Replace the content of `config/harness-delegate.yml` file with the docker delegate yaml file generated in the prior step.
7. Run: `go run main.go`. This will generate vm.tf file which can be used to create the delegate vm. Execute `terraform apply` to create the vm. 
8. To create the vm directly, execute `CREATE_VM=true go run main.go'. It is preferable to use step 6 since it allows updating terraform file for vm separately.
