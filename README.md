# assignment_demo_2023

<!--![Tests](https://github.com/cheeheng/assignment_demo_2023/actions/workflows/test.yml/badge.svg)-->

This is my backend assignment submission of TikTok Tech Immersion Programme 2023. It simulates the backend of an instant-messaging system. 

<h2>How to setup the project</h2>

1. Download the source code.
2. Install Docker, PostgreSQL and Go 1.20 or later 
3. Change the user and password regarding PostgreSQL database connection in docker-compose.yml and rpc-server/handler.go 
4. Create databases assignment_demo_2023 and assignment_demo_2023_test in the local PostgreSQL instance.
5. Build the project using ```docker-compose build``` and run using ```docker-compose up -d```

To build the project, on the root directory of the project, execute the following command on the terminal:
```
docker-compose build 
```

To run the project, on the root directory of the project, execute the following command on the terminal:
```
docker-compose up -d
```

To stop the project, on the root directory of the project, execute the following command on the terminal:
```
docker-compose down
```

<h2>How to use the backend API</h2>

To send a message from [sender] to [receiver], the command format (in the terminal) is as follows:

```
curl -X POST 'localhost:8080/api/send?sender=[sender]&receiver=[receiver]&text=[message]'
```

Replace [sender] with the sender, [receiver] with the receiver, and [message] with the actual text of the message. The parameters can appear in any order.

To retrieve what is in the chat, one can use the following command:

```
localhost:8080/api/pull?chat=[chat]&cursor=[cursor]&limit=[limit]&reverse=[reverse]
```

Replace [chat], [cursor], [limit] and [reverse] with the appropriate values accordingly. Note that [cursor], [limit] and [reverse] are optional, with default values of 0 (i.e. starting from first message), 10 (i.e. at most 10 messages returned) and false (i.e. starting from the newest message) respectively. The parameters can appear in any order.

This request returns a list of up to [limit] messages from the conversation between the two members in [chat]. The return value in the list returns the ([cursor]+1)th message to the ([cursor]+[limit])th message (both inclusive, if a message of that index exists) when the messages are sorted in descending order (ascending order if [reverse] is set to true) of sending time.

- [chat] refers to the two users who are currently using the application. It must be in the format [member1]:[member2], where[member1] and [member2] are the two people whose conversation you would want to retrieve. Note that [member1]:[member2] and [member2]:[member1] are equivalent. 

- [cursor] represents the number of messages to skip, sorted in descending order of sending time (ascending order if [reverse] is set to true). In other words, return value starts from the ([cursor]+1)th newest message (earliest if [reverse] is true).

- [limit] represents the maximum number of results to be returned. 

- [reverse] is either true or false. If true, the messages are sorted from oldest to newest. If false (default), the messages are returned from newest to oldest.

<h2>How to run tests</h2>

1. To run RPC Server tests, change directory to ```rpc-server```, then run ```go test```.
2. To run performance tests, install k6 software, then change directory to ```performance-tests```, then run ```k6 run load-test.js```.