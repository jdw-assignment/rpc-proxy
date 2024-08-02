locals {
  ecs_cluster_name = "proxy-app"
}

module "ecs" {
  source  = "terraform-aws-modules/ecs/aws"
  version = "5.11.3"

  cluster_name = local.ecs_cluster_name

  fargate_capacity_providers = {
    FARGATE = {
      default_capacity_provider_strategy = {
        weight = 1
      }
    }
  }
}
