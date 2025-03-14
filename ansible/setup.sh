pwd=$(pwd)
absPath=$(realpath "$pwd")
tfDir=$(realpath "$absPath/../tf")
servers=$(realpath "$tfDir/servers.ini")

set -e  # Stops script on first error


# Mount the volume if not already mounted
ansible-playbook -i $servers mount_volume.yml -f 55

# Install necessary package dependencies and repositories on all servers
ansible-playbook -i $servers installation.yml -f 55
