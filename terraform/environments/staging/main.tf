module "networking" {
  source       = "../../modules/networking"
  project_name = "atlasdb"
  environment  = "staging"
  vpc_cidr     = "10.0.0.0/16"
  region       = "us-east-1"
}

module "database" {
  source         = "../../modules/database"
  project_name   = "atlasdb"
  environment    = "staging"
  vpc_id         = module.networking.vpc_id
  subnet_ids     = module.networking.private_subnet_ids
  instance_class = "db.t3.medium"
}

module "redis" {
  source       = "../../modules/redis"
  project_name = "atlasdb"
  environment  = "staging"
  vpc_id       = module.networking.vpc_id
  subnet_ids   = module.networking.private_subnet_ids
  node_type    = "cache.t3.micro"
}

module "kubernetes" {
  source        = "../../modules/kubernetes"
  project_name  = "atlasdb"
  environment   = "staging"
  vpc_id        = module.networking.vpc_id
  subnet_ids    = module.networking.private_subnet_ids
  instance_type = "t3.medium"
  desired_nodes = 2
}
