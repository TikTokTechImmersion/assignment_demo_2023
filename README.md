# assignment_demo_2023

![Tests](https://github.com/TikTokTechImmersion/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

This is a demo and template for backend assignment of 2023 TikTok Tech Immersion.

![whiteboard_exported_image](https://github.com/riyandrika/8-D/assets/26215621/5f25945c-e5e3-411f-81a2-ece38ab94931)

## Goals of Assignment
1. Send API
```bash=
# send "hi" from a to b
curl -X POST 'localhost:8080/api/send?sender=a&receiver=b&text=hi'

# pull messages from the chat of a and b
curl 'localhost:8080/api/pull?chat=a%3Ab'
```

The API call is made to the HTTP server, which subsequently interfaces with the RPC server via RPC. <b><i>Send</i></b> is performed using a POST request with the following parameters in the request body.
#### Send Request Body
| Field Name | Description                                                               |
|------------|---------------------------------------------------------------------------|
| Chat       | Identifier for the chat room. Ordered in the format ```sender1:sender2``` |
| Text       | Message between participants of the chat                                  |
| Sender     | Participant that is sending the message                                   |
#### Send Response Body