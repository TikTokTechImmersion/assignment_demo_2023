# assignment_demo_2023

![Tests](https://github.com/TikTokTechImmersion/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

This is the completed demo for backend assignment of 2023 TikTok Tech Immersion done by Aaron Sng (Find me on Linkedin: http://linkedin.com/in/aaronsng ;))

## Installation

Requirement:

- golang 1.18+
- docker

To install dependency tools:

```bash
make pre
```

## Run

```bash
docker-compose up -d
```

Check if it's running:

```bash
curl localhost:8080/ping
```

![whiteboard_exported_image](https://github.com/TikTokTechImmersion/assignment_demo_2023/assets/26215621/a25c262e-e2a5-4163-abe1-ca0abf50c029)

## 1. Send API
```bash=
# send "hi" from a to b
curl -X POST 'localhost:8080/api/send?sender=a&receiver=b&text=hi'
```

The API call is made to the HTTP server, which subsequently interfaces with the RPC server via RPC. <b><i>Send</i></b> is performed using a POST request with the following parameters in the request body.
### Send Request Body
| Field Name | Description                                                               |
|------------|---------------------------------------------------------------------------|
| Chat       | Identifier for the chat room. Ordered in the format ```sender1:sender2``` |
| Text       | Message between participants of the chat                                  |
| Sender     | Participant that is sending the message                                   |
### Send Response Body
| Field Name | Description                                 |
|------------|---------------------------------------------|
| Code       | Code describing the status of the HTTP call |
| Msg        | Message detailing the code                  |

As depicted in the image shown above, the HTTP <b><i>Send</i></b> requests are sent directly to the RPC server over remote procedure calls (RPC). The RPC server receives this and interfaces with a database server. In this case, the RPC server uses a MongoDB server. The choice of MongoDB was based on the following reasons:
1. Non structured database, permitting dynamic schemas which may allow for images and other media to be stored, which might change in the future
2. Mongo's data model allow for rather efficient querying of information stored in the database, and as well as the storage of the data
3. I want to learn and get a bit of experience with MongoDB :D

### 1.1 Database Design
Well, there's no schema since MongoDB is used. However, there's a 'design' for each entry of the MongoDB. 

#### 1.1.1 Fields in each row
| Field Name | Description                                 |
|---------|---------------------------------------------|
| ObjectID | Unique identifier, automatically created by MongoDB |
| ChatRoom | Identifier for the ChatRoom, as shown in the Send Request Body, the ChatRoom is formatted as such ```sender1:sender2```, where ```sender1``` is lexicographically smaller than ```sender2``` |
| Message | Actual message in a string-like object |
| Sender | Identity of who sent the message |
| SendTime | Time when the message reached the RPC server. Formatted in UNIX Epoch Time. |

Each message that is sent to the database. To perform the ordering for ChatRoom, the function ```ObtainChat``` was used. ```ObtainChat``` uses the delimiter ':' to separate ```sender1``` and ```sender2```. The rough function is as shown below
```go=
func (c *MongoClient) ObtainChat(chat string) string {
	// Ensure that the chat ID is a consistent form
	// The lexicographically smaller party will be the first followed by the other party
	chat = strings.ToLower(chat)
	parties := strings.Split(chat, ":")
	var finalChatID string

	if parties[0] < parties[1] {
		finalChatID = parties[0] + ":" + parties[1]
	} else {
		finalChatID = parties[1] + ":" + parties[0]
	}

	return finalChatID
}
```

## 2. Pull API
```bash=
# pull messages from the chat of a and b
curl 'localhost:8080/api/pull?chat=a%3Ab'
```

The API call is made to the HTTP server, which subsequently interfaces with the RPC server via RPC. <b><i>Pull</i></b> is performed using a GET request with the following parameters in the request body.

### Pull Request Body
| Field Name | Description                                                               |
|------------|---------------------------------------------------------------------------|
| Chat       | Identifier for the chat room. Ordered in the format ```sender1:sender2``` |
| Cursor       | Next message to pull |
| Limit | Maximum number of messages to pull at a time from the database |
| Reverse     | Whether to display the messages in a reversed order |
### Pull Response Body
| Field Name | Description                                 |
|------------|---------------------------------------------|
| Code       | Code describing the status of the HTTP call |
| Msg        | Message detailing the code                  |
| Messages | An array of messages given the ChatRoom |
| HasMore | Whether there are more messages to pull other than given the value of the cursor |
| NextCursor | Next message to pull after all the different messages that have been pulled |

To make the query from the database, Mongo's data model was employed to retrieve the messages based on the <b><i>Pull</i></b> request body. To make queries based on the fields Chat, Cursor, Limit and Reverse, the <i>match</i>, <i>skip</i>, <i>limit</i> and <i>sort</i> query operations were used respectively.

#### Determining HasMore
To determine the boolean <i>HasMore</i> as required in the response body, an extra message will be pulled, hence the value of <i>limit</i> will be incremented by one. The following logic then determines whether HasMore is true or false.

#### If limit + 1 messages are pulled,
HasMore = true
#### else
HasMore = false

## 3. Tests

JMeter was used to stress test the Dockerised implementation. Using the script provided by <a href="http://www.linkedin.com/in/weixingp">Wei Xing</a>, the IM system under 1000 and 5000 users was investigated.

![one_thousand_users](https://github.com/TikTokTechImmersion/assignment_demo_2023/assets/26215621/bd46e8d3-52bc-4231-925c-770064a30868)

The system was able to handle 1000 users relatively easily. A maximum of 310ms response was reported.

![five_thousand_users](https://github.com/TikTokTechImmersion/assignment_demo_2023/assets/26215621/affaff74-a111-4fb8-a7ab-4d8299e3f4d6)

However, the system wasn't able to keep up with 5000 users. A significant high error rate of 54.55% was reported. 

## 4. Elastic Deployment
To support elastic deployment, the IP addresses of the MongoDB and RPC server were obtained via environment variables. This is to allow support for Kubernetes services, which identifies the IP addresses of individual pods and nodes in the Kubernetes cluster.

The Kubernetes configuration files are also hosted on the following github repo: https://github.com/aaronsng/kubernetes-demo-2023. To apply and realise the Kubernetes cluster, run the following steps in terminal.

```bash=
# Apply the secrets of the Kubernetes cluster
kubectl apply -f ./secret 

# Launch the mongodb database next
kubectl apply -f ./database

# Finally the rest of the Kubernetes cluster
kubectl apply -f ./app
```

In the elastic deployment of the IM service, the database will only have one replica. This is done to avoid potential race conditions. I'll be creating 4 replicas for ```http-server``` and ```rpc-server```. The elastic deployment was tested similarly with 1000 and 5000 users, and the results are as follows. The deployment was performed on a local computer using minikube.

![one_thousand_multiple](https://github.com/TikTokTechImmersion/assignment_demo_2023/assets/26215621/dfde73fc-a2af-44d5-afbe-914d4cf67177)

It appears that the overall performance of the cluster has dropped. The lack of parallelism from having a single computer to host all 8 pods may not be ideal in providing the most optimal performance for the IM service.

![five_thousand_multiple](https://github.com/TikTokTechImmersion/assignment_demo_2023/assets/26215621/fba0c7a8-88c8-4692-be9b-2c51fc124d7b)

The performance dropped further for the elastic case. The system might be different if it were to be deployed on the cloud instead, with multiple different computers used.
