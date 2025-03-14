cd ../tf
set -e  # Stops script on first error
terraform destroy -auto-approve

# remove the inventory files
rm servers.ini
rm ips.txt
rm ips2.txt

cd ../
rm -rf kwil