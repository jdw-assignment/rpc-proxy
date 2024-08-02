resource "aws_appautoscaling_target" "as_target" {
  max_capacity = 5
  min_capacity = 1
  resource_id        = "service/${var.ecs_cluster_name}/${var.svc_name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace = "ecs"

  depends_on = [
    aws_ecs_service.main
  ]
}

resource "aws_appautoscaling_policy" "as_cpu" {
  name = "as_cpu"
  policy_type = "TargetTrackingScaling"
  resource_id = aws_appautoscaling_target.as_target.resource_id
  scalable_dimension = aws_appautoscaling_target.as_target.scalable_dimension
  service_namespace = aws_appautoscaling_target.as_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    target_value = 65
  }
}