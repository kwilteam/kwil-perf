# kwil-perf

This repository contains tooling for running performance tests on the KWIL network.

## Amazon Instance Setup

First step is to setup an amazon instance and install all the necessary dependencies on it before creating an AMI and a template from the instance.

### Setup the filesystem

`df -h` command to view the volumes that are formatted and mounted.

```shell
$ df -h
Filesystem      Size  Used Avail Use% Mounted on
devtmpfs        3.8G   72K  3.8G   1% /dev
tmpfs           3.8G     0  3.8G   0% /dev/shm
/dev/nvme0n1p1  7.9G  1.2G  6.6G  15% /
```

Use the `lsblk` to view any volumes that were mapped at launch but not formatted and mounted.

```shell
$ lsblk
NAME          MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
nvme0n1       259:1    0    8G  0 disk
├─nvme0n1p1   259:2    0    8G  0 part /
└─nvme0n1p128 259:3    0    1M  0 part
nvme1n1       259:0    0 69.9G  0 disk
```

To format and mount an instance store volume that was mapped only, do the following:
Create a file system on the device using the mkfs command.

```shell
sudo mkfs -t ext4 /dev/nvme1n1
```

Create a directory on which to mount the device using the mkdir command.

```shell
sudo mkdir /data
```

Mount the device on the newly created directory using the mount command.

```shell
sudo mount /dev/nvme1n1 /data
```

Get the UUID of the /data mount point using the lsblk command.

```shell
sudo lsblk  -fs
```

Add the following line to the /etc/fstab file to mount the device on boot.

```shell
UUID=31330301-850f-4a51-8b4e-ffa5bb1c5b82 /data ext4 defaults,nofail 0 2
```

Create the required directories and give the ubuntu user ownership

```shell
sudo chown ubuntu:ubuntu /data
sudo chown ubuntu:ubuntu /data/kwil
```

Reboot the server to ensure that the changes take effect.

```shell
sudo reboot
```

Verify that the /data directory is mounted after the server reboots using the `df -h` command.

### Install the required software

Run the following commands to install the required software on the instance.

```shell
wget https://github.com/kwilteam/kwil-perf/scripts/install.sh
chmod +x install.sh
./install.sh
```

### Verify the installation

1. Ensure the presence of the following directories
    - /data/pg
    - /data/bin
    - /data/stats
  
2. Ensure that the following binaries exist in the directory `/data/bin`:
    - kwild
    - kwil-cli
    - kwil-admin
    - stats
    - stress

3. Ensure that the postgres is installed

```shell
psql --version
```

### Create AMI and Template

Ensure that the security groups are configured to allow the traffic corresponding to the ports (26656, 26657, 8080, 8484, 22, 443) used by the KWIL network.


Create an AMI from the instance and then create a template from the AMI. This process must be repeated on different zones if you want to deploy instances on different zones.

## Terraform Setup

Terraform can be used to automate the creation of the AWS instances in bulk as shown below.

``` shell
git clone https://github.com/kwilteam/kwil-perf/
cd kwil-perf/terraform
```

`provider.tf`: This file contains the configuration for the instance providers.

`ec2.tf` file contains the configuration for the instances to be created. Update the file with the required values. specifically the `launch_template` and `count` values.

Useful commands:

```shell
# Configure the keys required for setting up the instances. 
aws configure
```

```shell
terraform init

terraform validate

terraform plan

terraform apply
# terraform apply will generate the ips.txt file with the ip addresses of the instances created.
```

## Testnet Configuration

Run the following commands to setup the testnet configuration.

```sh
cd setup
go build

# To get the help on the setup command
./setup -h

# To setup the testnet configuration with 4 validators and 1 non-validator
./setup -vals 4 -nvals 1 -addresses /path/to/ips.txt -dir /path/to/testnet/dir -kwil-admin /path/to/kwil-admin
```

## Ansible configuration

1. copy the ip addresses of the instances to the `inventory` file in the ansible directory (`ansible/servers.ini`).

2. Update the `ansible_ssh_private_key_file` config in the `ansible/servers.ini` file to point to the private key file to login to the instances.

3. configure.yml can be used to do the following:
    - Configure the instances with the required configuration files.
    - Start and Stop the Postgres services
    - Start and Stop the KWIL services
  
Update the configure.yml file with the required values and the ansible script can be run as shown below.

```shell
# --tags can be adjusted to run specific tasks
ansible-playbook -i servers.ini --private-key ../kwil-login.pem --user ubuntu configure.yml -f 20 --tags "stopkwild,cleanup,startpg,kwild,verify"
```

## Running the tests

Currently the tests are to be run manually, complete automation is still a work in progress

We use `stats` and the `stress` binaries to run the tests. `stress` tool is used to generate the traffic and `stats` tool is used to collect the stats.

Few considerations:

- For better throughput, always use a `sentry` node to handle rpc traffic, therefore, run the stress tool on the `sentry` nodes. 
  
```sh
    Usage of ./stress:
    -authcall
            sign our call requests expecting the server to authenticate (private mode)
    -bi duration
            badger kwild with read-only metadata requests at this interval (default -1ns)
    -cb
            concurrent broadcast (do not wait for broadcast result before releasing nonce lock, will cause nonce errors due to goroutine race)
    -chain string
            chain ID to require (default is any)
    -ddi duration
            deploy/drop datasets at this interval (but after drop tx confirms) (default -1ns)
    -ddn int
            immediately drop new dbs at a rate of 1/ddn (disable with <1)
    -ec int
            max concurrent unconfirmed action and procedure executions (to get multiple tx in a block), split between actions and procedures (default 4)
    -ei duration
            initiate non-view action execution at this interval (limited by max concurrency setting) (default 10ms)
    -el int
            content length in an executed post action (default 50000)
    -gw
            gateway provider instead of vanilla provider, need to make sure host is same as gateway's domain
    -host string
            provider's http url, schema is required (default "http://127.0.0.1:8484")
    -key string
            existing key to use instead of generating a new one
    -nc int
            nonce chaos rate (apply nonce jitter every 1/nc times)
    -ne
            don't make intentionally failed txns
    -nodrop
            don't drop the datasets deployed in the deploy/drop program
    -pollint duration
            polling interval when waiting for tx confirmation (default 400ms)
    -q    only print errors
    -run duration
            terminate after running this long (default 30m0s)
    -v    print RPC durations
    -vi duration
            make view action call at this interval (default -1ns)
    -vl
            pseudorandom variable content lengths, on (0,el]
```

1. Run the stats tool on a validator node to collect the stats related to the validator performace, such as blockRate, txRate etc.

```sh
./stats -h
Usage of ./stats:
  -output string
        stats directory to write stats.json and analysis.json files (default ".stats")
  -rpcserver string
        rpc server address to query stats from (default "http://localhost:26657")
```
