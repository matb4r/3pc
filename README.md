# Three-phase commit protocol

### Example usage:

```
go run coordinator/coordinator.go localhost 5672 4 failure1
```

```
go run cohort/cohort.go localhost 5672 
```

```
go run cohort/cohort.go localhost 5672 abort
```

```
go run cohort/cohort.go localhost 5672 failure2
```

```
go run cohort/cohort.go localhost 5672 write /tmp/file.txt hello world 
```
