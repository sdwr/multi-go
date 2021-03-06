
let SERVER_URL = window.location.href
SERVER_URL = SERVER_URL.replace(/^https:\/\//, "")
SERVER_URL = SERVER_URL.replace(/^http:\/\//, "")
let socket = "";
if(SERVER_URL.startsWith("localhost")) {
	socket = new WebSocket("ws://"+SERVER_URL+"socket");
} else {
	socket = new WebSocket("wss://"+SERVER_URL+"socket");
}
//GLOBALS
//note: some div ids are hardcoded in here
//scores, cooldown, others??
//do not change on html file
let queued = false;
let queueStart = null;

let player = null

//SOCKET
socket.onopen = function(e) {
	console.log("socket connection open")
}

socket.onmessage = function(e) {
	let message = JSON.parse(e.data);
        if(message.Type === "INIT") {
	        console.log("init", message)
        } else if(message.Type === "UPDATE") {
		console.log("recieved update", message)
		updateScores(message.Payload.Players)
		updateStones(message.Payload)
	} else if(message.Type === "GAMESTART") {
		console.log("game start", message)
		player = message.Payload.Player
		closeMenu()
	} else if(message.Type === "QUEUEUPDATE") {
		updateQueue(message.Payload.Queued)
	}
}

function timeElapsed() {
	let elapsed = Date.now() - lastUpdated
	lastUpdated = Date.now()
	return elapsed
}

//MESSAGES
function sendMessage(message) {
	socket.send(JSON.stringify(message));
}

function sendClickMessage(x, y) {
	let clickMessage = {}
	clickMessage.Type = "CLICK"
	clickMessage.Payload = {Move:{Coords:{X:x,Y:y}, Player:player}}
	console.log("sent click", clickMessage)
	sendMessage(clickMessage)
}

function sendFindMatchMessage(isBot) {
	let findMatchMessage = {}
	findMatchMessage.Type = "QUEUE"
	findMatchMessage.IsBotMatch = isBot
	sendMessage(findMatchMessage)
}

function sendCancelMessage() {
	let cancelMessage = {}
	cancelMessage.Type = "LEAVEQUEUE"
	sendMessage(cancelMessage)
}

//OVERLAY MENU
const menuElement = document.getElementById("menu")
const menuButtons = document.getElementById("menu-buttons")
const playWithBotsButton = document.getElementById("button-play-bots")
const queueButton = document.getElementById("button-queue")
const queuePlayers = document.getElementById("menu-queue-players")

const menuName = document.getElementById("input-name")
const menuQueue = document.getElementById("menu-queue")
const menuQueueTimer = document.getElementById("menu-queue-timer")
const buttonLeaveQueue = document.getElementById("button-leave-queue")

const cooldownElement = document.getElementById("cooldown")
const scores = document.getElementById("scores")

playWithBotsButton.onclick = function() { return searchGame(true)}
queueButton.onclick = function() { return searchGame(false)}
buttonLeaveQueue.onclick = function() { return leaveQueue()}

function openQueue() {
	queued = true
	queueStart = Date.now()
	queueTimer()
	menuButtons.style.display = "none"
	menuQueue.style.display = "block"
}

async function queueTimer() {
	while(queued) {
		menuQueueTimer.textContent = "seconds in queue: " + Math.floor((Date.now() - queueStart) / 1000)
		await new Promise(r => setTimeout(r, 1000))
	}
}

async function cooldownTimer() {
	while(true) {
		if(player && player.Cooldown > 0) {
			await new Promise(r => setTimeout(r, 100))
			player.Cooldown -= 100
		        updateCDElement(cooldownElement.children[0], player.Cooldown)	
		}
	}
}

function closeQueue() {
	queued = false
	menuQueue.style.display = "none"
	menuButtons.style.display = "block"
}

function openMenu() {
	closeQueue()
	menuElement.style.display = "block"
}

function closeMenu() {
	closeQueue()
	menuElement.style.display = "none"
}

function searchGame(isBot) {
	sendFindMatchMessage(isBot)
	openQueue()
}

function leaveQueue() {
	sendCancelMessage()
	closeQueue()
}

//safety to make sure it doesn't loop forever
function updateCDElement(amt, iter = 0) {
	let circle = document.getElementById("cooldownCircle")
	let over = document.getElementById("cooldownOver")
	over.style.height= calcCDOverlayHeight(amt)
	if(amt > 0 && iter < 10) {
		setTimeout(() => updateCDElement(amt-1, iter+1), 1000)
	}
}

function calcCDOverlayHeight(amt) {
	return "" + ((amt*1.0) / 5) * 100 + "%"
}

function updateScores(players) {
	players.sort(sortPlayers)
	scores.innerHTML = ""
	players.forEach(p => {
		let li = document.createElement("div")
		li.className += " score"
		li.style.color = p.Color
		li.appendChild(getScoreElement(p))
		scores.appendChild(li)
	});
}

function sortPlayers(p1, p2) {
	let p1Score = p1.Territory + p1.Captures
	let p2Score = p2.Territory + p2.Captures
	return p2Score - p1Score
}

function getScoreElement(player) {
	ele = document.createTextNode(player.Name + ": " + player.Territory + " + " + player.Captures)
	return ele
}

function updateQueue(num) {
	queuePlayers.innerHTML = "" + num + "/5 players"
}



let boardElement = document.getElementById("board")
let board = new WGo.Board(boardElement, {width:600})

//BOARD API LETS GOOOO
let players = ["#bbbbbb", "#abcabc", "#defdef"]

//player cd hardcoded here
function updateStones(payload) {
	addStone(payload.Move.Coords.X, payload.Move.Coords.Y, payload.Move.Player.Color)
	if(player && player.ID && payload.Move.Player.ID === player.ID) {
		updateCDElement(5)
	}
	payload.Remove.forEach(pos => {
		removeStone(pos.X,pos.Y)
	})
}
function addStone(x, y, color){
	board.addObject({x:x,y:y,type:drawFactory(color)})
}

function removeStone(x, y){
	board.removeObjectsAt(x,y)
}

function drawFactory(color) {
	return {
	stone: {
		draw: function(args, board) {
			let xr = board.getX(args.x)
			let yr = board.getY(args.y)
			let sr = board.stoneRadius;

			this.strokeStyle=color
			this.lineWidth = 5
			this.beginPath()
			this.arc(xr, yr, sr*0.9, 0, 2*Math.PI)
			this.stroke()
			this.fillStyle=color
			this.fill()
		}
	}
	}
}
	

	board.addEventListener("click", function(x, y){
		sendClickMessage(x, y)
	})
