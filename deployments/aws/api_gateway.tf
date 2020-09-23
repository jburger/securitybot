#---- API GATEWAY ----

resource "aws_apigatewayv2_api" "securitybot_apigw" {
  name          = "securitybot"
  protocol_type = "HTTP"
}

