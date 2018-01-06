# Three-phase commit protocol

### Example usage:

```
go run coordinator/coordinator.go localhost 5672 3
```

```
go run cohort/cohort.go localhost 5672
```

```
go run cohort/cohort.go localhost 5672
abort
```

```
go run cohort/cohort.go localhost 5672
write /path/to/file.txt content
```
