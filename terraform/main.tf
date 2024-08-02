module "ap_southeast_1" {
  source                   = "./aws/region"

  providers = {
    aws = aws.ap-southeast-1
  }
}