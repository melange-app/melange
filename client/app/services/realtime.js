var S = require("../store");
var api = require("./api");

// The variable that holds the actual websocket.
var socket;

// A map that holds a list of subscribed functions for
// updates from the websocket.
var subscribers = {};

function sendData(type, data) {
    if (socket == undefined) {
        console.log("Attempted to send data to WebSocket" +
                    " before connection was opened.");
        return;
    }
    
    // Construct a msg object containing the data the server needs to
    // process the message from the chat client.
    var msg = {
        type: type,
        data: data,
    };

    // Send the msg object as a JSON-formatted string.
    socket.send(JSON.stringify(msg));
}

var buffer = Immutable.List();
var firstPassDone = false;
var startedMessages = false;

var onMessage = function(event) {
    var msg = JSON.parse(event.data);
    if (msg["type"] == "initDone") {
        console.log("Received", buffer.length, "messages, loading now.");
        
        firstPassDone = true;
        
        console.log(S);
        S.dispatch(S.actions.messages.mergeMessages, buffer.toJS());

        S.dispatch(S.actions.status.setStatus, {
            loading: false,
            text: "",
        });
        
        return;
    } else if (msg["type"] == "message") {
        if (!startedMessages) {
            console.log("Receving new messages.");
            startedMessages = true;
        }
        
        // Go ahead and parse the date.
        msg.data["_parsedDate"] = moment(msg.data.date);
        
        if (!firstPassDone) {
            buffer = buffer.push(msg.data);
        } else {
           S.dispatch(S.actions.messages.loadMessage, msg.data); 
        }
        
        return;
    } else if (msg["type"] == "waitingForSetup") {
        // This is the first message returned on the websocket.
        // Should be ignored.
        return;
    }

    if(msg["type"] in subscribers) {
        var receivers = subscribers[msg["type"]];
        var toRemove = [];
        
        for(var i in receivers) {
            try {
                var r = receivers[i].callback(msg["data"]);

                if (r === true) {
                    toRemove.push(i);
                }
            } catch (err) {
                console.log(
                    "Couldn't get to callback (" + i + ") for " +
                    msg["type"] + "."
                );
                console.log(err);
                toRemove.push(i);
            }
        }

        // Remove all the subscribers that we were unable to access.
        for (var i in toRemove) {
            delete receivers[toRemove[i]];
        }
        
        subscribers[msg["type"]] = receivers;
    } else {
        console.log("Got unknown message with type " + msg["type"])
    }
}

var onClose = function(event) {
    console.log("Websocket disconnected...");
    console.log(event);
}

var connect = function() {
    socket = new WebSocket(api.endpoints.ws);

    // After openeing the WebSocket, we tell the server that we are
    // ready to listen for new data.
    socket.onopen = function(event) {
        console.log("Opened realtime connection.");
        sendData("startup", "ok");

        S.dispatch(S.actions.status.setConnection, {
            color: "green",
            text: "Connected",
        });

        S.dispatch(S.actions.status.setStatus, {
            loading: true,
            text: "Syncing messages...",
        });
    }

    socket.onmessage = onMessage;
    socket.onclose = onClose;
    socket.onerror = onClose;
}

module.exports = {
    connect: connect,
    subscribe: function(callback) {
        subscribers.push({
            callback: callback,
        });
    },
}
