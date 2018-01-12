# Three-phase commit protocol

### Example usage:

```
go run coordinator/coordinator.go localhost 5672 4 failure1
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

### Docker:
```
docker build -t "matb4r:3pc" .
docker run -it matb4r:3pc bash
```

### RabbitMQ server:
```
echo "[{rabbit, [{loopback_users, []}]}]." | sudo tee /etc/rabbitmq/rabbitmq.config > /dev/null
```
