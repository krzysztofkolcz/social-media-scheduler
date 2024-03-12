# scp -i ~/AWSkeys/testing-golang-2.pem app/cmd/gorilla/gorilla ec2-user@ec2-3-70-206-107.eu-central-1.compute.amazonaws.com:/home/ec2-user/source/api/app/cmd/gorilla/gorilla
# scp -i ~/AWSkeys/testing-golang-2.pem app/Dockerfile ec2-user@ec2-3-70-206-107.eu-central-1.compute.amazonaws.com:/home/ec2-user/source/api/app/Dockerfile
# scp -i ~/AWSkeys/testing-golang-2.pem /home/krzysztof/source/api/compose.yaml ec2-user@ec2-3-70-206-107.eu-central-1.compute.amazonaws.com:/home/ec2-user/source/api/compose.yaml
scp -i ~/AWSkeys/testing-golang-2.pem /home/krzysztof/source/api/Makefile ec2-user@ec2-3-70-206-107.eu-central-1.compute.amazonaws.com:/home/ec2-user/source/api/Makefile
