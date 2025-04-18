resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = "MainVPC"
  }
}

moved {
  from = aws_vpc.primary
  to   = aws_vpc.main
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

moved {
  from = aws_subnet.external
  to   = aws_subnet.public
}
