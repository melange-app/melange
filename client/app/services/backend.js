var exports = {};

var registerModule = function(name, mod) {
    exports[name] = mod;
}

registerModule("api", require("./api"));
registerModule("realtime", require("./realtime"));
registerModule("network", require("./network"));
registerModule("plugins", require("./plugins"));
registerModule("identity", require("./identity"));
registerModule("messages", require("./messages"));

module.exports = exports;
