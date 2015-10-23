var api = require("./api");
var prefix = Immutable.List(["identity"]);

var identity = {
    create: function() {
        return api.post(prefix.push("new"));
    },
    current: function() {
        return api.get(prefix.push("current"));
    },
    setCurrent: function(id) {
        return api.post(prefix.push("current"), {
            fingerprint: id.Fingerprint,
        });
    },
    all: function() {
        return api.get(prefix);
    },
    remove: function(id) {
        return api.post(prefix.push("remove"));
    }
};

module.exports = identity;
