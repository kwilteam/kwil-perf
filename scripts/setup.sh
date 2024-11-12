# Install software dependencies required for running docker, postgres etc.

sudo snap install go --classic
sudo snap install task --classic
sudo apt-get update
sudo apt-get install build-essential
sudo apt install postgresql -y postgresql-contrib
sudo service postgresql status
sudo systemctl enable postgresql.service
sudo systemctl start postgresql.service

cd /data/
mkdir bin kwil

cd /data/kwil

# Clone the Kwil-DB repository and build the project and copy the binaries to /data/bin
git clone https://github.com/kwilteam/kwil-db.git
cd kwil-db
task build
cp /data/kwil/kwil-db/.build/* /data/bin

# Build the stress tool and copy the binary to /data/bin
cd test/stress
go build
cp stress /data/bin

cd /data/kwil
# Build the stats collector tool and copy the binary to /data/bin
git clone https://github.com/kwilteam/kwil-perf.git
cd kwil-perf/stats
go build
cp stats /data/bin/

