# dtcg-server

## Description

This project is [dtcg game](https://github.com/MrToy/dtcg) server, write by go

## How to run

1. Install [golang](https://go.dev/) lastest version on your computer, and make sure ```go``` command in your environment
2. Install project dependencies with ```go mod install``` on project dictionary
3. Run server

```bash
cd ./cmd/server
go run .
```

4. Open another terminal, add ai to fight with you

```bash
cd ./cmd/client
go run .
```


## Design

* Use **tcp** protocol to communicate with unity
* Default port is 2333,  you can change it in ./cmd/server/main.go
* Each player uses a coroutine to handle events
* Each game uses a coroutine to handle game logic
* The main logic of the game is in the ./app/game.go
