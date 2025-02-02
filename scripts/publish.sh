
version=`cat ./VERSION`

version=$((version+1))

docker build -f deploy/Dockerfile -t 192.168.1.24:30515/cm:$version .

docker push 192.168.1.24:30515/cm:$version

echo $version > ./VERSION

helm upgrade cm ./deploy/helmChart/cluster-monitor --set image.tag=$version

