resource "loopia_zone_record" "something_example_com" {
  domain    = "example.com"
  subdomain = "something"
  record = {
    type  = "A"
    value = "192.0.2.1"
  }
}