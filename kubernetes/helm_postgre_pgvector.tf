resource "helm_release" "pgvector" {
  name       = "pgvector"
  namespace  = "vector-db"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "postgresql"
  version    = "13.2.24"

  create_namespace = true

  values = [
    file("${path.module}/pgvector-values.yaml")
  ]
}
