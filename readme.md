# Web Scale Management Project

Here you can find the instructions on how to deploy the k8s cluster locally and how to access each endpoint

### Local K8s deployment

#### Setup the postgres locally (1)
```bash
kubectl apply -f k8s/postgres-config.yaml
kubectl apply -f k8s/postgres.yaml
```

#### Setup the services (2)
```bash
kubectl apply -f k8s
```

#### Forward ports to each service (3) - run each command in a different terminal window
```bash
kubectl port-forward service/order-service 5000
kubectl port-forward service/stock-service 5001
kubectl port-forward service/payment-service 5002
```
#### Endpoints after port forwarding (4)

Order service is accessible at http://localhost:5000/orders
Stock service is accessible at http://localhost:5001/stock
Payment service is accessible at http://localhost:5002/payment
