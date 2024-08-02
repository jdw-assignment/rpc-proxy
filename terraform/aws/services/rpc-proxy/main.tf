data "aws_region" "current" {}

resource "aws_cloudwatch_log_group" "main" {
  name              = var.svc_name
  retention_in_days = 30
}

resource "aws_lb_target_group" "main" {
  name_prefix          = "api-"
  protocol             = "HTTP"
  port                 = 80
  target_type          = "ip"
  vpc_id               = var.vpc_id
  deregistration_delay = 120
  slow_start           = 45

  lifecycle {
    create_before_destroy = true
  }

  health_check {
    enabled             = true
    interval            = 30
    path                = "/health"
    port                = "traffic-port"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    timeout             = 10
    protocol            = "HTTP"
    matcher             = "200"
  }
}

resource "aws_lb_listener_rule" "main" {
  listener_arn = var.lb_listener_arn

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  condition {
    http_request_method {
      values = ["POST"]
    }
  }
}

resource "aws_ecs_service" "main" {
  name            = var.svc_name
  cluster         = var.ecs_cluster_name
  task_definition = aws_ecs_task_definition.main.arn

  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200
  desired_count                      = 3 // To showcase all az
  health_check_grace_period_seconds  = 240

  enable_execute_command = true

  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    base              = 0
    weight            = 1
  }

  network_configuration {
    subnets          = var.subnets
    security_groups  = [aws_security_group.container_sg.id]
    assign_public_ip = false
  }

  deployment_controller {
    type = "ECS"
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = "rpc-proxy-app"
    container_port   = 8080
  }
}

resource "aws_security_group" "container_sg" {
  name        = "Default App Rules"
  description = "Allow container ports and internet access"
  vpc_id      = var.vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "allow_container_port" {
  security_group_id = aws_security_group.container_sg.id
  referenced_security_group_id = var.alb_security_group
  from_port         = 8080
  ip_protocol       = "tcp"
  to_port           = 8080
}

resource "aws_vpc_security_group_egress_rule" "allow_all_egress" {
  security_group_id = aws_security_group.container_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

resource "aws_ecs_task_definition" "main" {
  family = var.svc_name

  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  network_mode             = "awsvpc"
  execution_role_arn       = var.execution_role_arn
  task_role_arn            = var.task_role_arn

  container_definitions = jsonencode([
    {
      name   = "rpc-proxy-app"
      image  = "ghcr.io/joeldavidw/rpc-proxy-app:${var.tag}"
      cpu    = 256
      memory = 512
      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "PORT"
          value = "8080"
        },
        {
          name  = "RPC_URL"
          value = var.rpc_url
        },
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.main.name
          awslogs-region        = data.aws_region.current.name
          awslogs-stream-prefix = "rpc-proxy-app"
        }
      }
    }
  ])
}
