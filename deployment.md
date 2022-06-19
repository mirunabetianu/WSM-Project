### Setup the postgres locally
```bash
kubectl apply -f k8s/postgres-config.yaml
kubectl apply -f k8s/postgres.yaml
```

### Setup the services
```bash
kubectl apply -f k8s
```

### Forward ports to each service
```bash
kubectl port-forward service/order-service 5000
kubectl port-forward service/stock-service 5001
kubectl port-forward service/payment-service 5002
```
### Endpoints

Order service is accessible at http://localhost:5000/orders
Stock service is accessible at http://localhost:5001/stock
Payment service is accessible at http://localhost:5002/payment