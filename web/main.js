
const SERVER_URL = "localhost:4404"
const socket = new WebSocket("ws://"+SERVER_URL+"/socket");

//GLOBALS
let queued = false;
let queueStart = null;

//SOCKET
socket.onopen = function(e) {
	console.log("socket connection open")
}

socket.onmessage = function(e) {
	let message = JSON.parse(e.data);
	if(message.Type === "UPDATE") {
		updateStones(message.Payload)
	} else if(message.Type === "FOUNDMATCH") {
		closeMenu()
	}
}

//MESSAGES
function sendMessage(message) {
	socket.send(JSON.stringify(message));
}

function sendClickMessage(x, y) {
	let clickMessage = {}
	clickMessage.Type = "CLICK"
	sendMessage(clickMessage)
}

function sendFindMatchMessage(isBot) {
	let findMatchMessage = {}
	findMatchMessage.Type = "FINDMATCH"
	findMatchMessage.IsBotMatch = isBot"
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


playWithBotsButton.onclick = searchGame(true)
queueButton.onclick = searchGame(false)
buttonLeaveQueue.onclick = leaveQueue()

function openQueue() {
	queued = true
	queueStart = Date.now()
	queueTimer()
	menuButtons.setAttribute("display", "none")
	menuQueue.setAttribute("display", "block")
}

async function queueTimer() {
	while(queued) {
		menuQueueTimer.textContent = "seconds in queue: " + (Date.now() - queueStart) / 1000
		await new Promise(r => setTimeout(r, 1000)
	}
}

function closeQueue() {
	queued = false
	menuQueue.setAttribute("display","none")
	menuButtons.setAttribute("display", "block")
}

function openMenu() {
	closeQueue()
	menuElement.setAttribute("display","block")
}

function closeMenu() {
	closeQueue()
	menuElement.setAttribute("display", "none")
}

function searchGame(isBot) {
	sendFindMatchMessage(isBot)
	openQueue()
}

function leaveQueue() {
	sendCancelMatchMessage()
	closeQueue()
}


let boardElement = document.getElementById("board")
let toolElement = document.getElementById("tool")
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
