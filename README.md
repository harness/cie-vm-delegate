# cie-vm-delegate
Easy installation of docker delegate for CIE AWS VM feature via terraform. With this script, it creates a terraform file which can be used to create the delegate vms for CIE feature.

# Pre-requisites:
1. Set up an access key and access secret [aws secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_CreateAccessKey) which is needed for configuration of the runner.
2. Provide EC2FullAccess policy to the created IAM user. This will be used by runner to create, update, delete, list and get on vms.
3. (windows only instance) You need to add the AdministratorAccess policy to the IAM role associated with the access key and access secret [IAM](https://console.aws.amazon.com/iamv2/home#/users). You will use the instance profile arn iam_profile_arn, in your pipeline. This permission is required to enable SSH service on windows VM.
4. Setup up vpc firewall rules for the build instances ec2 [ec2 authorizing-access-to-an-instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/authorizing-access-to-an-instance.html)
authorizing-access-to-an-instance. We need to allow ingress to port 22 and 9079. (Optional) RDP port 3389 can be opened for debugging purpose. Once complete you will have a security group id, which is needed for configuration of the runner.
5. Create a EC2 key pair. This is used in step 2.
6. Install golang & terraform on my local system.

# Steps:
1. Clone the repository locally via `git clone https://github.com/harness/cie-vm-delegate.git`.
2. Update `config/.env` file in cloned codebase. DRONE_SETTINGS_AWS_ACCESS_KEY_ID, DRONE_SETTINGS_AWS_ACCESS_KEY_SECRET & DRONE_SETTINGS_KEY_PAIR_NAME attribute values needs to be updated. This file is used by the drone aws runner to connect to AWS. Use [AWS EC2 environment variables](https://docs.drone.io/runner/aws/installation/#aws-ec2-environment-variables) to reference attributes present in .env file.
3. Update `config/.drone_pool.yml` file. This file is used by drone aws runner to instantiate cache of AWS instances which will be used by CIE builds. Use [Pool](https://docs.drone.io/runner/aws/configuration/pool/) to reference attributes of .drone_pool.yml file.
4. Update `config/harness-delegate.yml` file with the docker delegate yaml file generated on adding a new docker delegate in harness.
5. Run: `go run main.go`. This will generate vm.tf file which can be used to create the delegate vm.
6. Alternative to step 5. To create the vm directly, execute `CREATE_VM=true go run main.go`. It is preferable to use step 5 since it allows updating terraform file for vm separately.