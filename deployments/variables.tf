variable "profile" {
  description = "AWS profile to use assume in the local credentials file"
}

variable "project" {
  description = "Project ID of GCP project that contains the tf state GCS bucket"
}

variable "aws_account" {
  description = "Account number of AWS account"
}

variable "aws_region" {
  description = "AWS region"
}
