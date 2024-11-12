resource "aws_instance" "virginia-nodes" {
    provider = aws.virginia
    count = 12
    launch_template {
        id = "lt-091a5e2dcd9c3e9d8"
        version = "$Latest"
    }
    tags = {
        Name = "testnode-nv-${count.index}"
    }
}

resource "aws_instance" "california-nodes" {
    provider = aws.california
    count = 12
    launch_template {
        id = "lt-005f1be8687435f21"
        version = "$Latest"
    }
    tags = {
        Name = "testnode-cal-${count.index}"
    }
}

resource "aws_instance" "frankfurt-nodes" {
    provider = aws.frankfurt
    count = 11
    launch_template {
        id = "lt-09eee0590b4c5c925"
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
        echo '${join("\n", concat(aws_instance.virginia-nodes.*.public_ip, aws_instance.california-nodes.*.public_ip, aws_instance.frankfurt-nodes.*.public_ip))}' > ips.txt
        EOT
    }
}