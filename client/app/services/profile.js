var api = require("./api");
var prefix = Immutable.List(["profile"]);

var profiles = {
    update: function(profile) {
        return api.post(prefix.push("update"), profile);
    },
    current: function() {
        return api.get(prefix.push("current"));
    }
}

module.exports = profiles;
