#!/bin/bash
set -e

# Constants (adjust as needed)
blockTimeNs=1000000000
block_sz_bytes=$((6*1024*1024))
block_size=6

# Test case parameters
# Test 1: ec=2,  nStress=2,  el=1000000
# Test 2: ec=3,  nStress=10, el=100000
# Test 3: ec=10, nStress=20, el=10000
# Test 4: ec=50, nStress=24, el=1000
declare -a ec_arr=("2" "3" "10" "50")
declare -a nStress_arr=("2" "10" "20" "24")
declare -a el_arr=("1000000" "100000" "10000" "1000")

# Block times to test
declare -a blockTime_arr=("1s" "500ms")

# Loop over each test case
for idx in "${!ec_arr[@]}"; do
    ec="${ec_arr[$idx]}"
    nStress="${nStress_arr[$idx]}"
    el="${el_arr[$idx]}"
    
    # Loop over each blockTime value
    for bt in "${blockTime_arr[@]}"; do
        blockTime="$bt"
        # Create a descriptive folder for this test case
        folder="logs_ec${ec}_nStress${nStress}_el${el}_blockTime${blockTime}"
        mkdir -p "$folder"
        echo "Running test case: $folder"
        
        # Combine the variables for this run
        vars="ec=${ec} el=${el} nStress=${nStress} blockTime=${blockTime} blockTimeNs=${blockTimeNs} block_sz_bytes=${block_sz_bytes} block_size=${block_size}"
        
        # Setup phase
        ansible-playbook -i ../tf/servers.ini setup.yml -f 55 -e "$vars" | tee "${folder}/setup.log"
        
        # Prerun phase
        ansible-playbook -i ../tf/servers.ini prerun.yml -f 55 -e "$vars" | tee "${folder}/prerun.log"
        
        # Run the load in parallel based on nStress
        for (( i=1; i<=nStress; i++ )); do
            ansible-playbook -i ../tf/servers.ini runload.yml -f 55 -e "$vars stressID=$i" >> "${folder}/runload.log" 2>&1 &
        done
        # Wait for all parallel jobs to complete
        wait
        
        # Postrun phase
        ansible-playbook -i ../tf/servers.ini postrun.yml -f 55 -e "$vars" | tee "${folder}/postrun.log"
    done
done
