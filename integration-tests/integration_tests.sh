funtion set_up_tests
{
  echo "Running Integration Tests"
}

function tear_down_tests
{
  echo "Integration Tests Complete"
}


# configure cli
# ecs-cli configure --region us-west-2 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --cluster ecs-integrations-tests
# create a cluster using the default config
# ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 1 --instance-type t2.medium --force
# create a second cluster using flags, in Ohio region
#ecs-cli up --keypair MyFirstKeyPair --capability-iam --size 1 --instance-type t2.medium --cluster ecs-integrations-tests-other-cluster --region us-east-2

for FILE in $(ls)
do
	if [[ -d $FILE ]]; then
		cd $FILE
    chmod +x test.sh
		source test.sh
    doTest | sed "s/^/[$FILE] /"
		cd ..
	fi
done
