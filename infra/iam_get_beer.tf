resource "aws_iam_user" "get-beer" {
  name = "get-beer"
}

resource "aws_iam_access_key" "get-beer" {
  user = aws_iam_user.get-beer.name
}

data "aws_iam_policy_document" "s3-put" {
  statement {
    actions = [
      "s3:PutObject"
    ]

    resources = [
      aws_s3_bucket.mother-kellys.bucket,
    ]
  }
}

resource "aws_iam_policy" "s3_put" {
  name        = "s3_put"
  policy      = data.aws_iam_policy_document.s3-put.json
}

resource "aws_iam_user_policy_attachment" "attach-s3" {
  user       = aws_iam_user.get-beer.name
  policy_arn = aws_iam_policy.s3_put.arn
}

data "aws_iam_policy_document" "dynamo-put" {
  statement {
    actions = [
      "dynamodb:PutItem"
    ]

    resources = [
      aws_dynamodb_table.mother-kellys.arn,
    ]
  }
}

resource "aws_iam_policy" "dynamo_put" {
  name        = "dynamo_put"
  policy      = data.aws_iam_policy_document.dynamo-put.json
}

resource "aws_iam_user_policy_attachment" "attach-dynamo" {
  user       = aws_iam_user.get-beer.name
  policy_arn = aws_iam_policy.dynamo_put.arn
}
