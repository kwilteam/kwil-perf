#!/bin/bash
set -e  # Stops script on first error

cd ../ansible

# Run Ansible Playbook and save the output
ansible-playbook -i ../tf/servers.ini installation.yml 
