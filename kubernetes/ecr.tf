resource "aws_ecr_repository" "pgvector_postgres" {
  name                 = "pgvector-postgres"
  image_tag_mutability = "MUTABLE"

  encryption_configuration {
    encryption_type = "AES256"
  }

  force_delete = true
}

resource "aws_ecr_repository" "rag_api_go" {
  name                 = "felixmoronge-rag-api-go"
  image_tag_mutability = "MUTABLE"

  encryption_configuration {
    encryption_type = "AES256"
  }

  force_delete = true
}

resource "aws_ecr_repository" "rag_embed_py" {
  name                 = "felixmoronge-rag-embed-py"
  image_tag_mutability = "MUTABLE"

  encryption_configuration {
    encryption_type = "AES256"
  }

  force_delete = true
}
output "embed_api_access_key_id" {
  value = aws_iam_access_key.embed_api_key.id
  sensitive = true
}

output "embed_api_access_key_secret" {
  value = aws_iam_access_key.embed_api_key.secret
  sensitive = true
}
resource "aws_iam_user" "terraform_admin" {
  name = "terraform-admin"
}

resource "aws_iam_user_policy" "terraform_admin_ecr_inline" {
  name = "terraform-admin-ecr-inline"
  user = aws_iam_user.terraform_admin.name

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:CreateRepository"
        ],
        Resource = "*"
      }
    ]
  })
}
resource "aws_iam_user" "github_actions_embed_api" {
  name = "github-actions-embed-api"
}

resource "aws_iam_access_key" "embed_api_key" {
  user = aws_iam_user.github_actions_embed_api.name
}

resource "aws_iam_user_policy" "embed_api_ecr_push" {
  name = "ecr-push-policy"
  user = aws_iam_user.github_actions_embed_api.name

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:CreateRepository"
        ],
        Resource = "*"
      }
    ]
  })
}