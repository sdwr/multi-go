to run:
on heroku: just deploy. auto-builds the whole module
locally: 
	go build
	go run game.go coordinator.go server.go <loglevel>
	access from localhost:4404