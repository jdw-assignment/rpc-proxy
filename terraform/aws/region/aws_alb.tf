resource "random_integer" "bucket_random" {
  min = 1000000
  max = 9999999
}

module "s3_bucket" {
  source        = "terraform-aws-modules/s3-bucket/aws"
  bucket        = "alb-${random_integer.bucket_random.result}"
  acl           = "log-delivery-write"
  force_destroy = true

  control_object_ownership = true
  object_ownership         = "ObjectWriter"

  attach_elb_log_delivery_policy = true
  attach_lb_log_delivery_policy  = true
}

module "alb" {
  source = "terraform-aws-modules/alb/aws"
  version = "~> 9.0"

  load_balancer_type = "application"

  vpc_id  = module.vpc.vpc_id
  subnets = module.vpc.public_subnets
  security_groups = [aws_security_group.default_sg.id]

  access_logs = {
    enabled = true
    bucket  = module.s3_bucket.s3_bucket_id
  }

  // For testing
  enable_deletion_protection = false
}

resource "aws_lb_listener" "http_tg" {
  load_balancer_arn = module.alb.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"
    fixed_response {
      status_code  = "401"
      content_type = "text/plain"
    }
  }
}