#!/bin/bash
# Check if a command-line argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <numStressTools>"
    exit 1
fi

numStressTools=$1
ec=$2
el=$3

count=0
while [ $count -lt $numStressTools ]; do
    echo "Starting stress instance $((count+1)) of $numStressTools"
    nohup /data/bin/stress -ec 40 -el 10000  -ne -run 3m > /data/node/stress.log 2>&1 &
    count=$((count+1))
done

# Optional: Wait for all background processes to complete
wait
echo "All stress instances have been started."