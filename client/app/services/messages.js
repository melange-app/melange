var api = require("./api");
var prefix = Immutable.List(["messages"]);
var S = require('../store');

var messages = {
    _get: function(alias, name) {
        return api.post(prefix.push("get"), {
            alias: alias,
            name: name,
        });
    },
    _sentMessages: function() {
        return api.post();
    },
    _getFromUser: function(alias) {
        return api.post(prefix.push("at"));
    },
    _publish: function(data) {
        return api.post(prefix.push("new"), data);
    },
    get: function(alias, name) {
        var id = alias + "/" + name;
        
        S.dispatch(S.actions.messages.loadMessage, {
            id: id,
            state: {
                loaded: false,
            },
        });

        messages._get(alias, name).then(function(data) {
            S.dispatch(S.actions.messages.loadMessage, {
                id: id,
                message: data,
                state: {
                    loaded: true,
                },
            })
        });
    }
};

module.exports = messages;
