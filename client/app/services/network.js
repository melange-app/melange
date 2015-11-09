var S = require("../store");

var connect = function() {
    window.addEventListener('online', function(event) {
        console.log("Navigator is now online.");
    });

    window.addEventListener('offline', function(event) {
        console.log("Navigator is now offline.");
    });
}

connect();
console.log("Navigator online:", navigator.onLine);

module.exports = {};
