resource "aws_dynamodb_table" "mother-kellys" {
  name           = "MotherKellysMenus"
  billing_mode   = "PROVISIONED"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "bar_location"
  range_key      = "created"

  attribute {
    name = "created"
    type = "S"
  }

  attribute {
    name = "bar_location"
    type = "S"
  }
}
