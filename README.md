# assignment_demo_2023

![Tests](https://github.com/TikTokTechImmersion/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

This is a demo and template for backend assignment of 2023 TikTok Tech Immersion.

![whiteboard_exported_image](https://github.com/riyandrika/8-D/assets/26215621/5f25945c-e5e3-411f-81a2-ece38ab94931)

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
To determine the boolean <i>HasMore</i> as required in the response body, an extra message will be pulled, hence the value of <i>limit</i> will be incremented by one. 

#### If limit + 1 messages are pulled,
HasMore = true
#### else
HasMore = false