# events_on-the-way
Представляет собой систему по отправке событий через Kafka, хранящихся в базе, пришедших keeper'у по http.

![alt text](static/blockschema.png) .

Отправка событий (**producer**) происходят **идемпотентно** и способом **exactly once**:
```sh
config.Producer.RequiredAcks = sarama.WaitForAll
config.Producer.Idempotent = true
```

Poller проверяет БД каждые 5 секунд на наличие событий со статусом NEW и отдает продюсеру.

Происходит отправка метрик в **prometheus**.