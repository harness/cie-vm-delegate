# cie-vm-delegate
Easy installation of docker delegate for CIE AWS VM feature

# Steps:
1. Update config/.env file. This file is used by the drone aws runner to connect to AWS.
2. Update config/.drone_pool.yml file. This file is used by drone aws runner to instantiate cache of AWS instances which will be used by CIE builds. This reduces the time for build completion since vms would already be present in ready state.
3. Update config/harness-delegate.yml file with the docker delegate yaml file generated on adding a new docker delegate in harness.
4. Run: go run main.go 
