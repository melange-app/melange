var api = require("./api");
var prefix = Immutable.List(["plugins"]);

var plugins = {
    installed: function() {
        return api.get(prefix);
    },
    store: function() {
        return api.get(prefix.push("store"));
    },
    updates: function() {
        return api.get(prefix.push("updates"));
    },
    update: function(plugin) {
        return api.post(prefix.push("update"));
    },
    install: function(url) {
        return api.post(prefix.push("install"));
    },
    uninstall: function(plugin) {
        return api.post(prefix.push("uninstall"));
    }
};

module.exports = plugins;
