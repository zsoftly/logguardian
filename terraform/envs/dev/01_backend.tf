terraform {
  backend "s3" {
    bucket = "zsoftly-poc-sandbox-terraform-cac1"
    key    = "zsoftly/logguardian/terraform/envs/dev/terraform.tfstate"
    region = "ca-central-1"
  }
}
