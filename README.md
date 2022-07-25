# Talkoger

A prototype of an application to store and view conversations in real time

This application is similar to a logger. The name of this application is derived from it, Talkog is Talk + Logger.

## Actors

- Client (HoloLens 2)
  - Convert the audio of conversations into text and send it to the server
- Client (Web Client)
  - View the text saved on the server
- Server (Amazon API Gateway, AWS Lambda, Amazon DynamoDB)
  - Save and send text

```mermaid
  graph LR

  c1((HoloLens 2))
  c2((Web Client))
  s1{Amazon API Gateway}
  s2["AWS Lambda"]
  s3[(Amazon DynamoDB)]

  c1 -- save --> s1
  c2 -- fetch --> s1
  s1 -- push --> c2

  subgraph AWS
  s1 --> s2
  s2 --> s1
  s2 --> s3
  s3 --> s2
  end
```

## Communication

Use WebSocket.

### Save a talk

Send data in the following format.

```json
{
  "action": "saveTalk",
  "UserId": "19F095D0-911F-4FAD-B43C-FF06A8E91020", // UUID
  "Talk": "Nice to meet you."
}
```

The response is a status code only.

### Fetch talkogs

Send data in the following format.

```json
{
  "action": "fetchTalkogs",
  "UserId": "19F095D0-911F-4FAD-B43C-FF06A8E91020" // UUID
}
```

The following data will be sent in real time.

```json
{
  "UserId": "19F095D0-911F-4FAD-B43C-FF06A8E91020", // UUID
  "Timestamp": 1657921427, // Unix time
  "Talk": "Nice to meet you."
}
```
