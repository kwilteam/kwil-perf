terraform {
    required_providers {
        aws = {
            source = "hashicorp/aws"
            version = "~> 5.0"
        }
    }
}

provider "aws" {
    region = "us-east-1"
    alias = "virginia"
}

provider "aws" {
    region = "us-west-1"
    alias = "california"
}

provider "aws" {
    region = "eu-central-1"
    alias = "frankfurt"
}