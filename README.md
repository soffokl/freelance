# Freelance exchange basic implementation

This implementation provides basic operations for a freelance exchange. 

There are two basic objects in the application: Users and Orders. 

A new user can be added to the database. Existing users can be listed by single requests. All users will be provided at the same time, no padding supported at that moment. 

Orders can be placed by any user. To place an ordering user should have enough balance. Any user can take and finish order to get a reward on his balance. 

No auth operation supported, and some operations require to pass user_id manually

The application expect PostgreSQL service to be installed and running on the same host with no auth. (`"host=localhost port=5432 dbname=freelance sslmode=disable"`)

## REST API
`GET "/users/"` - get the list of all users.

`POST "/users/"` - add new user.

`GET "/orders/"` - get the list of all available orders.

`POST "/orders/"` - add a new order.

`PUT "/orders/{order_id}/{status}"` - reserve the order or mark it as done.

## Database 

![image](https://user-images.githubusercontent.com/8612618/43514858-e7b7cc8a-95a2-11e8-8c3e-f197c6b904d4.png)

## Testing

You need to have an empty database for starting tests. It does not clear anything befor or after tests. 

To start a test you can use command like this:

```
go test ./... -v -count 1 -race
```