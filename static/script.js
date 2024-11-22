petBtn = document.getElementById('pet-henry-btn')
sendChatBtn = document.getElementById('send-chat-btn')
chatContent = document.getElementById('chat-content')

const ws = new WebSocket("ws://localhost:8080/ws");

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
	}
};

ws.onclose = () => {
	console.log("Connection closed");
};

ws.onerror = (error) => {
	console.error("WebSocket error:", error);
};

petBtn.addEventListener("click", (e) => {
	console.log("pet");
	petMessage = {
		action: "pet",
		userID: "test-1",
	}

	ws.send(JSON.stringify(petMessage));
})

sendChatBtn.addEventListener("click", (e) => {

	chatMessage = {
		action:	"chat",
		userID: "nathan",
		content: chatContent.value,
	}

	ws.send(JSON.stringify(chatMessage))
	chatContent.value = ""
})

