petBtn = document.getElementById('pet-henry-btn')
sendChatBtn = document.getElementById('send-chat-btn')
chatContent = document.getElementById('chat-content')

const ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = () => {
	console.log("Connected to the server");
	ws.send("Hello, server!");
};

ws.onmessage = (event) => {
	console.log("Received from server:", event.data);
	
	const data = JSON.parse(event.data)	
	if (data.type == "counter") {
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
	ws.send("pet");
})

sendChatBtn.addEventListener("click", (e) => {

	chatMessage = {
		type:	"chat",
		sender: "nathan",
		content: chatContent.value,
	}

	ws.send(JSON.stringify(chatMessage))
	chatContent.value = ""
})
