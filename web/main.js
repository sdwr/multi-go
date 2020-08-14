
let SERVER_URL = window.location.href
SERVER_URL.replace("^https\/\/$", "")
const socket = new WebSocket("wss://"+SERVER_URL+"socket");

//GLOBALS
let queued = false;
let queueStart = null;

let player = null;
let cooldowns = new Map()
let cooldownElements = new Map()

let lastUpdated = Date.now()
loopCooldowns()

//SOCKET
socket.onopen = function(e) {
	console.log("socket connection open")
}

socket.onmessage = function(e) {
	let message = JSON.parse(e.data);
	if(message.Type === "UPDATE") {
		console.log("recieved update", message)
		updateScores(message.Payload.Players)
		updateStones(message.Payload)
		updateCooldowns(message.Payload.Players)
	} else if(message.Type === "GAMESTART") {
		console.log("game start", message)
		player = message.Payload.Player
		closeMenu()
		createCDElements(message.Payload.Players)
	}
}

function loopCooldowns(){
	timeCooldowns()
	setTimeout(loopCooldowns, 100)
}

function timeCooldowns() {
	millis = timeElapsed()
	for (let k of cooldowns.keys()) {
		let cd = cooldowns.get(k)
		cd -= millis
		cooldowns.set(k, cd)
		let ele = cooldownElements.get(k)
		updateCDElement(ele, cd)
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
	cancelMessage.Type = "CANCELMATCH"
	sendMessage(cancelMessage)
}

//OVERLAY MENU
const menuElement = document.getElementById("menu")
const menuButtons = document.getElementById("menu-buttons")
const playWithBotsButton = document.getElementById("button-play-bots")
const queueButton = document.getElementById("button-queue")

const menuQueue = document.getElementById("menu-queue")
const menuQueueTimer = document.getElementById("menu-queue-timer")
const buttonLeaveQueue = document.getElementById("button-leave-queue")

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
		menuQueueTimer.textContent = "seconds in queue: " + (Date.now() - queueStart) / 1000
		await new Promise(r => setTimeout(r, 1000))
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

function updateCooldowns(players) {
    players.forEach(p => {
	cooldowns.set(p.ID,p.Cooldown)
	ele = cooldownElements.get(p.ID)
	updateCDElement(ele, p.Cooldown)
    })
}

function createCDElements(players) {
	players.forEach(p => {
		cooldownElements.set(p.ID, createCDElement(p))
	})
}

function createCDElement(player) {
	let ele = document.createElement("div")
	ele.className += " cooldown"
	let circle = document.createElement("div")
	circle.className += " cooldownCircle"
	circle.style.background = player.Color
	let over = document.createElement("div")
	over.className += " cooldownOver"
	ele = ele.appendChild(circle)
	ele = ele.appendChild(over)
	ele = updateCDElement(ele, player.Cooldown)
	return ele;
}

function updateCDElement(ele, amt) {
	if(ele && ele.children[1]) {
		let over = ele.children[1]
		over.style.height= calcCDOverlayHeight(amt)
	}
	return ele
}

function calcCDOverlayHeight(amt) {
	return "" + (amt/10000)*100 + "%"
}

function updateScores(players) {
	players.sort(sortPlayers)
	scores.innerHTML = ""
	players.forEach(p => {
		let li = document.createElement("div")
		li.className += " score"
		li.appendChild(getScoreElement(p))
		let ele = cooldownElements.get(p.ID)
		ele = li.appendChild(ele)
		cooldownElements.set(p.Id, ele)
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



let boardElement = document.getElementById("board")
let board = new WGo.Board(boardElement, {width:600})

//BOARD API LETS GOOOO
let players = ["#bbbbbb", "#abcabc", "#defdef"]

function updateStones(payload) {
	addStone(payload.Move.Coords.X, payload.Move.Coords.Y, payload.Move.Player.Color)
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
