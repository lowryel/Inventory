#### Inventory CRUD API 
* Golang Practice
------------------------
* Functions
---------------
- CreateProduct // Create a product by entering  [POST](http://localhost/api/products)
    `{
    "name": "Cauliflower",
    "price": 12.00,
    "instock": true,
    "quantity": 4
    }`

- GetProducts // Get all products by issueing a get request to the endpoint [GET](http://localhost/api/products)

- GetProduct // Get one product by issueing a get request to the endpoint [GET](http://localhost/api/product/{id})

- DeleteProduct // Delete a product by issueing a Delete request to the endpoint [DELETE](http://localhost/api/product/delete/{id})

- UpdateProduct // Update a product info by issueing a Put request to the endpoint [PUT](http://localhost/api/product/update/{id})
