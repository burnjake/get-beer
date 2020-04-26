provider "aws" {
  region                  = "eu-west-1"
  shared_credentials_file = "/Users/tf_user/.aws/creds"
  profile                 = var.profile
}

terraform {
  backend "gcs" {
    bucket  = "tf_state_jake_personal"
    prefix  = "terraform/state/get-beer"
  }
}
