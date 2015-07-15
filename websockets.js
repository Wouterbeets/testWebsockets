var sock = null;
var wsuri = "ws://127.0.0.1:8080";

window.onload = function() {

	console.log("onload");

	sock = new WebSocket(wsuri);

	sock.onopen = function() {
		console.log("connected to " + wsuri);
		send("give me tasks")
	}

	sock.onclose = function(e) {
		console.log("connection closed (" + e.code + ")");
	}

	sock.onmessage = function(e) {
		document.getElementById("task0").value = e.data;
		console.log("message received: " + e.data);
	}
};

function send(msg) {
	sock.send(msg);
};

