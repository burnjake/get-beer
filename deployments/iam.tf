resource "aws_iam_role" "lambda-role" {
  name = "lambda-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

data "aws_iam_policy_document" "dynamodb-read-policy-document" {
  statement {
    sid = "1"

    actions = [
      "dynamodb:Query"
    ]

    resources = [
      aws_dynamodb_table.mother-kellys.arn,
    ]
  }
}

resource "aws_iam_policy" "dynamodb-read" {
  name   = "dynamodb-read"
  policy = data.aws_iam_policy_document.dynamodb-read-policy-document.json
}

resource "aws_iam_role_policy_attachment" "dynamo-read-policy-attachment" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = aws_iam_policy.dynamodb-read.arn
}


resource "aws_iam_role_policy_attachment" "lambda-basic-policy-attachment" {
  role       = aws_iam_role.lambda-role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
