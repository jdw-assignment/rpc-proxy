module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.9.0"

  name   = "proxy-app"
  azs = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  public_subnets = ["10.0.0.0/20", "10.0.16.0/20", "10.0.32.0/24"]
  private_subnets = ["10.0.80.0/20", "10.0.112.0/20", "10.0.160.0/24"]

  enable_nat_gateway = true
}

resource "aws_security_group" "default_sg" {
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}