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
    count = 0
    launch_template {
        id = "lt-005f1be8687435f21"
        version = "$Latest"
    }
    tags = {
        Name = "testnode-cal-${count.index}"
    }
}


output "public_ip" {
    value = concat(aws_instance.virginia-nodes.*.public_ip, aws_instance.california-nodes.*.public_ip)
}