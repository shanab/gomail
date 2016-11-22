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

* [Web Client](#web-client)
* [Gomail API](#gomail-api)
* [Gomail Pipeline](#gomail-pipeline)

### Web Client

The web client is a static Javascript client developed using [React](https://facebook.github.io/react/) & [Redux](http://redux.js.org/). Gomail web client's main responsibility is to submit emails from the web interface to the API. It is also responsible for displaying any errors the API sends back (e.g. validation errors or service unavailability errors).

[click on this link](#my-multi-word-header)

### My Multi Word Header

### Gomail API

Gomail API exposes a single endpoint that forwards the email details to one of multiple [Amazon SQS](https://aws.amazon.com/sqs/) queues for later processing.

#### Usage

Note: You can find an example of the configuration file under `api/config.yaml.example`

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

### Gomail Pipeline

Pipeline's design is similar to a load balancer, where it reads messages from all SQS queues, and splits them between the 2 workers based on each worker's health status. This is how it works:

1. Each n seconds (configurable) pipeline loops over all SQS queues to get an estimate of the number of available messages for each queue.
2. Since SQS only allows a maximum of 10 messages per receive request, pipeline divides the number of messages in each queue by 10 to get an estimate of the number of readers it needs to potentially receive all messages on the queue at this point in time.
3. It runs all readers concurrently - where each reader tries to receive 10 messages from the queue.
4. Once all readers are done, all results are aggregated into a single slice that is going to be split between the 2 workers (SES worker & Sendgrid worker).
5. If the 2 workers are healthy - and they initially are - the slice of messages is split evenly between both workers.
6. Each worker sends all messages concurrently using its corresponding service API. Any failure is recorded, and the total number of failures per worker is then forwarded to the pipeline to update the health status of both workers.
7. If a worker fails to send a single message for n consecutive iterations, and n is greater than the unhealthy threshold, the worker is marked as unhealthy.
8. If a worker is healthy and the other worker is unhealthy, the unhealthy worker takes 1 message only to act as a health check (see if the worker is still unhealthy). The healthy worker takes all the rest of the messages.
9. In the unfortunate incident where both workers are unhealthy, messages are split between them equally again until one of them becomes healthy (successfully sends messages for n consecutive iterations, where n > the healthy threshold).
10. Pipeline sleeps for a configurable duration, then goes back to step 1.

#### Usage

Note: You can find an example of the configuration file under `pipeline/config.yaml.example`

``` shell
./pipeline -config=/path/to/config.yaml            # defaults to ./config.yaml

```

