# Terraform directory
pwd=$(pwd)
absPath=$(realpath "$pwd")
tfDir=$(realpath "$absPath/../tf")
ansibleDir=$(realpath "$absPath/../ansible")
scriptsDir=$(realpath "$absPath/../scripts")

cd $tfDir

set -e  # Stops script on first error

# Initialize Terraform
terraform init


# export TF_VAR_virginia_count=5
# export TF_VAR_california_count=3
# export TF_VAR_frankfurt_count=2
# export TF_VAR_ssh_key_path="~/.ssh/mykey.pem"

# Apply Terraform configuration
terraform apply -auto-approve

# Create node configuration
cd $ansibleDir
./configure.sh

servers=$(realpath "$tfDir/servers.ini")

# sleep for a min or two to let the nodes come up
sleep 120

# Setup the nodes
# Mount the volume if not already mounted
ansible-playbook -i $servers mount_volume.yml -f 55

# Install necessary package dependencies and repositories on all servers
ansible-playbook -i $servers installation.yml -f 55
