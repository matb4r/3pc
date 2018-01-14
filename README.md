# Three-phase commit protocol

### Example usage:

```
go run coordinator/coordinator.go 172.17.0.1 5672 4 failure1
```

```
go run cohort/cohort.go 172.17.0.1 5672 
```

```
go run cohort/cohort.go 172.17.0.1 5672 abort
```

```
go run cohort/cohort.go 172.17.0.1 5672 failure2
```

```
go run cohort/cohort.go 172.17.0.1 5672 write /tmp/file.txt hello world 
```

### Docker with shell:
```
docker build -t "matb4r:3pc" .
docker run -it matb4r:3pc /bin/sh
```

### Docker exec:

```
docker build -t matb4r:coordinator coordinator
```

```
docker build -t "matb4r:cohort" cohort
```

```
docker run -it matb4r:coordinator 172.17.0.1 5672 1 failure2
```

```
docker run -it matb4r:cohort 172.17.0.1 5672 write /tmp/file.txt hello world
```

### RabbitMQ server config:
```
echo "[{rabbit, [{loopback_users, []}]}]." | sudo tee /etc/rabbitmq/rabbitmq.config > /dev/null
```
