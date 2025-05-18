# Description

Go program for send and receive message through rabbitmq

## Prepare

Launch rabbitmq server

```bash
 docker run -d --hostname my-rabbit --name rabbit -p 8080:15672  -p 5672:5672 rabbitmq:4-management
 # d        - detach mode
 # hostname - name of host inside docker
 # name     - container name
 # p        - ports (8080 - GUI, 5672 - amqp )
```

## Configuration

```yaml
rabbit_url: amqp://guest:guest@localhost:5672/
exchange_name: ""
queue_name: hello
queue_durable: true
queue_auto_delete: false
queue_exclusive: false
queue_no_wait: false
```

## Run

```bash
go run .
2025/05/18 10:35:54  [x] Sender 1 sent {"content":"Hello World! 3"}
2025/05/18 10:35:54  [x] Sender 1 sent {"content":"Hello World! 4"}
2025/05/18 10:35:54  [x] Sender 0 sent {"content":"Hello World! 1"}
2025/05/18 10:35:54  [x] Sender 0 sent {"content":"Hello World! 2"}
2025/05/18 10:35:55 Consumer 3 received a message with routing key hello: {"content":"Hello World! 3"}
2025/05/18 10:35:55 Consumer 3 received a message with routing key hello: {"content":"Hello World! 4"}
2025/05/18 10:35:55 Consumer 3 received a message with routing key hello: {"content":"Hello World! 1"}
2025/05/18 10:35:55 Consumer 3 received a message with routing key hello: {"content":"Hello World! 2"}
```
