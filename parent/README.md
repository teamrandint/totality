The script run_parallel takes arguments:
    -m start the mock legacy quote server
    -b rebuild the images

It will set up a swarm, networks, remove all running services, and create a set of new ones

Use docker service ls to see when they're online

Arguments such as how many web, and trans servers to run are in .env

Entry point is the proxy server at localhost:$proxyport


