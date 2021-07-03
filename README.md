# Algotrading-GoLang-Rest
Golang APIs using REST

## Docker Operations

1. Build a Docker Image
```
docker build -t user-authentication-service .
```
```
docker build -t rest-exchange-service .
```
```
docker build -t rest-strategy-service .
```
```
docker build -t rest-price-service .
```
```
docker build -t rest-order-service .
```
2. Tag the docker Image
```
docker tag user-authentication-service:latest vikash99/user-authentication-service
```
3. Push the docker image to dockerhub
```
docker push vikash99/user-authentication-service
```
4. Run the docker container
```
docker run -p 3000:3000 --name user-authentication-service user-authentication-service
```

## Kubernetes Deployments

### Setup EFK Stack on Kubernetes Cluster for Logging
1. Create a namespace
```
kubectl apply -f kubernetes-deployments/logging/namespace.yaml
```
If the namespace name is changed, change it in other elastisearch.yaml,kibana.yaml & fluentd.yaml as well.

2. Setup Elastisearch <br />
Enter your cluster name at line 55 & storage class name at line 92
```
kubectl apply -f kubernetes-deployments/logging/elastisearch.yaml
```
3. Setup Kibana
```
kubectl apply -f kubernetes-deployments/logging/kibana.yaml
```
4. Setup fluentd DeamonSet
```
kubectl apply -f kubernetes-deployments/logging/fluentd.yaml
```
5. Open Kibana dashboard
```
kubectl port-forward $kibanaPodName 5601:5601 -n logging
```
### Create a standalone mongodb statefulset
1. To create a standalone mongodb
```
kubectl apply -f kubernetes-deployments/mongodb/mongodb.yaml
```
2. To exec into mongodb from CLI
```
kubectl -n hedgina exec -it mongodb-0 -- mongo mongodb://mongoadmin:mongopassword@mongodb-0.database:27017/?authSource=admin
```
or
```
kubectl -n hedgina exec -it mongodb-0 -- mongo mongodb://mongodb-0.database:27017 --username mongoadmin --password mongopassword
```
3. To connect to mongodb from inside the cluster use the connection String:
```mongodb://mongoadmin:mongopassword@mongodb-0.database:27017/?authSource=admin``` using standard connection <br />
string format: ```mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]```
4. Example of mongodb connection string for multiple replicaset: ```mongodb://mongoadmin:mongopassword@mongodb-0.database:27017,mongodb-1.database:27017,mongodb-2.database:27017/?authSource=admin```
5. To connect to mongodb running on k8s cluster from your local, port-forward the mongodb pod to localhost:27017 using ```kubectl -n hedgina port-forward mongodb-0 27017:27017```. To stop, hit Ctrl+C

