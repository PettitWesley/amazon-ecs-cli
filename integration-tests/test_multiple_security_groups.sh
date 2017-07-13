ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-cli-test0SG
ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 1 --instance-type t2.medium --force
ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-cli-test1SG
echo "Testing with --security-group"
ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 2 --instance-type t2.medium --security-group sg-40a6a43b --vpc vpc-26c72340 --subnets subnet-efccfd88,subnet-63622a2a --force
ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-cli-test2SG
ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 2 --instance-type t2.medium --security-group sg-40a6a43b,sg-fc2b2087 --vpc vpc-26c72340 --subnets subnet-efccfd88,subnet-63622a2a --force
ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-cli-test3SG
ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 2 --instance-type t2.medium --security-group sg-40a6a43b,sg-fc2b2087,sg-12faf169 --vpc vpc-26c72340 --subnets subnet-efccfd88,subnet-63622a2a --force
echo "--security-groups succeeded"
# test the new flag option --security-groups
echo "Testing with --security-groups"
ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-cli-test2SGs
ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 2 --instance-type t2.medium --security-groups sg-40a6a43b,sg-fc2b2087 --vpc vpc-26c72340 --subnets subnet-efccfd88,subnet-63622a2a --force
ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-cli-test1SGs
ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 2 --instance-type t2.medium --security-groups sg-40a6a43b --vpc vpc-26c72340 --subnets subnet-efccfd88,subnet-63622a2a --force
echo "--security-groups succeeded"

After subcommand:
ecs-cli compose create --cluster other --region us-east-2
ecs-cli compose start --cluster other --region us-east-2
ecs-cli compose up --cluster other --region us-east-2
ecs-cli compose ps --cluster other --region us-east-2
ecs-cli compose scale 2 --cluster other --region us-east-2
ecs-cli compose run --cluster other --region us-east-2
ecs-cli compose stop --cluster other --region us-east-2
ecs-cli compose down --cluster other --region us-east-2

Before subcommand:
ecs-cli compose --cluster other --region us-east-1 create
ecs-cli compose --cluster other --region us-east-1 start
ecs-cli compose --cluster other --region us-east-1 scale 2
ecs-cli compose --cluster other --region us-east-1 stop
ecs-cli compose --cluster other --region us-east-1 up
ecs-cli compose --cluster other --region us-east-1 down


# Service commands

After subcommand:
ecs-cli compose service create --cluster other --region us-east-1
ecs-cli compose service start --cluster other --region us-east-1
ecs-cli compose service up --cluster other --region us-east-1
ecs-cli compose service ps --cluster other --region us-east-1
ecs-cli compose service scale 2 --cluster other --region us-east-1
ecs-cli compose service stop --cluster other --region us-east-1
ecs-cli compose service down --cluster other --region us-east-1

Before subcommand:
ecs-cli compose service --cluster other --region us-east-1 start
ecs-cli compose service --cluster other --region us-east-1 scale 2
ecs-cli compose service --cluster other --region us-east-1 stop
ecs-cli compose service --cluster other --region us-east-1 ps
ecs-cli compose service --cluster other --region us-east-1 up
ecs-cli compose service --cluster other --region us-east-1 down

Cats:
ecs-cli compose --cluster other --region us-east-1 service start
ecs-cli compose --cluster other --region us-east-1 service scale 2
ecs-cli compose --cluster other --region us-east-1 service stop
ecs-cli compose --cluster other --region us-east-1 service ps
ecs-cli compose --cluster other --region us-east-1 service up
ecs-cli compose --cluster other --region us-east-1 service down


ecs-cli compose service up --cluster other --region us-east-1
ecs-cli compose service --cluster other --region us-east-1 up
ecs-cli compose --cluster other --region us-east-1 service up
