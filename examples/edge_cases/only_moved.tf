moved {
  from = aws_instance.old
  to   = aws_instance.new
}

moved {
  from = aws_s3_bucket.old
  to   = aws_s3_bucket.new
}
