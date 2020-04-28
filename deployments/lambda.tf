resource "aws_lambda_function" "get_latest_menu" {
  filename      = "/tmp/main.zip"
  function_name = "get_latest_menu"
  role          = aws_iam_role.lambda-role.arn
  handler       = "main"
  runtime       = "go1.x"
}

resource "aws_lambda_permission" "allow_apigateway" {
  statement_id  = "allow_apigateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.get_latest_menu.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.mother-kellys.execution_arn}/*/*${aws_api_gateway_resource.mother-kellys.path}"
}
