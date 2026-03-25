Запрос к БД Mongo
```bash
$ curl -X POST http://localhost:8080/products \
     -H "Content-Type: application/json" \
     -d '{"name": "iPhone 15", "sku": "apple-15", "price": 999.99, "quantity": 10}'
{"id":"69c39d5e7d29a894c5559204","name":"iPhone 15","sku":"apple-15","price":999.99,"quantity":10,"updated_at":"2026-03-25T08:31:26.9049421Z"}
```

Получение данных из БД mongodb
```bash
curl http://localhost:8080/products/apple-15
```