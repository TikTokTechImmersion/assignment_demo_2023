// API for pull mode IM service.
namespace go rpc

struct Message {
    1: string Chat   // format "<member1>:<member2>", e.g. "john:doe"
    2: string Text   // message text content
    3: string Sender // sender identifier
    4: i64 SendTime  // unit: microseconds
}

struct SendRequest {
    1: required Message message // message to be sent
}

struct SendResponse {
    1: required i32 Code   // zero for success, non-zero for failures
    2: required string Msg // prompt information
}

struct PullRequest {
    1: required string Chat  // format "<member1>:<member2>", e.g. "john:doe"
    2: required i64 Cursor   // starting position of message's send_time, inclusively, 0 by default
    3: required i32 Limit    // the maximum number of messages returned per request, 10 by default
    4: optional bool Reverse // if false, the results will be sorted in ascending order by time
}

struct PullResponse {
    1: required i32 Code   // zero for success, non-zero for failures
    2: required string Msg // prompt information
    3: optional list<Message> Messages
    4: optional bool HasMore   // if true, can use next_cursor to pull the next page of messages
    5: optional i64 NextCursor // starting position of next page, inclusively
}

service IMService {
    SendResponse Send(1: SendRequest req)
    PullResponse Pull(2: PullRequest req)
}
