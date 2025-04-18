terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

module "vpc" {
  source = "./networking"
}

moved {
  from = module.network
  to   = module.vpc
}
