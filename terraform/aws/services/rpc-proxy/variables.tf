variable "ecs_cluster_name" {
  type = string
}

variable "execution_role_arn" {
  type = string
}

variable "task_role_arn" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "lb_listener_arn" {
  type = string
}

variable "tag" {
  type = string
}

variable "svc_name" {
  type = string
}

variable "subnets" {
  type = set(string)
}

variable "alb_security_group" {
  type = string
}

variable "rpc_url" {
  type = string
}
