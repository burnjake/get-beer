resource "aws_lambda_function" "get_latest_menu" {
  filename      = "/tmp/payload.zip"
  function_name = "get_latest_menu"
  role          = "${aws_iam_role.lambda-role.arn}"
  handler       = "main"
  runtime       = "go1.x"
}
