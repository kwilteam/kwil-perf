# Scripts to Run the end-to-end peformance tests 

set -e # Exit immediately if a command exits with a non-zero status 

# Adjust the below parameters to tune the load tests 
el=10000
ec=4
nStress=300
blockTime="500ms"
blockTimeNs=500000000
block_sz_bytes=6291456
block_size=6

## Tests to run
## 1. ec = 2,  nStress = 2,  el = 1000000, blockTime = 1s(also with 500ms), block_sz_bytes = 6*1024*1024
## 2. ec = 3,  nStress = 10, el = 100000,  blockTime = 1s(also with 500ms), block_sz_bytes = 6*1024*1024
## 3. ec = 10, nStress = 20, el = 10000,   blockTime = 1s(also with 500ms), block_sz_bytes = 6*1024*1024
## 4. ec = 50, nStress = 24, el = 1000,    blockTime = 1s(also with 500ms), block_sz_bytes = 6*1024*1024

vars="ec=$ec el=$el nStress=$nStress blockTime=$blockTime blockTimeNs=$blockTimeNs block_sz_bytes=$block_sz_bytes block_size=$block_size"

# Setup
ansible-playbook -i ../tf/servers.ini setup.yml -f 15  -e "$vars"

# Prereun
ansible-playbook -i ../tf/servers.ini prerun.yml -f 15  -e "$vars"

# Runload
ansible-playbook -i ../tf/servers.ini runload.yml -f 15 -e "$vars"

# Postrun
ansible-playbook -i ../tf/servers.ini postrun.yml -f 15  -e "$vars"