resource "aws_api_gateway_rest_api" "mother-kellys" {
  name = "mother-kellys"
}

resource "aws_api_gateway_resource" "mother-kellys" {
  rest_api_id = aws_api_gateway_rest_api.mother-kellys.id
  parent_id   = aws_api_gateway_rest_api.mother-kellys.root_resource_id
  path_part   = "beer"
}

resource "aws_api_gateway_method" "mother-kellys" {
  rest_api_id   = aws_api_gateway_rest_api.mother-kellys.id
  resource_id   = aws_api_gateway_resource.mother-kellys.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "mother-kellys" {
  rest_api_id             = aws_api_gateway_rest_api.mother-kellys.id
  resource_id             = aws_api_gateway_resource.mother-kellys.id
  http_method             = aws_api_gateway_method.mother-kellys.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.aws_region}:lambda:path/2015-03-31/functions/arn:aws:lambda:${var.aws_region}:${var.aws_account}:function:${aws_lambda_function.get_latest_menu.function_name}/invocations"
}

resource "aws_api_gateway_deployment" "mother-kellys" {
  rest_api_id = aws_api_gateway_rest_api.mother-kellys.id
  stage_name  = "mk"

  depends_on = [aws_api_gateway_integration.mother-kellys]
}
