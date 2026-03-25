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

Создать товары через API
```bash
curl -X POST http://localhost:8080/products \
     -H "Content-Type: application/json" \
     -d '{"name": "iPhone 15", "sku": "iphone", "price": 1000, "quantity": 10}'

curl -X POST http://localhost:8080/products \
     -H "Content-Type: application/json" \
     -d '{"name": "Samsung S24", "sku": "samsung", "price": 900, "quantity": 10}'
```

Подключение к БД Mongo
```bash
 docker exec -it mongodb_inventory mongosh
 ```

 ```bash
 use inventory_db
db.products.find({sku: "iphone"})
db.products.find({sku: "samsung"})
```

Получение всех товаров
```bash
docker exec -it mongodb_inventory mongosh inventory_db --eval "db.products.find()"
```

Просматривать логи 
```bash
 docker compose logs -f app
 ```

 Получнеие всех товаров
 ```bash
  curl "http://localhost:8080/products?limit=1&offset=1"
  ```