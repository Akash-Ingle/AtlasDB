module "networking" {
  source       = "../../modules/networking"
  project_name = "atlasdb"
  environment  = "production"
  vpc_cidr     = "10.1.0.0/16"
  region       = "us-east-1"
}

module "database" {
  source         = "../../modules/database"
  project_name   = "atlasdb"
  environment    = "production"
  vpc_id         = module.networking.vpc_id
  subnet_ids     = module.networking.private_subnet_ids
  instance_class = "db.r6g.large"
}

module "redis" {
  source       = "../../modules/redis"
  project_name = "atlasdb"
  environment  = "production"
  vpc_id       = module.networking.vpc_id
  subnet_ids   = module.networking.private_subnet_ids
  node_type    = "cache.r6g.large"
}

module "kubernetes" {
  source        = "../../modules/kubernetes"
  project_name  = "atlasdb"
  environment   = "production"
  vpc_id        = module.networking.vpc_id
  subnet_ids    = module.networking.private_subnet_ids
  instance_type = "t3.xlarge"
  desired_nodes = 3
}
