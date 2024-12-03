const petBtn = document.getElementById('pet-henry-btn')
const sendChatBtn = document.getElementById('send-chat-btn')
const changeDisplayNameBtn = document.getElementById('change-name-btn')
const chatContent = document.getElementById('chat-content')
const personalCounter = document.getElementById('personal-counter')

const ws = new WebSocket("ws://localhost:8080/ws");
console.log(personalNumber)

ws.onopen = () => {
	console.log("Connected to the server");
	connectMessage = {
		action: "connect",
		userID: "nathan",
	}
	ws.send(JSON.stringify(connectMessage));
};

ws.onmessage = (event) => {
	console.log("Received from server:", event.data);
	
	const data = JSON.parse(event.data)	
	if (data.action == "counter") {
		console.log(`New pet count: ${data.value}`)
		document.getElementById('pet-counter').textContent = `Henry has been pet ${data.value} times!`
	} else if (data.action == "init") {
		console.log(`Personal pet count: ${data.value}`)
	}
};

ws.onclose = () => {
	console.log("Connection closed");
};

ws.onerror = (error) => {
	console.error("WebSocket error:", error);
};

petBtn.addEventListener("click", () => {
	console.log("pet");
	petMessage = {
		action: "pet",
		userID: "test-1",
	}

	personalNumber++
	console.log(personalNumber)
	personalCounter.textContent = `You have pet him ${personalNumber} times!`


	ws.send(JSON.stringify(petMessage));

})

sendChatBtn.addEventListener("click", () => {

	chatMessage = {
		action:	"chat",
		userID: "nathan",
		content: chatContent.value,
	}

	ws.send(JSON.stringify(chatMessage))
	chatContent.value = ""
})

changeDisplayNameBtn.addEventListener("click", changeDisplayName)

function changeDisplayName() {

	let name = prompt("Enter your new display name: ", `${myDisplayName}`)

	fetch('/cd', {
		method: 'POST',
		body: JSON.stringify({ displayName: name })
	})
	.then(response => response.json())
	.then(data => {
		console.log(data)
		if (!data.success) { alert("Invalid name!") }
		else { myDisplayName = name; updateDisplayName() }
	})
}

function updateDisplayName() {
	document.getElementById('greeting').textContent = `Welcome, ${myDisplayName}! We're glad to have you here.`
}
