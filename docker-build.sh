#!/bin/bash

version=$1

docker build --no-cache --build-arg "VERSION=$version" --tag "korylprince/chronicle-server:$version" .

docker push "korylprince/chronicle-server:$version"
