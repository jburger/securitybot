#---- API GATEWAY ----

resource "aws_apigatewayv2_api" "securitybot_apigw" {
  name          = var.api_gateway_name
  protocol_type = "HTTP"
}

