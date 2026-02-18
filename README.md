# Mockery

Simple and lightweight mock server.

Features:

- [x] Path matching
- [x] Regex Path matching
- [x] Query matching
- [x] Body matching
- [x] Header matching

- [x] JSON File export
- [x] JSON File import

- [x] Proxy

Persistence:

- [x] SqLite (.db file)
- [x] SqLite (in-memory)
- [x] PostgreSQL

## [ToC]

- [Running](#running)
- [Schema](#schema)
  - [Class diagram](#class-diagram)
  - [Validations](#validations)
- [Config](#config)
  - [Endpoints](#endpoints)
- [Examples](#examples)
  - [Not matched](#not-matched)
  - [Path Matching](#simple-path)
  - [RequestBody Matching](#requestbody-matching)
  - [RequestQuery Matching](#requestquery-matching)
  - [RequestHeader Matching](#requestheader-matching)
  - [Regexp Matching](#regexp-matching)
  - [Proxy](#proxy)

## Running

### Config file

Set up connection details in `mockery.yml` file

```yaml
dbType: SqLite
port: 8080
dbConfig:
  sqlite:
    connStr: "file:mockery.db?cache=shared&mode=rwc&_journal_mode=WAL"
  postgres:
    connStr: "postgresql://postgres:password@localhost:5432/mockery?sslmode=disable&application_name=mockery"
autoImport: false
```

Valid `dbType`:
- InMemory
- SqLite
- Postgres

Import directory: `.import/`

Export directory: `.export/`

### Overrides

Override values from `mockery.yml` file by providing additional arguments

```sh
./mockery --port=8081 --db=SqLite
```

### Default values

Skip `mockery.yml` file / override flags to use default values

- Db: InMemory
- Port: 8080

### Config validations

- If `dbType` is provided in config file or as an override flag, the connection string for is required.
- Other parameters will use default values

### Docker

Build image

```sh
docker build -t mockery .
```

Run container

```sh
docker run --name "mockery" -p 8080:8080 mockery
```

> Uncomment `# COPY .import/ ./.import/` to copy stubs for import.

## Schema

### Class diagram

```mermaid
classDiagram
  Mock <|-- QueryMatcher
  Mock <|-- BodyMatcher
  Mock <|-- HeaderMatcher

  class Mock {
      +int64 id
      +string method
      +string path
      +string regexPath
      +[]HeaderMatcher requestHeaderMatchers
      +[]QueryMatcher requestQueryMatchers
      +[]BodyMatcher requestBodyMatchers
      +int responseStatus
      +any responseBody
      +Validate()
  }
  class QueryMatcher {
      +string key
      +any value
  }
  class BodyMatcher {
      +string key
      +any value
  }
  class HeaderMatcher {
      +string key
      +any value
  }
```

### Validations

- Mock.method
  - Not empty
  - Valid http method
- Mock.path xor Mock.regexPath
  - Not empty
- Mock.regexPath
  - valid RegExp
- Mock.responseStatus
  - Not empty
  - Valid http status
- Mock.BodyMatcher
  - key: valid JsonPath
- HeaderMatcher | QueryMatcher | BodyMatcher
  - Both fields required if present

## Config

### Endpoints

- POST /config

  - ResponseStatus: 201
  - RequestBody:

    ```json
    {
      "method": "GET",
      "path": "/foo/bar",
      "response": {
        "status": 418,
        "body": { "foo": "bar", "baz": 1 }
      }
    }
    ```

- GET /config?method=GET&path=/foo/bar

  - ResponseStatus: 200
  - ResponseBody:

    ```json
    {
      "method": "GET",
      "path": "/foo/bar",
      "response": {
        "status": 418,
        "body": { "foo": "bar", "baz": 1 }
      }
    }
    ```

- GET /config/list

  - ResponseBody:

    ```json
    [
      {
        "id": 1,
        "method": "GET",
        "path": "/asd/asd",
        "response": {
          "status": 200,
          "body": { "asd": "zxc", "qwe": 1 }
        }
      },
      {
        "id": 2,
        "method": "GET",
        "path": "/asd/asd",
        "response": {
          "status": 200,
          "body": { "asd": "zxc", "qwe": 1 }
        }
      },
      {
        "id": 3,
        "method": "GET",
        "path": "/asd/qwe",
        "response": {
          "status": 418,
          "body": { "asd": "zxc", "qwe": 1 }
        }
      },
      {
        "id": 4,
        "method": "POST",
        "path": "/zxc",
        "response": {
          "status": 201,
          "body": { "asd": "zxc", "qwe": 1 }
        }
      }
    ]
    ```

- DELETE /config?id=1

  - ResponseStatus: 200

- GET /config/import

  - ResponseStatus: 200
  - ResponseBody : []string - List of imported files
    > Import directory `.import/`

- GET /config/export
  - ResponseStatus: 200
  - ResponseBody : []string - List of exported files
    > Export directory `.export/`

## Examples

### Not matched

- GET /foo/baz
  - ResponseStatus: 418
  - RequestBody: {}

### Simple path

- POST /config
  - ResponseStatus: 201
  - RequestBody:

    ```json
    {
      "method": "GET",
      "path": "/foo/bar",
      "response": {
        "status": 418,
        "body": { "foo": "bar", "baz": 1 }
      }
    }
    ```

- GET /foo/bar

  - ResponseStatus: 418
  - ResponseBody:

    ```json
    {
      "foo": "bar",
      "baz": 1
    }
    ```

### RequestBody Matching

- POST /config
  - ResponseStatus: 201
  - RequestBody:

    ```json
    {
      "method": "POST",
      "path": "/person",
      "requestBodyMatchers": [{ "key": "$.firstName", "value": "John" }],
      "response": {
        "status": 418,
        "body": {
          "firstName": "John",
          "lastName": "doe",
          "age": 26,
          "address": {
            "streetAddress": "naist street",
            "city": "Nara",
            "postalCode": "630-0192"
          },
          "phoneNumbers": [
            { "type": "iPhone", "number": "0123-4567-8888" },
            { "type": "home", "number": "0123-4567-8910" }
          ]
        }
      }
    }
    ```

- POST /person
  - RequestBody:

    ```json
    {
      "firstName": "John"
    }
    ```

  - ResponseStatus: 418
  - ResponseBody:
    ```json
    {
      "firstName": "John",
      "lastName": "doe",
      "age": 26,
      "address": {
        "city": "Nara",
        "postalCode": "630-0192",
        "streetAddress": "naist street"
      },
      "phoneNumbers": [
        { "number": "0123-4567-8888", "type": "iPhone" },
        { "number": "0123-4567-8910", "type": "home" }
      ]
    }
    ```

### RequestQuery Matching

- POST /config
  - ResponseStatus: 201
  - RequestBody:

    ```json
    {
      "method": "GET",
      "path": "/person",
      "requestQueryMatchers": [{ "key": "id", "value": 1 }],
      "response": {
        "status": 418,
        "body": {
          "firstName": "John",
          "lastName": "doe",
          "age": 26,
          "address": {
            "streetAddress": "naist street",
            "city": "Nara",
            "postalCode": "630-0192"
          },
          "phoneNumbers": [
            { "type": "iPhone", "number": "0123-4567-8888" },
            { "type": "home", "number": "0123-4567-8910" }
          ]
        }
      }
    }
    ```

- GET /person?id=1
  - ResponseStatus: 418
  - ResponseBody:
    ```json
    {
      "firstName": "John",
      "lastName": "doe",
      "age": 26,
      "address": {
        "city": "Nara",
        "postalCode": "630-0192",
        "streetAddress": "naist street"
      },
      "phoneNumbers": [
        { "number": "0123-4567-8888", "type": "iPhone" },
        { "number": "0123-4567-8910", "type": "home" }
      ]
    }
    ```

### RequestHeader Matching

- POST /config
  - ResponseStatus: 201
  - RequestBody:

    ```json
    {
      "method": "GET",
      "path": "/fooHeader",
      "requestHeaderMatchers": [{ "key": "foo", "value": "false" }],
      "response": {
        "status": 418,
        "body": {
          "foo": false,
          "bar": [{ "baz": 3 }, { "baz": 4 }]
        }
      }
    }
    ```

- GET /fooHeader
  - RequestHeaders
    - foo: false
  - ResponseStatus: 418
  - ResponseBody:
    ```json
    {
      "foo": false,
      "bar": [{ "baz": 3 }, { "baz": 4 }]
    }
    ```

### Regexp Matching

- POST /config
  - ResponseStatus: 201
  - RequestBody:

    ```json
    {
      "method": "GET",
      "regexPath": "\/foo\/bar\/\\d+\/baz",
      "response": {
        "status": 200,
        "body": {
          "foo": true,
          "bar": {
            "id": 123,
            "baz": {
              "zxc": 18
            }
          }
        }
      }
    }
    ```

- GET /foo/bar/123/baz
  - ResponseStatus: 200
  - ResponseBody:
    ```json
    {
      "foo": true,
      "bar": {
        "id": 123,
        "baz": {
          "zxc": 18
        }
      }
    }
    ```

### Proxy

- POST /config
  - ResponseStatus: 201
  - RequestBody:
    ```json
    {
      "method": "GET",
      "regexPath": "\/products\/\\d+",
      "response": {
        "proxy": {
          "enabled": true,
          "hostUrl": "https://fakestoreapi.com"
        }
      }
    }
    ```

- GET https://fakestoreapi.com/products/1
  - ResponseStatus: 200
  - ResponseBody: `Actual response body`

