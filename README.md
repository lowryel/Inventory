#### Inventory CRUD API 
* Golang Practice
------------------------
* Functions
---------------
- CreateOwnerProduct // Create a product by entering  [POST](http://localhost/api/{userID}/products)
    ```json
        {
            "name": "Cauliflower",
            "price": 12.00,
            "instock": true,
            "quantity": 4
        }
    ```

- GetProducts // Get all products by issueing a get request to the endpoint [GET](http://localhost/api/products)

- GetProduct // Get one product by issueing a get request to the endpoint [GET](http://localhost/api/product/{id})

- DeleteProduct // Delete a product by issueing a Delete request to the endpoint [DELETE](http://localhost/api/product/delete/{id})

- UpdateProduct // Update a product info by issueing a Put request to the endpoint [PUT](http://localhost/api/product/update/{id})

    ```json
        {
            "name": "Appless",
            "price": 10.50,
            "instock": true,
            "quantity": 12
        }
    ```

- CreateUser // Create a user [POST](http://localhost/api/products/users)
    ```json
        {
            "username": "username",
            "phone": "0594212599",
            "email": "e2e7w@gmail.com",
            "password": "password",
            "first_name": "Fiden",
            "user_type":"USER" // or ADMIN
        }
    ```
- GetUser // Get a user by issueing a get request to the endpoint [GET](http://localhost/api/user/{userID})

- LoginHandler // Login by issueing a POST request to the endpoint [POST](http://localhost/api/login})
    ```json
    {
        "username":"username",
        "password":"password"
    }
    ```




# New Project Idea
    -   An expense tracker API
    -   Medium clone API