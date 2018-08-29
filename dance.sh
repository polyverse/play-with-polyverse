docker-compose stop
docker rm $(docker ps -qa) 
docker-compose up -d
