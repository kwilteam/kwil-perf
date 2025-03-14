variable "ssh_key_path" {
  default = "~/.ssh/kwil-aws"
  description = "Path to the SSH private key file"
}

variable "virginia_count" {
  default = 5
  description = "Number of nodes to create in Virginia"
}

variable "california_count" {
  default = 5
  description = "Number of nodes to create in Virginia"
}

variable "frankfurt_count" {
  default = 5
  description = "Number of nodes to create in Virginia"
}
