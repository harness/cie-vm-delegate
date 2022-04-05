# cie-vm-delegate
Easy installation of docker delegate for CIE AWS VM feature via terraform. With this script, it creates a terraform file which can be used to create the delegate vms for CIE feature.

Before following the guide, please verify **go** and **terraform** are installed in your env 

# Pre-requisites:
1. Set up an "access key ID" and "access key secret" [aws secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_CreateAccessKey) to be used during the configuration of the runner for communication with AWS. <br />
   (OR) create an [IAM role](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#create-iam-role) with EC2AdminstratorFullAccess policy.
3. Setup up vpc firewall rules for the build instances [ec2 authorizing-access-to-an-instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/authorizing-access-to-an-instance.html). We need to allow ingress to port 9079. (Optional) RDP port 3389 on windows and ssh port 22 on linux can be opened for debugging purpose. Once complete you will have a security group id, which is needed for configuration of the runner.


# Steps:
1. Clone this git repo 
2. Edit file `config/.env` and provide values for the following keys in the file:
    * DRONE_SETTINGS_AWS_REGION
    * DRONE_SETTINGS_AWS_ACCESS_KEY_ID
    * DRONE_SETTINGS_AWS_ACCESS_KEY_SECRET
    * DRONE_SETTINGS_AWS_KEY_PAIR_NAME
    * DELEGATE_IAM_PROFILE_NAME - Set this field with instance profile name to use IAM role instead of access/secret keys. <br /> <br />
This file is used by the drone aws runner to connect to AWS. Use [AWS EC2 environment variables](https://docs.drone.io/runner/aws/installation/#aws-ec2-environment-variables) to reference attributes present in .env file. <br /> 
**Note**: Either one of DELEGATE_IAM_PROFILE_NAME or (DRONE_SETTINGS_AWS_ACCESS_KEY_ID & DRONE_SETTINGS_AWS_ACCESS_KEY_SECRET) needs to be set. <br />
3. Update `config/.drone_pool.yml` file. This file is used by the drone aws runner to instantiate a pool of AWS instances which will be used by Hanress CIE builds. This reduces the time for builds by cutting the time it takes to provision a VM. Use [Pool](https://docs.drone.io/runner/aws/configuration/pool/) to reference attributes of .drone_pool.yml file. 
4. Install a docker delegate from harness nextgen UI and copy the docker delegate yaml to use it in the next step.
5. Replace the content of `config/harness-delegate.yml` file with the docker delegate yaml file generated in the prior step.
6. Run: `go run main.go`. This will generate vm.tf file which can be used to create the delegate vm. Execute `terraform apply` to create the vm. 
7. To create the vm directly, execute `CREATE_VM=true go run main.go'. It is preferable to use step 6 since it allows updating terraform file for vm separately.

# Notes:
1. For windows pool, use public ami with name "Microsoft Windows Server 2019 Base with Containers".
2. For linux pool, ubuntu 20.04 is officially supported.
3. To create a custom windows ami that works with windows pool, follow the steps mentioned in "Run Sysprep with EC2Config or EC2Launch" section of  https://aws.amazon.com/premiumsupport/knowledge-center/sysprep-create-install-ec2-windows-amis/

## Troubleshooting:

# CIE build stuck at initialize step on health check
If CIE build is stuck at initialize step on health check connectivity with lite-engine, either lite-engine is not running on build VM or there is a connectivity issue between runner and lite-engine.

a) Verify whether lite-engine is running on build VM
    -> Select a VM in running state from the pool.
    -> SSH/RDP to the VM
    -> Check whether lite-engine process is running on the VM
    -> Lite-engine process is started at VM startup via cloud init script. Analyse cloud init output logs to debug issues related to startup of lite-engine process.

b) To verify whether runner is able to communicate to lite-engine from delegate VM:
    nc -vz <build-vm-ip> 9079
If status is not successful & lite-engine is not running on build VM, then security group is not setup correctly on the vm. Update security group in pool yaml such that runner can communicate with the pool VMs.

# Log location:

Linux:
Lite-engine logs:       /var/log/lite-engine.log
Cloud init output log:  /var/log/cloud-init-output.log

Windows:
Lite-engine logs:       C:\Program Files\lite-engine\log.out
Cloud init output logs: C:\ProgramData\Amazon\EC2-Windows\Launch\Log\UserdataExecution.log