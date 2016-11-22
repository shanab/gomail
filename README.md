# gomail

A reliable, scalable and fault-tolerant email service.

Gomail queues in incoming emails, and dispatches them to one of multiple third party email services - taking into account the health status for each of these services. Currently we support AWS SES and Sendgrid.

## Architecture
```
            +---------------------------+
            |                           |
            | Web client sending emails |
            |                           |
            +-------------+-------------+
                          |
                          v
                  +---------------+
                  |               |
                  | Load Balancer |
                  |               |
                  +---+-------+---+                     
                      |       |
                  +---+       +---+
                  |               |
                  v               v
           +------+-----+     +---+--------+
           |            |     |            |
           | Gomail API |     | Gomail API |
           |            |     |            |
           +------+-----+     +------+-----+
                  |                  |
 SQS Queue        v                  |
------------------+----+------------------------+-->
                       ^             |          ^
     SQS Queue         |             v          |
  -----------------+-----------------+------+---------->
                   ^   |                    ^   |
                   |   |                    |   |
 +-----------------+---+----+    +----------+---+-----------+
 |    Gomail Pipeline       |    |    Gomail Pipeline       |
 |                          |    |                          |
 | +---------+ +----------+ |    | +---------+ +----------+ |
 | |Sendgrid | | SES      | |    | |Sendgrid | | SES      | |
 | |Worker   | | Worker   | |    | |Worker   | | Worker   | |
 | |         | |          | |    | |         | |          | |
 | +---------+ +----------+ |    | +---------+ +----------+ |
 +--------------------------+    +--------------------------+
```
Gomail is composed of 3 separate components:

1. Web Client
2. Gomail API
3. Gomail Pipeline

### Web Client

The web client is a static Javascript client developed using [React](https://facebook.github.io/react/) & [Redux](http://redux.js.org/). Gomail web client's main responsibility is to submit emails from the web interface to the API. It is also responsible for displaying any errors the API sends back (e.g. validation errors or service unavailability errors).

### Gomail API

Gomail API exposes a single endpoint that forwards the email details to one of multiple [Amazon SQS](https://aws.amazon.com/sqs/) queues for later processing.

#### Usage

``` shell
./api -config=/path/to/config.yaml            # defaults to ./config.yaml

```

#### Endpoints

* `POST /email/send`.

**Required parameters**: `fromEmail` as the sender email, `toEmail` as the receiver email, and `body` as the content of the email.

_Optional parameters_: `fromName` as the sender name, `toName` as the receiver name, and `subject` as the email subject.

This endpoint can return:

* `200 OK` if the request succeeds.
* `400 Bad Request` if the JSON body was malformed or exceeds the maximum body size (configurable via config file).
* `422 Unprocessable Entity` if request validation failed.
* `503 Service Unavailable` if the request to SQS returned an error.

###### Example JSON Request
``` json
{
    "email": {
        "fromEmail": "from@example.com",
        "fromName": "From Name",
        "toEmail": "to@example.com",
        "toName": "To Name",
        "subject": "Test subject",
        "body": "Test body"
    }
}
```

###### Example JSON Response (success)
``` json
{
    "messageId": "12345678910"
}
```

###### Example JSON Response (validation failure)
``` json
{
    "errors": {
        "base": "Service unavailable",
        "fromEmail": "From email is not a valid email",
        "body": "Body is required",
    }
}
```

#### API configuration file

API uses a simple YAML config file to specify the port it should run on, the maximum body size, etc. You can find a sample config file under `api/config.yaml.example`.
