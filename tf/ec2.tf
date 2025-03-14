
resource "aws_instance" "virginia-nodes" {
    provider = aws.virginia
    count = var.virginia_count
    launch_template {
        id = "lt-05fccdfc45053a91d"
        version = "$Latest"
    }
    tags = {
        Name = "testnode-nv-${count.index}"
    }
}

resource "aws_instance" "california-nodes" {
    provider = aws.california
    count = var.california_count
    launch_template {
        id = "lt-06a002fc7848e7d43"
        version = "$Latest"
    }
    tags = {
        Name = "testnode-cal-${count.index}"
    }
}

resource "aws_instance" "frankfurt-nodes" {
    provider = aws.frankfurt
    count = var.frankfurt_count
    launch_template {
        id = "lt-063c73300d51ae591"
        version = "$Latest"
    }
    tags = {
        Name = "testnode-fr-${count.index}"
    }
}

output "public_ip" {
    value = concat(aws_instance.virginia-nodes.*.public_ip, aws_instance.california-nodes.*.public_ip, aws_instance.frankfurt-nodes.*.public_ip)
}

resource "null_resource" "write_public_ips_to_file" {
    # This depends on both instance resources to ensure IPs are written after they're created
    depends_on = [
        aws_instance.virginia-nodes,
        aws_instance.california-nodes,
        aws_instance.frankfurt-nodes
    ]

    provisioner "local-exec" {
        command = <<EOT
            echo '${join("\n", concat(
                aws_instance.virginia-nodes.*.public_ip, 
                aws_instance.california-nodes.*.public_ip, 
                aws_instance.frankfurt-nodes.*.public_ip
            ))}' > ips2.txt

            # Generate servers.ini
            echo '[leader]' > servers.ini
            echo '${element(aws_instance.virginia-nodes.*.public_ip, 0)}' >> servers.ini
            echo '${element(aws_instance.virginia-nodes.*.public_ip, 0)}' >> ips.txt


            echo '[validators]' >> servers.ini
            echo '${join("\n", slice(aws_instance.virginia-nodes.*.public_ip, 1, (var.virginia_count)-2))}' >> servers.ini
            echo '${join("\n", slice(aws_instance.virginia-nodes.*.public_ip, 1, (var.virginia_count)-2))}' >> ips.txt
            echo '${join("\n", slice(aws_instance.california-nodes.*.public_ip, 0, (var.california_count)-1))}' >> servers.ini
            echo '${join("\n", slice(aws_instance.california-nodes.*.public_ip, 0, (var.california_count)-1))}' >> ips.txt
            echo '${join("\n", slice(aws_instance.frankfurt-nodes.*.public_ip, 0, (var.frankfurt_count)-2))}' >> servers.ini
            echo '${join("\n", slice(aws_instance.frankfurt-nodes.*.public_ip, 0, (var.frankfurt_count)-2))}' >> ips.txt

            # Last 2 IPs in virginia and frankfurt and last 1 from california are sentry

            echo '[sentry]' >> servers.ini
            echo '${join("\n", slice(aws_instance.virginia-nodes.*.public_ip, (var.virginia_count)-2, (var.virginia_count)))}' >> servers.ini
            echo '${join("\n", slice(aws_instance.virginia-nodes.*.public_ip, (var.virginia_count)-2, (var.virginia_count)))}' >> ips.txt
            echo '${element(aws_instance.california-nodes.*.public_ip, (var.california_count)-1)}' >> servers.ini
            echo '${element(aws_instance.california-nodes.*.public_ip, (var.california_count)-1)}' >> ips.txt
            echo '${join("\n", slice(aws_instance.frankfurt-nodes.*.public_ip, (var.frankfurt_count)-2, (var.frankfurt_count)))}' >> servers.ini
            echo '${join("\n", slice(aws_instance.frankfurt-nodes.*.public_ip, (var.frankfurt_count)-2, (var.frankfurt_count)))}' >> ips.txt
            echo '' >> servers.ini
            echo '[all:vars]' >> servers.ini
            echo 'ansible_ssh_user=ubuntu' >> servers.ini
            echo 'ansible_ssh_private_key_file="${var.ssh_key_path}"' >> servers.ini
            echo 'ansible_ssh_common_args="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"' >> servers.ini
            echo 'ansible_python_interpreter=/usr/bin/python3' >> servers.ini
        EOT
    }
}
