variable "url" {
  type    = string
  default = "postgres://admin:postgres@postgres:5432/golang-structure-db?sslmode=disable"
}

env "local" {
  url = var.url

  migration {
    dir = "file://migrations"
  }
}
