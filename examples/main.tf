provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "web_server" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  
  tags = {
    Name = "WebServer"
  }
}

moved {
  from = aws_instance.web
  to   = aws_instance.web_server
}

resource "aws_s3_bucket" "data" {
  bucket = "my-data-bucket"
}

moved {
  from = aws_s3_bucket.logs
  to   = aws_s3_bucket.data
}
