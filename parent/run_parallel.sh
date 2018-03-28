export BUILD=false
export MOCK=false
while test $# -gt 0; do  
    case "$1" in
        -b)
                if test $# -gt 0; then
                    export BUILD=true
                fi
                shift
                ;;
        -m)
                if test $# -gt 0; then
                     export MOCK=true
                fi
                shift
                ;;
    esac
done

if [ $BUILD = true ] ; then
    ./create_images.sh
fi

docker swarm init
docker network create -d overlay --attachable --subnet=172.20.0.0/16 --gateway=172.20.0.1 overlay-network
docker network create -d overlay --attachable --subnet=172.19.0.0/16 --gateway=172.19.0.1 bridge-network

if [ $MOCK = true ] ;
    then
        go build -o ../mock-legacy-quoteserve/mockQuoteServe
        ../mock-legacy-quoteserve/mockQuoteServe &
fi

docker service  rm stack_trigger stack_quote stack_transaction stack_database stack_audit stack_proxy_web
env $(cat .env | grep ^[A-Za-z_] | xargs) docker stack deploy -c docker-compose-deploy.yml stack

echo "check docker service ls to make sure services are ready before running workload gen"
