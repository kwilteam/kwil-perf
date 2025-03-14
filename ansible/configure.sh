#!/bin/bash
PWD=$(pwd)  # Get current working directory
DIR="../kwil"  # Relative path

# Check if the directory exists, if not, create it
if [[ ! -d "$DIR" ]]; then
    echo "Directory does not exist. Creating: $DIR"
    mkdir -p "$DIR"
else
    echo "Directory already exists: $DIR"
fi

# Get absolute path
AbsPath=$(realpath "$PWD/$DIR")


# Read IPs from the file, remove extra spaces, and join with commas
IPS=$(tr '\n' ',' < ../tf/ips.txt | sed 's/,$//')

# Generate the kwil-db binary
git clone https://github.com/kwilteam/kwil-db.git
cd kwil-db
cp go.work.example go.work
task build
cd ..

mkdir bin
cp kwil-db/.build/kwild bin/

# Run the command with formatted IPs
../bin/kwild setup testnet -v 10 -n 5 -o $AbsPath -H "$IPS" --db-owner 0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7
