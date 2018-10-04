#!/bin/bash

buildDockerfile() {
  filename="$1"
  imagePrefix="$2"
  
  echo "Obtaining current git sha for tagging the docker image"
  headsha=$(git rev-parse --verify HEAD)
  echo "Git sha is $headsha"

  imageName="$imagePrefix:$headsha"

  docker build -t $imageName -f $filename .
  docker push $imageName 
}

buildDockerfile "Dockerfile" "polyverse/play-with-polyverse"

buildDockerfile "Dockerfile" "polyverse/play-with-polyverse-l2"
