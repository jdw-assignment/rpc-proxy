terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

module "rpc-proxy-svc" {
  source = "../services/rpc-proxy"

  ecs_cluster_name      = module.ecs.cluster_name
  vpc_id                = module.vpc.vpc_id
  lb_listener_arn = aws_lb_listener.http_tg.arn
  subnets               = module.vpc.private_subnets
  alb_security_group    = aws_security_group.default_sg.id
  execution_role_arn    = aws_iam_role.execution_role.arn
  task_role_arn         = aws_iam_role.task_role.arn

  svc_name = "rpc-proxy-svc"
  tag  = "0.0.1"

  rpc_url = "https://polygon-rpc.com"
}

