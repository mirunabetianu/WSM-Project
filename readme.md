# Web Scale Management Project

### Current state

In the project's root directory, the _main.go_ file establishes a simple connection to a Postgress database. Don't forget to fill in your credentials.

Sample implementation of an endpoint: **coming soon!**

### Task implementation

Create a new branch for any added feature/fix and use pull requests. 

### TODO

- [ ] Implement order microservices (endpoints + db)
- [ ] Implement payment microservices (endpoints + db)
- [ ] Implement stock microservices (endpoints + db)
- [ ] Setup event based communication between services
- [ ] Local deployment
- [ ] Cloud deployment

### Docker commands
```bash
docker build -t payment/Dockerfile payment .
docker build -f stock/Dockerfile -t stock .
docker build -f order/Dockerfile -t order .
```

```bash
docker pull emqx/emqx-ee:4.4.4
docker run -d --name emqx-ee -p 1883:1883 -p 8081:8081 -p 8083:8083 -p 8084:8084 -p 8883:8883 -p 18083:18083 emqx/emqx-ee:4.4.4
```